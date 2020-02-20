package protofilter

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/internal/builderutil"
)

func (s *filteringState) Pass2() error {
	for _, descriptor := range s.descriptors {
		err := s.Pass2File(descriptor)
		if err != nil {
			return fmt.Errorf("Failed to filter file %s: %w", descriptor.GetName(), err)
		}
	}

	return nil
}

func (s *filteringState) Pass2File(descriptor *desc.FileDescriptor) error {
	_, found := s.fileBuilders[descriptor.GetFullyQualifiedName()]
	if !found {
		return nil
	}

	for _, message := range descriptor.GetMessageTypes() {
		err := s.Pass2Message(message)
		if err != nil {
			return fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
	}

	for _, service := range descriptor.GetServices() {
		err := s.Pass2Service(service)
		if err != nil {
			return fmt.Errorf("Error in enum %s: %w", service.GetName(), err)
		}
	}

	return nil
}

func (s *filteringState) Pass2Message(descriptor *desc.MessageDescriptor) error {
	result, found := s.messageBuilders[descriptor.GetFullyQualifiedName()]
	if !found {
		return nil
	}

	for _, message := range descriptor.GetNestedMessageTypes() {
		err := s.Pass2Message(message)
		if err != nil {
			return fmt.Errorf("Error in message %s: %w", message.GetName(), err)
		}
	}

	for _, field := range descriptor.GetFields() {
		fieldBuilder, err := s.Pass2Field(field)
		if err != nil {
			return fmt.Errorf("Error in field %s: %w", field.GetName(), err)
		}
		if fieldBuilder != nil {
			if err := result.TryAddField(fieldBuilder); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *filteringState) Pass2Field(descriptor *desc.FieldDescriptor) (*builder.FieldBuilder, error) {
	if !s.IsIncluded(descriptor) {
		return nil, nil
	}

	var fieldType *builder.FieldType
	messageType := descriptor.GetMessageType()
	enumType := descriptor.GetEnumType()
	if messageType != nil {
		messageTypeBuilder := s.messageBuilders[messageType.GetFullyQualifiedName()]
		fieldType = builder.FieldTypeMessage(messageTypeBuilder)
	} else if enumType != nil {
		enumTypeBuilder := s.enumBuilders[enumType.GetFullyQualifiedName()]
		fieldType = builder.FieldTypeEnum(enumTypeBuilder)
	} else {
		fieldType = builder.FieldTypeScalar(descriptor.GetType())
	}

	result := builder.NewField(descriptor.GetName(), fieldType)
	result.SetLabel(descriptor.GetLabel())
	result.SetNumber(descriptor.GetNumber())
	builderutil.SetComments(result.GetComments(), descriptor.GetSourceInfo())

	return result, nil
}

func (s *filteringState) Pass2Service(descriptor *desc.ServiceDescriptor) error {
	result, found := s.serviceBuilders[descriptor.GetFullyQualifiedName()]
	if !found {
		return nil
	}

	for _, method := range descriptor.GetMethods() {
		methodBuilder := s.Pass2Method(method)
		if methodBuilder != nil {
			if err := result.TryAddMethod(methodBuilder); err != nil {
				return err
			}
		}
	}

	return nil
}

func (s *filteringState) Pass2Method(descriptor *desc.MethodDescriptor) *builder.MethodBuilder {
	if !s.IsIncluded(descriptor) {
		return nil
	}

	reqBuilder := s.messageBuilders[descriptor.GetInputType().GetFullyQualifiedName()]
	req := builder.RpcTypeMessage(reqBuilder, descriptor.IsClientStreaming())

	respBuilder := s.messageBuilders[descriptor.GetOutputType().GetFullyQualifiedName()]
	resp := builder.RpcTypeMessage(respBuilder, descriptor.IsServerStreaming())

	result := builder.NewMethod(descriptor.GetName(), req, resp)
	builderutil.SetComments(result.GetComments(), descriptor.GetSourceInfo())

	return result
}
