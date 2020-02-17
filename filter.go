package protofilter

import (
	"fmt"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/configuration"
	"github.com/vbfox/proto-filter/internal/included"
	"github.com/vbfox/proto-filter/internal/utils"
)

type fileState struct {
	builder    *builder.FileBuilder
	descriptor *desc.FileDescriptor
}

type messageState struct {
	builder    *builder.MessageBuilder
	descriptor *desc.MessageDescriptor
}

type enumState struct {
	builder    *builder.EnumBuilder
	descriptor *desc.EnumDescriptor
}

type serviceState struct {
	builder    *builder.ServiceBuilder
	descriptor *desc.ServiceDescriptor
}

type filteringState struct {
	descriptors []*desc.FileDescriptor
	config      *configuration.Configuration
	included    map[string]bool

	fileBuilders    map[string]fileState
	messageBuilders map[string]messageState
	enumBuilders    map[string]enumState
	serviceBuilders map[string]serviceState
}

const (
	// File_packageTag is the tag number of the package element in a file
	// descriptor proto.
	filePackageTag = 2

	// File_syntaxTag is the tag number of the syntax element in a file
	// descriptor proto.
	fileSyntaxTag = 12
)

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

func setComments(c *builder.Comments, loc *dpb.SourceCodeInfo_Location) {
	c.LeadingDetachedComments = loc.GetLeadingDetachedComments()
	c.LeadingComment = loc.GetLeadingComments()
	c.TrailingComment = loc.GetTrailingComments()
}

func setAllComments(fileBuilder *builder.FileBuilder, descriptor *desc.FileDescriptor) {
	setComments(fileBuilder.GetComments(), descriptor.GetSourceInfo())

	// find syntax and package comments, too
	for _, loc := range descriptor.AsFileDescriptorProto().GetSourceCodeInfo().GetLocation() {
		if len(loc.Path) == 1 {
			if loc.Path[0] == fileSyntaxTag {
				setComments(&fileBuilder.SyntaxComments, loc)
			} else if loc.Path[0] == filePackageTag {
				setComments(&fileBuilder.PackageComments, loc)
			}
		}
	}
}

func setFileBasicInfo(fileBuilder *builder.FileBuilder, descriptor *desc.FileDescriptor) {
	fileBuilder.IsProto3 = descriptor.IsProto3()
	fileBuilder.Package = descriptor.GetPackage()
	fileBuilder.Options = descriptor.GetFileOptions()
}

func (s *filteringState) AddFileDescriptor(descriptor *desc.FileDescriptor) error {
    path := utils.GetDescriptorPathString(descriptor)
	fb := builder.NewFile(descriptor.GetName())
	s.fileBuilders[] = fileState{descriptor: descriptor, builder: fb}

	setFileBasicInfo(fb, descriptor)
	setAllComments(fb, descriptor)

	for _, message := range descriptor.GetMessageTypes() {
		err := s.AddMessage(fb, message)
		if err != nil {
			return fmt.Errorf("Error while filtering message %s: %w", message.GetName(), err)
		}
	}

	return nil
}

func initState(descriptors []*desc.FileDescriptor, config *configuration.Configuration) (*filteringState, error) {
	included, err := included.BuildIncluded(descriptors, config)
	if err != nil {
		return nil, err
	}
	return &filteringState{
		descriptors:     descriptors,
		config:          config,
		included:        included,
		fileBuilders:    map[string]fileState{},
		messageBuilders: map[string]messageState{},
		enumBuilders:    map[string]enumState{},
		serviceBuilders: map[string]serviceState{},
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

func (s *filteringState) RunFilter() error {
	for _, descriptor := range s.descriptors {
		err := s.AddFileDescriptor(descriptor)
		if err != nil {
			return fmt.Errorf("Failed to filter file %s: %w", descriptor.GetName(), err)
		}
	}

	return nil
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
