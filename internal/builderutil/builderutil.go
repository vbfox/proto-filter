package builderutil

import (
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/builder"
)

const (
	// File_packageTag is the tag number of the package element in a file
	// descriptor proto.
	filePackageTag = 2

	// File_syntaxTag is the tag number of the syntax element in a file
	// descriptor proto.
	fileSyntaxTag = 12
)

func SetComments(c *builder.Comments, loc *dpb.SourceCodeInfo_Location) {
	c.LeadingDetachedComments = loc.GetLeadingDetachedComments()
	c.LeadingComment = loc.GetLeadingComments()
	c.TrailingComment = loc.GetTrailingComments()
}

func SetAllComments(fileBuilder *builder.FileBuilder, descriptor *desc.FileDescriptor) {
	SetComments(fileBuilder.GetComments(), descriptor.GetSourceInfo())

	// find syntax and package comments, too
	for _, loc := range descriptor.AsFileDescriptorProto().GetSourceCodeInfo().GetLocation() {
		if len(loc.Path) == 1 {
			if loc.Path[0] == fileSyntaxTag {
				SetComments(&fileBuilder.SyntaxComments, loc)
			} else if loc.Path[0] == filePackageTag {
				SetComments(&fileBuilder.PackageComments, loc)
			}
		}
	}
}

func SetFileBasicInfo(fileBuilder *builder.FileBuilder, descriptor *desc.FileDescriptor) {
	fileBuilder.IsProto3 = descriptor.IsProto3()
	fileBuilder.Package = descriptor.GetPackage()
	fileBuilder.Options = descriptor.GetFileOptions()
}
