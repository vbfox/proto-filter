package protofilter

import (
	"github.com/jhump/protoreflect/desc"
	_ "github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/configuration"
)

func FilterSet(set []*desc.FileDescriptor, config *configuration.Configuration) ([]*desc.FileDescriptor, error) {
	return set, nil
}
