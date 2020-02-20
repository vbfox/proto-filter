package protofilter

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/internal/builderutil"
)

func (s *filteringState) Pass1() error {
	for _, descriptor := range s.descriptors {
		fileBuilder, err := s.Pass1File(descriptor)
		if err != nil {
			return fmt.Errorf("Failed to filter file %s: %w", descriptor.GetName(), err)
		}
		if fileBuilder != nil {
			s.fileBuilders[descriptor.GetFullyQualifiedName()] = fileBuilder
		}
	}

	return nil
}

func (s *filteringState) Pass1File(descriptor *desc.FileDescriptor) (*builder.FileBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	result := builder.NewFile(descriptor.GetName())

	builderutil.SetFileBasicInfo(result, descriptor)
	builderutil.SetAllComments(result, descriptor)

	for _, message := range descriptor.GetMessageTypes() {
		messageBuilder, err := s.Pass1Message(message)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
		if messageBuilder != nil {
			if err := result.TryAddMessage(messageBuilder); err != nil {
				return nil, err
			}
		}
	}

	for _, enum := range descriptor.GetEnumTypes() {
		enumBuilder, err := s.Pass1Enum(enum)
		if err != nil {
			return nil, fmt.Errorf("Error in enum %s: %w", enum.GetName(), err)
		}
		if enumBuilder != nil {
			if err := result.TryAddEnum(enumBuilder); err != nil {
				return nil, err
			}
		}
	}

	for _, service := range descriptor.GetServices() {
		serviceBuilder, err := s.Pass1Service(service)
		if err != nil {
			return nil, fmt.Errorf("Error in service %s: %w", service.GetName(), err)
		}
		if serviceBuilder != nil {
			if err := result.TryAddService(serviceBuilder); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (s *filteringState) Pass1Message(descriptor *desc.MessageDescriptor) (*builder.MessageBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	result := builder.NewMessage(descriptor.GetName())
	s.messageBuilders[descriptor.GetFullyQualifiedName()] = result

	for _, message := range descriptor.GetNestedMessageTypes() {
		messageBuilder, err := s.Pass1Message(message)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
		if messageBuilder != nil {
			if err := result.TryAddNestedMessage(messageBuilder); err != nil {
				return nil, err
			}
		}
	}

	for _, enum := range descriptor.GetNestedEnumTypes() {
		enumBuilder, err := s.Pass1Enum(enum)
		if err != nil {
			return nil, fmt.Errorf("Error in enum %s: %w", enum.GetName(), err)
		}
		if enumBuilder != nil {
			if err := result.TryAddNestedEnum(enumBuilder); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (s *filteringState) Pass1Enum(descriptor *desc.EnumDescriptor) (*builder.EnumBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	result := builder.NewEnum(descriptor.GetName())
	s.enumBuilders[descriptor.GetFullyQualifiedName()] = result

	for _, enumValue := range descriptor.GetValues() {
		enumValueBuilder, err := s.Pass1EnumValue(enumValue)
		if err != nil {
			return nil, fmt.Errorf("Error in message %s: %w", enumValue.GetName(), err)
		}
		if enumValueBuilder != nil {
			if err := result.TryAddValue(enumValueBuilder); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (s *filteringState) Pass1EnumValue(descriptor *desc.EnumValueDescriptor) (*builder.EnumValueBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	result := builder.NewEnumValue(descriptor.GetName())

	result.SetNumber(descriptor.GetNumber())

	return result, nil
}

func (s *filteringState) Pass1Service(descriptor *desc.ServiceDescriptor) (*builder.ServiceBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	result := builder.NewService(descriptor.GetName())
	s.serviceBuilders[descriptor.GetFullyQualifiedName()] = result

	return result, nil
}
