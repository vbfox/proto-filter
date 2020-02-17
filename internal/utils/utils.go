package utils

import (
	"strings"

	"github.com/jhump/protoreflect/desc"
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

func reverseStringSlice(ss []string) {
	last := len(ss) - 1
	for i := 0; i < len(ss)/2; i++ {
		ss[i], ss[last-i] = ss[last-i], ss[i]
	}
}

func GetDescriptorPath(descriptor desc.Descriptor) []string {
	result := []string{}
	current := descriptor.GetParent()
	for {
		if current == nil {
			break
		}
		result = append(result, current.GetName())
		current = current.GetParent()
	}
	reverseStringSlice(result)
	return result
}

func GetDescriptorPathString(descriptor desc.Descriptor) string {
	return BuildPath(GetDescriptorPath(descriptor))
}
