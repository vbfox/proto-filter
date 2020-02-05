package protofilter

import (
	"fmt"

	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
	"github.com/vbfox/proto-filter/configuration"
	"github.com/vbfox/proto-filter/included"
)

type filteringState struct {
	descriptors []*desc.FileDescriptor
	config      *configuration.Configuration
	builders    []*builder.FileBuilder
	included    map[string]bool
}

const (
	// File_packageTag is the tag number of the package element in a file
	// descriptor proto.
	file_packageTag = 2

	// File_syntaxTag is the tag number of the syntax element in a file
	// descriptor proto.
	file_syntaxTag = 12
)

func (s *filteringState) RunFilter() error {
	for _, descriptor := range s.descriptors {
		err := s.AddFileDescriptor(descriptor)
		if err != nil {
			return fmt.Errorf("Failed to filter file %s: %w", descriptor.GetName(), err)
		}
	}

	return nil
}

func (s *filteringState) AddField(mb *builder.MessageBuilder, field *desc.FieldDescriptor) error {
	//builder.FieldTypeScalar(field.GetType())
	//fb := builder.NewField(field.GetName(), field.GetType())
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
			if loc.Path[0] == file_syntaxTag {
				setComments(&fileBuilder.SyntaxComments, loc)
			} else if loc.Path[0] == file_packageTag {
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
	fb := builder.NewFile(descriptor.GetName())
	s.builders = append(s.builders, fb)

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

func initState(descriptors []*desc.FileDescriptor, config *configuration.Configuration) *filteringState {
	return &filteringState{
		descriptors: descriptors,
		config:      config,
		builders:    []*builder.FileBuilder{},
		included:    included.BuildIncluded(descriptors, config),
	}
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
	state := initState(descriptors, config)

	err := state.RunFilter()
	if err != nil {
		return nil, err
	}

	return state.GetDescriptors()
}
