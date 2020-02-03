package main

import (
	"fmt"

	"github.com/jhump/protoreflect/desc"
	protofilter "github.com/vbfox/proto-filter"
	"github.com/vbfox/proto-filter/configuration"
)

func main() {
	set, err := protofilter.LoadProtoSet("../../test_files/simple.fdset")
	if err != nil {
		fmt.Println("ERR Proto load:", err.Error())
		return
	}

	config, err := configuration.LoadConfigurationFile("../../test_files/simple.yml")
	if err != nil {
		fmt.Println("ERR Conf load:", err.Error())
		return
	}

	fmt.Println("Loaded set", set.GetFullyQualifiedName())
	filtered, err := protofilter.FilterSet([]*desc.FileDescriptor{set}, config)
	if err != nil {
		fmt.Println("ERR Filter:", err.Error())
		return
	}
	protofilter.OutputSet(filtered)
}
