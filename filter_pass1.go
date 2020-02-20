package protofilter

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/internal/builderutil"
	"github.com/vbfox/proto-filter/internal/utils"
)

func (s *filteringState) Pass1() error {
	for _, descriptor := range s.descriptors {
		fileBuilder, err := s.Pass1File(descriptor)
		if err != nil {
			return fmt.Errorf("Failed to filter file %s: %w", descriptor.GetName(), err)
		}
		if fileBuilder != nil {
			s.builders = append(s.builders, fileBuilder)
		}
	}

	return nil
}

func (s *filteringState) Pass1File(descriptor *desc.FileDescriptor) (*builder.FileBuilder, error) {
	path := descriptor.GetName()
	if !s.IsIncluded(path) {
		return nil, nil
	}

	result := builder.NewFile(path)

	builderutil.SetFileBasicInfo(result, descriptor)
	builderutil.SetAllComments(result, descriptor)

	for _, message := range descriptor.GetMessageTypes() {
		messageBuilder, err := s.Pass1Message(*message, path)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
		if messageBuilder != nil {
			result.AddMessage(messageBuilder)
		}
	}

	for _, enum := range descriptor.GetEnumTypes() {
		enumBuilder, err := s.Pass1Enum(*enum, path)
		if err != nil {
			return nil, fmt.Errorf("Error in enum %s: %w", enum.GetName(), err)
		}
		if enumBuilder != nil {
			result.AddEnum(enumBuilder)
		}
	}

	for _, service := range descriptor.GetServices() {
		enumBuilder, err := s.Pass1Service(*service, path)
		if err != nil {
			return nil, fmt.Errorf("Error in enum %s: %w", service.GetName(), err)
		}
		if enumBuilder != nil {
			result.AddService(enumBuilder)
		}
	}

	return result, nil
}

func (s *filteringState) Pass1Message(descriptor desc.MessageDescriptor, parentPath string) (*builder.MessageBuilder, error) {
	path := utils.PathConcat(parentPath, descriptor.GetName())
	if !s.IsIncluded(path) {
		return nil, nil
	}

	result := builder.NewMessage(descriptor.GetName())

	for _, message := range descriptor.GetNestedMessageTypes() {
		messageBuilder, err := s.Pass1Message(*message, path)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
		if messageBuilder != nil {
			result.AddNestedMessage(messageBuilder)
		}
	}

	for _, enum := range descriptor.GetNestedEnumTypes() {
		enumBuilder, err := s.Pass1Enum(*enum, path)
		if err != nil {
			return nil, fmt.Errorf("Error in enum %s: %w", enum.GetName(), err)
		}
		if enumBuilder != nil {
			result.AddNestedEnum(enumBuilder)
		}
	}

	return result, nil
}

func (s *filteringState) Pass1Enum(descriptor desc.EnumDescriptor, parentPath string) (*builder.EnumBuilder, error) {
	path := utils.PathConcat(parentPath, descriptor.GetName())
	if !s.IsIncluded(path) {
		return nil, nil
	}

	result := builder.NewEnum(descriptor.GetName())

	for _, enumValue := range descriptor.GetValues() {
		enumValueBuilder, err := s.Pass1EnumValue(*enumValue, path)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", enumValue.GetName(), err)
		}
		if enumValueBuilder != nil {
			result.AddValue(enumValueBuilder)
		}
	}

	return result, nil
}

func (s *filteringState) Pass1EnumValue(descriptor desc.EnumValueDescriptor, parentPath string) (*builder.EnumValueBuilder, error) {
	path := utils.PathConcat(parentPath, descriptor.GetName())
	if !s.IsIncluded(path) {
		return nil, nil
	}

	result := builder.NewEnumValue(descriptor.GetName())

	result.SetNumber(descriptor.GetNumber())

	return result, nil
}

func (s *filteringState) Pass1Service(descriptor desc.ServiceDescriptor, parentPath string) (*builder.ServiceBuilder, error) {
	path := utils.PathConcat(parentPath, descriptor.GetName())
	if !s.IsIncluded(path) {
		return nil, nil
	}

	result := builder.NewService(descriptor.GetName())

	return result, nil
}
