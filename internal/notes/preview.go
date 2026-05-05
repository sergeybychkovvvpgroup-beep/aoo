package notes

import (
	"strings"
	"unicode"
)

const maxOccurrencesPerTerm = 4

type lineSnippet struct {
	start int
	end   int
	hits  int
	line  int
}

type Occurrence struct {
	Start int
	End   int
}

type PreviewMatch struct {
	Section     string
	Text        string
	Occurrences []Occurrence
	Snippets    []PreviewSnippet
}

type PreviewSnippet struct {
	Section    string
	Text       string
	Occurrence Occurrence
}

type previewCandidate struct {
	section string
	text    string
	score   int
}

type previewRank struct {
	matchedTerms    int
	exactTerms      int
	occurrenceCount int
	sectionScore    int
}

func buildPreview(entry Entry, terms []string) PreviewMatch {
	candidates := previewCandidates(entry)
	best := PreviewMatch{}
	var bestRank previewRank
	hasBest := false

	for _, candidate := range candidates {
		if strings.TrimSpace(candidate.text) == "" {
			continue
		}
		rank := rankPreviewCandidate(candidate, terms)
		if rank.matchedTerms == 0 {
			continue
		}
		occurrences := collectOccurrences(candidate.text, terms, 0)
		rank.occurrenceCount = len(occurrences)
		if !hasBest || betterPreviewRank(rank, bestRank) {
			hasBest = true
			bestRank = rank
			best = PreviewMatch{
				Section:     candidate.section,
				Text:        candidate.text,
				Occurrences: occurrences,
			}
		}
	}

	if hasBest {
		best.Snippets = buildSnippets(best.Section, best.Text, terms)
		if len(best.Snippets) > 0 {
			best.Occurrences = make([]Occurrence, 0, len(best.Snippets))
			for _, snippet := range best.Snippets {
				best.Occurrences = append(best.Occurrences, snippet.Occurrence)
			}
		}
		return best
	}
	return fallbackPreview(entry)
}

func BuildPreview(entry Entry, query string) PreviewMatch {
	terms := queryTerms(query)
	if len(terms) == 0 {
		return fallbackPreview(entry)
	}
	return buildPreview(entry, terms)
}

func fallbackPreview(entry Entry) PreviewMatch {
	for _, candidate := range previewCandidates(entry) {
		if strings.TrimSpace(candidate.text) == "" {
			continue
		}
		return PreviewMatch{
			Section: candidate.section,
			Text:    candidate.text,
			Snippets: []PreviewSnippet{{
				Section: candidate.section,
				Text:    candidate.text,
			}},
		}
	}

	return PreviewMatch{
		Section: "entry",
		Text:    strings.TrimSpace(entry.DisplayName()),
		Snippets: []PreviewSnippet{{
			Section: "entry",
			Text:    strings.TrimSpace(entry.DisplayName()),
		}},
	}
}

func previewCandidates(entry Entry) []previewCandidate {
	candidates := make([]previewCandidate, 0, 4+len(entry.ActionsList())*3)
	candidates = append(candidates,
		previewCandidate{section: "title", text: entry.DisplayName(), score: 10},
		previewCandidate{section: "source", text: entry.SourceFile, score: 9},
	)
	if entry.IsGroup() && strings.TrimSpace(entry.GroupSummary) != "" {
		candidates = append(candidates, previewCandidate{
			section: "group",
			text:    entry.GroupSummary,
			score:   12,
		})
	}
	if len(entry.Tags) > 0 {
		candidates = append(candidates, previewCandidate{
			section: "tags",
			text:    strings.Join(entry.Tags, " "),
			score:   8,
		})
	}

	for _, action := range entry.ActionsList() {
		if text := strings.TrimSpace(action.Text); text != "" {
			candidates = append(candidates, previewCandidate{
				section: "show",
				text:    text,
				score:   40,
			})
		}
		if cmd := strings.TrimSpace(action.Cmd); cmd != "" {
			candidates = append(candidates, previewCandidate{
				section: "command",
				text:    cmd,
				score:   30,
			})
		}
		if tmpl := strings.TrimSpace(action.Template); tmpl != "" {
			candidates = append(candidates, previewCandidate{
				section: "template",
				text:    tmpl,
				score:   28,
			})
		}
	}

	return candidates
}

