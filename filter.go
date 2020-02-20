package protofilter

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/configuration"
	"github.com/vbfox/proto-filter/internal/included"
)

type filteringState struct {
	descriptors     []*desc.FileDescriptor
	config          *configuration.Configuration
	fileBuilders    map[string]*builder.FileBuilder
	messageBuilders map[string]*builder.MessageBuilder
	enumBuilders    map[string]*builder.EnumBuilder
	serviceBuilders map[string]*builder.ServiceBuilder
	included        map[string]bool
}

func (s *filteringState) IsIncluded(descriptor desc.Descriptor) bool {
	value, found := s.included[descriptor.GetFullyQualifiedName()]
	return found && value
}

func (s *filteringState) RunFilter() error {
	err := s.Pass1()
	if err != nil {
		return err
	}

	return s.Pass2()
}

func initState(descriptors []*desc.FileDescriptor, config *configuration.Configuration) (*filteringState, error) {
	included, err := included.BuildIncluded(descriptors, config)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Included = %+v", included)

	return &filteringState{
		descriptors:     descriptors,
		config:          config,
		fileBuilders:    map[string]*builder.FileBuilder{},
		messageBuilders: map[string]*builder.MessageBuilder{},
		enumBuilders:    map[string]*builder.EnumBuilder{},
		serviceBuilders: map[string]*builder.ServiceBuilder{},
		included:        included,
	}, nil
}

func (s *filteringState) GetDescriptors() ([]*desc.FileDescriptor, error) {
	result := []*desc.FileDescriptor{}

	for _, builder := range s.fileBuilders {
		descriptor, err := builder.Build()
		if err != nil {
			return nil, fmt.Errorf("Failed to build descriptor for %v: %w", builder.GetName(), err)
		}
		result = append(result, descriptor)
	}

	return result, nil
}

func FilterSet(descriptors []*desc.FileDescriptor, config *configuration.Configuration) ([]*desc.FileDescriptor, error) {
	state, err := initState(descriptors, config)
	if err != nil {
		return nil, err
	}

	err = state.RunFilter()
	if err != nil {
		return nil, err
	}

	return state.GetDescriptors()
}
