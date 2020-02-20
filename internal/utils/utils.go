package utils

import (
	"strings"
)

func PathConcat(pathA string, pathB string) string {
	var result strings.Builder
	result.WriteString(pathA)
	if len(pathA) > 0 && len(pathB) > 0 {
		result.WriteString("/")
	}
	result.WriteString(pathB)

	return result.String()
}

func BuildPath(parts []string) string {
	var result strings.Builder
	empty := true
	for _, part := range parts {
		if empty {
			result.WriteString(part)
			empty = false
		} else {
			result.WriteString("/")
			result.WriteString(part)
		}
	}
	return result.String()
}
