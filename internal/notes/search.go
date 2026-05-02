package notes

import (
	"sort"
	"strconv"
	"strings"
	"unicode"
)

type Match struct {
	Entry  Entry
	Score  int
	Label  string
	Detail string
}

func Filter(entries []Entry, query string) []Match {
	query = normalize(query)
	if query == "" {
		matches := make([]Match, 0, len(entries))
		for _, entry := range entries {
			matches = append(matches, Match{
				Entry:  entry,
				Score:  1,
				Label:  entry.Title(),
				Detail: formatDetail(entry),
			})
		}
		return matches
	}

	terms := strings.Fields(query)
	matches := make([]Match, 0, len(entries))

	for _, entry := range entries {
		score, ok := scoreEntry(entry, terms)
		if !ok {
			continue
		}

		matches = append(matches, Match{
			Entry:  entry,
			Score:  score,
			Label:  entry.Title(),
			Detail: formatDetail(entry),
		})
	}

	sort.Slice(matches, func(i, j int) bool {
		if matches[i].Score == matches[j].Score {
			return matches[i].Label < matches[j].Label
		}
		return matches[i].Score > matches[j].Score
	})

	return matches
}

func scoreEntry(entry Entry, terms []string) (int, bool) {
	fields := weightedFields(entry)
	total := 0

	for _, term := range terms {
		best := 0
		for _, field := range fields {
			score := matchScore(field, term)
			if score > best {
				best = score
			}
		}

		if best == 0 {
			return 0, false
		}
		total += best
	}

	if entry.IsRun() {
		total += 4
	}

	return total, true
}

type weightedField struct {
	value  string
	weight int
}

func weightedFields(entry Entry) []weightedField {
	fields := []weightedField{
		{value: normalize(entry.Desc), weight: 8},
		{value: normalize(strings.TrimSuffix(entry.SourceFile, filepathExt(entry.SourceFile))), weight: 7},
		{value: normalize(strings.Join(entry.Tags, " ")), weight: 7},
		{value: normalize(entry.Run), weight: 5},
		{value: normalize(entry.Note), weight: 3},
		{value: normalize(entry.Banner), weight: 2},
	}

	out := make([]weightedField, 0, len(fields))
	for _, field := range fields {
		if field.value != "" {
			out = append(out, field)
		}
	}
	return out
}

func matchScore(field weightedField, term string) int {
	if isDigits(term) {
		return matchNumericTerm(field, term)
	}

	if strings.Contains(field.value, term) {
		return 120 + field.weight*10
	}

	normalizedTerm := normalizeNumericToken(term)

	bestWord := 0
	for _, word := range strings.Fields(field.value) {
		if normalizeNumericToken(word) == normalizedTerm {
			if len(normalizedTerm) >= 2 {
				return 118 + field.weight*10
			}
			return 112 + field.weight*10
		}
		if score := subsequenceScore(word, term); score > bestWord {
			bestWord = score
		}
	}
	if bestWord > 0 {
		return bestWord + field.weight*10
	}

	compactField := strings.ReplaceAll(field.value, " ", "")
	if len(term) <= 3 {
		return 0
	}
	if score := subsequenceScore(compactField, term); score > 0 {
		return score + field.weight*6
	}

	return 0
}

func matchNumericTerm(field weightedField, term string) int {
	normalizedTerm := normalizeNumericToken(term)
	best := 0

	for _, word := range strings.Fields(field.value) {
		if normalizeNumericToken(word) != normalizedTerm {
			continue
		}

		score := 118 + field.weight*10
		if isDigits(word) {
			score = 122 + field.weight*10
		}
		if score > best {
			best = score
		}
	}

	return best
}

func subsequenceScore(haystack, needle string) int {
	if needle == "" {
		return 1
	}

	hr := []rune(haystack)
	nr := []rune(needle)

	ni := 0
	first := -1
	last := -1

	for i, r := range hr {
		if ni < len(nr) && r == nr[ni] {
			if first == -1 {
				first = i
			}
			last = i
			ni++
			if ni == len(nr) {
				span := last - first + 1
				gaps := span - len(nr)
				if !isCloseSubsequence(span, len(nr), first) {
					return 0
				}
				return 100 - gaps*2
			}
		}
	}

	return 0
}

func isCloseSubsequence(span, needleLen, first int) bool {
	if needleLen <= 2 {
		return span == needleLen
	}

	maxExtra := 1
	if needleLen >= 5 {
		maxExtra = 2
	}
	if needleLen >= 8 {
		maxExtra = 3
	}

	if span > needleLen+maxExtra {
		return false
	}

	// Prefer matches near the start of a token to avoid noisy deep subsequences.
	if first > maxExtra+1 {
		return false
	}

	return true
}

func normalize(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	lastSpace := false

	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			lastSpace = false
			continue
		}

		if !lastSpace {
			b.WriteByte(' ')
			lastSpace = true
		}
	}

	return strings.TrimSpace(b.String())
}

func formatDetail(entry Entry) string {
	return entry.Action() + " | " + entry.DisplayValue()
}

func normalizeNumericToken(value string) string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return !(unicode.IsLetter(r) || unicode.IsDigit(r))
	})
	if len(parts) == 0 {
		return value
	}

	normalized := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			continue
		}
		if isDigits(part) {
			number, err := strconv.Atoi(part)
			if err == nil {
				normalized = append(normalized, strconv.Itoa(number))
				continue
			}
		}
		normalized = append(normalized, part)
	}

	return strings.Join(normalized, " ")
}

func isDigits(value string) bool {
	for _, r := range value {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return value != ""
}

func filepathExt(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '.' {
			return path[i:]
		}
		if path[i] == '/' {
			break
		}
	}
	return ""
}
