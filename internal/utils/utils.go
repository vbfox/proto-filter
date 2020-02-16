package utils

import (
	"strings"
)

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