func rankPreviewCandidate(candidate previewCandidate, terms []string) previewRank {
	rank := previewRank{sectionScore: candidate.score}
	field := weightedField{value: normalize(candidate.text), weight: 1}

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		if matchScore(field, term) > 0 {
			rank.matchedTerms++
		}
		if containsNormalizedTerm(candidate.text, term) {
			rank.exactTerms++
		}
	}

	return rank
}

func containsNormalizedTerm(text, term string) bool {
	normalizedText := normalize(text)
	normalizedTerm := normalizeNumericToken(strings.TrimSpace(term))
	if normalizedText == "" || normalizedTerm == "" {
		return false
	}
	return strings.Contains(normalizedText, normalizedTerm)
}

func betterPreviewRank(left, right previewRank) bool {
	if left.matchedTerms != right.matchedTerms {
		return left.matchedTerms > right.matchedTerms
	}
	if left.exactTerms != right.exactTerms {
		return left.exactTerms > right.exactTerms
	}
	if left.sectionScore != right.sectionScore {
		return left.sectionScore > right.sectionScore
	}
	return left.occurrenceCount > right.occurrenceCount
}

func buildSnippets(section, text string, terms []string) []PreviewSnippet {
	lines := splitPreviewTextLines(text)
	candidates := make([]lineSnippet, 0, len(lines))

	for _, line := range lines {
		limit := maxOccurrencesPerTerm
		if len(terms) > 1 {
			limit = maxOccurrencesPerTerm * len(terms)
		}
		occurrences := collectOccurrences(line.text, terms, limit)
		if len(occurrences) == 0 {
			continue
		}
		rank := rankPreviewCandidate(previewCandidate{
			section: section,
			text:    line.text,
			score:   1,
		}, terms)
		if rank.matchedTerms == 0 {
			continue
		}

		start := line.offset + occurrences[0].Start
		end := line.offset + occurrences[0].End
		for _, occurrence := range occurrences[1:] {
			if line.offset+occurrence.Start < start {
				start = line.offset + occurrence.Start
			}
			if line.offset+occurrence.End > end {
				end = line.offset + occurrence.End
			}
		}

		candidates = append(candidates, lineSnippet{
			start: start,
			end:   end,
			hits:  rank.matchedTerms*10 + rank.exactTerms*3 + len(occurrences),
			line:  line.index,
		})
	}

	if len(candidates) == 0 {
		return nil
	}
	sortLineSnippets(candidates)
	snippets := make([]PreviewSnippet, 0, len(candidates))
	for _, item := range candidates {
		snippets = append(snippets, PreviewSnippet{
			Section: section,
			Text:    text,
			Occurrence: Occurrence{
				Start: item.start,
				End:   item.end,
			},
		})
	}
	return snippets
}

type previewTextLine struct {
	index  int
	offset int
	text   string
}

func splitPreviewTextLines(text string) []previewTextLine {
	rawLines := strings.Split(text, "\n")
	lines := make([]previewTextLine, 0, len(rawLines))
	offset := 0
	for i, line := range rawLines {
		lines = append(lines, previewTextLine{
			index:  i,
			offset: offset,
			text:   strings.TrimRight(line, " "),
		})
		offset += len([]rune(line)) + 1
	}
	return lines
}

func previewSnippetLine(snippet PreviewSnippet) string {
	runes := []rune(snippet.Text)
	if len(runes) == 0 {
		return ""
	}
	start := snippet.Occurrence.Start
	if start < 0 {
		start = 0
	}
	if start > len(runes) {
		start = len(runes)
	}
	end := snippet.Occurrence.End
	if end < start {
		end = start
	}
	if end > len(runes) {
		end = len(runes)
	}
	lineStart := start
	for lineStart > 0 && runes[lineStart-1] != '\n' {
		lineStart--
	}
	lineEnd := end
	for lineEnd < len(runes) && runes[lineEnd] != '\n' {
		lineEnd++
	}
	return strings.TrimSpace(string(runes[lineStart:lineEnd]))
}

