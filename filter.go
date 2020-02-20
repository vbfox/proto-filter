package protofilter

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/configuration"
	"github.com/vbfox/proto-filter/internal/included"
)

type filteringState struct {
	descriptors []*desc.FileDescriptor
	config      *configuration.Configuration
	builders    []*builder.FileBuilder
	included    map[string]bool
}

func (s *filteringState) IsIncluded(path string) bool {
	value, found := s.included[path]
	return found && value
}

func (s *filteringState) RunFilter() error {
	err := s.Pass1()
	if err != nil {
		return err
	}

	return nil
}

/*
func (s *filteringState) AddField(mb *builder.MessageBuilder, field *desc.FieldDescriptor) error {
	var fieldType *builder.FieldType

	messageType := field.GetMessageType()
	if messageType != nil {

		builder.FieldTypeMessage()
	} else {
		builder.FieldTypeScalar(field.GetType())
	}

	fb := builder.NewField(field.GetName(), field.GetType())
	mb.AddField(fb)
	return nil
}

func (s *filteringState) AddMessage(fb *builder.FileBuilder, message *desc.MessageDescriptor) error {
	mb := builder.NewMessage(message.GetName())

	for _, field := range message.GetFields() {
		err := s.AddField(mb, field)
		if err != nil {
			return fmt.Errorf("Error in field %s: %w", field.GetName(), err)
		}
	}

	return fb.TryAddMessage(mb)
}

func (s *filteringState) AddFileDescriptor(descriptor *desc.FileDescriptor) error {
	fb := builder.NewFile(descriptor.GetName())
	s.builders = append(s.builders, fb)

	builderutil.SetFileBasicInfo(fb, descriptor)
	builderutil.SetAllComments(fb, descriptor)

	for _, message := range descriptor.GetMessageTypes() {
		err := s.AddMessage(fb, message)
		if err != nil {
			return fmt.Errorf("Error while filtering message %s: %w", message.GetName(), err)
		}
	}

	return nil
}
*/

func initState(descriptors []*desc.FileDescriptor, config *configuration.Configuration) (*filteringState, error) {
	included, err := included.BuildIncluded(descriptors, config)
	if err != nil {
		return nil, err
	}
	return &filteringState{
		descriptors: descriptors,
		config:      config,
		builders:    []*builder.FileBuilder{},
		included:    included,
	}, nil
}

func (s *filteringState) GetDescriptors() ([]*desc.FileDescriptor, error) {
	result := []*desc.FileDescriptor{}

	for _, builder := range s.builders {
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
