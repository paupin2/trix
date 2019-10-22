package trix

import (
	"fmt"
	"strings"
)

// ParseKeys converts a slice of interfaces into a slice of strings; string
// items can also include more than one dot-separated element.
func ParseKeys(keys []interface{}) []string {
	spec := make([]string, 0, len(keys))
	for _, key := range keys {
		var strPart string
		switch key.(type) {
		case string:
			strPart = key.(string)
		default:
			strPart = fmt.Sprint(key)
		}

		for _, subkey := range strings.Split(strPart, ".") {
			spec = append(spec, subkey)
		}
	}
	return spec
}

// indexEsc returns the index of the first instance of substr in s that isn't preceded by escape, or -1 if substr is not present in s.
func indexEsc(s, substr, escape string) int {
	totalOffset := 0
	for {
		index := strings.Index(s, substr)
		if index == -1 {
			return index
		}
		if escape == "" || index < len(escape) || s[index-len(escape):index] != escape {
			return totalOffset + index
		}
		// this substring is escaped; try the next one
		offset := index + len(substr)
		s = s[offset:]
		totalOffset += offset
	}
}

// splitNEsc slices s into at most n substrings, separated by sep (not preceded
// by escape) and returns a slice of the substrings between those separators.
func splitNEsc(s, sep, escape string, n int) []string {
	escapedSep := escape + sep
	parts := []string{}
	unlimited := n == -1
	if n == 0 {
		return parts
	}

	for (unlimited || n > 1) && s != "" {
		sepIndex := indexEsc(s, sep, escape)
		if sepIndex < 0 {
			break
		}
		parts = append(parts, strings.Replace(s[:sepIndex], escapedSep, sep, -1))
		s = s[sepIndex+len(sep):]

		n--
	}
	parts = append(parts, strings.Replace(s, escapedSep, sep, -1))
	return parts
}

// splitEsc slices s into all substrings separated by sep (not preceded by escape)
// and returns a slice of the substrings between those separators.
func splitEsc(s, sep, escape string) []string {
	return splitNEsc(s, sep, escape, -1)
}
