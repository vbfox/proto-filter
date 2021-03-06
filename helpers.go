package protofilter

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	_ "github.com/jhump/protoreflect/desc/builder"
	"github.com/jhump/protoreflect/desc/protoprint"
)

func loadFileDescriptorSet(path string) (*dpb.FileDescriptorSet, error) {
	var fds dpb.FileDescriptorSet
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if err = proto.Unmarshal(bb, &fds); err != nil {
		return nil, err
	}

	return &fds, nil
}

func LoadProtoSet(path string) (*desc.FileDescriptor, error) {
	fds, err := loadFileDescriptorSet(path)
	if err != nil {
		return nil, err
	}
	return desc.CreateFileDescriptorFromSet(fds)
}

func OutputSet(set []*desc.FileDescriptor) {
	printer := protoprint.Printer{}
	printer.PrintProtosToFileSystem(set, "./out")
	fmt.Println("Printed set to ./out")
}
