package configuration

import (
	"fmt"
	"github.com/goccy/go-yaml"
)

func Foo() bool {
	yml := `---
foo: 1
bar: c
`
	var v struct {
		A int    `yaml:"foo"`
		B string `yaml:"bar"`
	}
	if err := yaml.Unmarshal([]byte(yml), &v); err != nil {
		return true
	}

	return false
}

type filterTreeNodeYaml struct {
}

type configurationYaml struct {
	Include []filterTreeNodeYaml `yaml:"include"`
	Exclude []filterTreeNodeYaml `yaml:"exclude"`
}

func LoadConfiguration(content string) (*Configuration, error) {
	var config configurationYaml
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}

	fmt.Println(config)

	result := NewConfiguration(nil, nil)
	return result, nil
}
