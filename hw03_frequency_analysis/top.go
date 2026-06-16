package hw03frequencyanalysis

import (
	"regexp"
	"sort"
	"strings"
)

type pair struct {
	word  string
	count int
}

var rePunct = regexp.MustCompile(`[^\p{L}\p{N}-]`)

func Top10(text string) []string {
	if text == "" {
		return nil
	}

	text = strings.ToLower(text)
	words := strings.Fields(text)
	mpFreq := make(map[string]int)

	for _, w := range words {
		clean := strings.TrimFunc(w, func(r rune) bool {
			return rePunct.MatchString(string(r)) && r != '-'
		})

		if clean == "" || clean == "-" || strings.Trim(clean, "-") == "" {
			continue
		}

		mpFreq[clean]++
	}

	pairs := make([]pair, 0, len(mpFreq))
	for w, c := range mpFreq {
		pairs = append(pairs, pair{w, c})
	}

	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count != pairs[j].count {
			return pairs[i].count > pairs[j].count
		}
		return pairs[i].word < pairs[j].word
	})

	result := make([]string, 0, 10)
	for i := 0; i < len(pairs) && i < 10; i++ {
		result = append(result, pairs[i].word)
	}

	return result
}