func sortLineSnippets(lines []lineSnippet) {
	for i := 0; i < len(lines); i++ {
		for j := i + 1; j < len(lines); j++ {
			if lines[j].hits > lines[i].hits ||
				(lines[j].hits == lines[i].hits && lines[j].line < lines[i].line) {
				lines[i], lines[j] = lines[j], lines[i]
			}
		}
	}
}

func collectOccurrences(text string, terms []string, limit int) []Occurrence {
	index := buildNormalizedIndex(text)
	if index.normalized == "" {
		return nil
	}
	normalizedRunes := []rune(index.normalized)

	seen := make(map[Occurrence]struct{})
	out := make([]Occurrence, 0, len(terms))

	for _, term := range terms {
		term = strings.TrimSpace(term)
		if term == "" {
			continue
		}
		normalizedTerm := normalizeNumericToken(term)
		if normalizedTerm == "" {
			continue
		}
		termRunes := []rune(normalizedTerm)
		if len(termRunes) == 0 {
			continue
		}

		offset := 0
		collectedForTerm := 0
		for {
			position := indexRunes(normalizedRunes[offset:], termRunes)
			if position == -1 {
				break
			}
			start := offset + position
			end := start + len(termRunes)
			if start < 0 || end <= 0 || end > len(index.rawOffsets) || start >= len(index.rawOffsets) {
				break
			}
			occurrence := Occurrence{
				Start: index.rawOffsets[start],
				End:   index.rawOffsets[end-1] + 1,
			}
			if _, ok := seen[occurrence]; !ok {
				seen[occurrence] = struct{}{}
				out = append(out, occurrence)
				collectedForTerm++
			}
			offset = start + 1
			if offset >= len(normalizedRunes) {
				break
			}
			if collectedForTerm >= maxOccurrencesPerTerm {
				break
			}
		}
	}

	sortOccurrences(out)
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

type normalizedIndex struct {
	normalized string
	rawOffsets []int
}

func buildNormalizedIndex(text string) normalizedIndex {
	runes := []rune(text)
	var normalized []rune
	offsets := make([]int, 0, len(runes))
	lastSpace := false

	for i, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if unicode.IsUpper(r) {
				r = unicode.ToLower(r)
			}
			normalized = append(normalized, r)
			offsets = append(offsets, i)
			lastSpace = false
			continue
		}

		if lastSpace {
			continue
		}
		normalized = append(normalized, ' ')
		offsets = append(offsets, i)
		lastSpace = true
	}

	return normalizedIndex{
		normalized: strings.TrimSpace(string(normalized)),
		rawOffsets: trimOffsets(normalized, offsets),
	}
}

func indexRunes(haystack, needle []rune) int {
	if len(needle) == 0 {
		return 0
	}
	if len(needle) > len(haystack) {
		return -1
	}

limit:
	for i := 0; i <= len(haystack)-len(needle); i++ {
		for j := 0; j < len(needle); j++ {
			if haystack[i+j] != needle[j] {
				continue limit
			}
		}
		return i
	}

	return -1
}

func trimOffsets(normalized []rune, offsets []int) []int {
	start := 0
	end := len(normalized)
	for start < end && normalized[start] == ' ' {
		start++
	}
	for end > start && normalized[end-1] == ' ' {
		end--
	}
	if start == 0 && end == len(normalized) {
		return offsets
	}
	return append([]int{}, offsets[start:end]...)
}

func sortOccurrences(occurrences []Occurrence) {
	for i := 0; i < len(occurrences); i++ {
		for j := i + 1; j < len(occurrences); j++ {
			if occurrences[j].Start < occurrences[i].Start {
				occurrences[i], occurrences[j] = occurrences[j], occurrences[i]
			}
		}
	}
}
