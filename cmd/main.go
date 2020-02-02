package main

import (
	"fmt"

	protofilter "github.com/vbfox/proto-filter"
)

func main() {
	set, err := protofilter.LoadProtoSet("test_files/simple.fdset")
	if err != nil {
		fmt.Println("Error:", err.Error())
	}

	fmt.Println("Loaded set", set.GetFullyQualifiedName())
	protofilter.OutputSet(set)
}
