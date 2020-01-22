package main

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/golang/protobuf/proto"
	dpb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/jhump/protoreflect/desc"
	_ "github.com/jhump/protoreflect/desc/builder"
)

func loadProtoset(path string) (*desc.FileDescriptor, error) {
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
	return desc.CreateFileDescriptorFromSet(&fds)
}

func main() {
	set, err := loadProtoset("test_files/simple.fdset")
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

	fmt.Println("Loaded set", set.GetFullyQualifiedName())
}
