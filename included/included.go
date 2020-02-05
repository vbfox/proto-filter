package included

import (
	"github.com/jhump/protoreflect/desc"
	"github.com/vbfox/proto-filter/configuration"
)

type inclusionType int

const (
	inclusionType_unknown inclusionType = iota
	inclusionType_included_implicit
	inclusionType_included_explicit
	inclusionType_excluded_explicit
)

func buildInclusions(descriptors []*desc.FileDescriptor, configuration *configuration.Configuration) map[string]inclusionType {
	inclusions := make(map[string]inclusionType)
	return inclusions
}

// BuildIncluded create a map of every field that can be encountered and
// if they are included or not
func BuildIncluded(descriptors []*desc.FileDescriptor, configuration *configuration.Configuration) map[string]bool {
	result := make(map[string]bool)

	inclusions := buildInclusions(descriptors, configuration)

	for path, inclusionType := range inclusions {
		switch inclusionType {
		case inclusionType_included_implicit:
		case inclusionType_included_explicit:
			result[path] = true
		default:
			result[path] = false
		}
	}

	return result
}
