package envmodel

import (
	"regexp"
	"strings"
)

var (
	gatherRegexp  = regexp.MustCompile("([^A-Z]+|[A-Z]+[^A-Z]+|[A-Z]+)")
	acronymRegexp = regexp.MustCompile("([A-Z]+)([A-Z][^A-Z]+)")
)

func pascal2snake(pascal string) string {
	words := gatherRegexp.FindAllStringSubmatch(pascal, -1)
	parts := make([]string, 0, len(words))
	for _, words := range words {
		if m := acronymRegexp.FindStringSubmatch(words[0]); len(m) == 3 {
			parts = append(parts, m[1], m[2])
		} else {
			parts = append(parts, words[0])
		}
	}

	return strings.ToUpper(strings.Join(parts, "_"))
}
