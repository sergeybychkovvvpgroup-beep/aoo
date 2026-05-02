package notes

import (
	"sort"
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
	if strings.Contains(field.value, term) {
		return 120 + field.weight*10
	}

	bestWord := 0
	for _, word := range strings.Fields(field.value) {
		if score := subsequenceScore(word, term); score > bestWord {
			bestWord = score
		}
	}
	if bestWord > 0 {
		return bestWord + field.weight*10
	}

	compactField := strings.ReplaceAll(field.value, " ", "")
	if score := subsequenceScore(compactField, term); score > 0 {
		return score + field.weight*6
	}

	return 0
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
				return 100 - gaps*2
			}
		}
	}

	return 0
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

