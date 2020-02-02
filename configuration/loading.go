package configuration

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/goccy/go-yaml"
	"github.com/goccy/go-yaml/ast"
	"github.com/goccy/go-yaml/lexer"
	"github.com/goccy/go-yaml/parser"
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

type filterTreeYaml struct {
	Name     string
	Children []*filterTreeYaml
}

func (v *filterTreeYaml) ToFilterTree() *FilterTreeNode {
	var children []*FilterTreeNode

	for _, childNode := range v.Children {
		children = append(children, childNode.ToFilterTree())
	}

	return NewFilterTreeNode(v.Name, children...)
}

func prettyPrint(i interface{}) string {
	s, _ := json.MarshalIndent(i, "", "\t")
	return string(s)
}

func yamlNodeToFilterTree(node ast.Node) (*filterTreeYaml, error) {
	if node == nil {
		return nil, nil
	}

	switch node.Type() {
	case ast.StringType:
		stringNode := node.(*ast.StringNode)
		name := stringNode.Value
		return &filterTreeYaml{Name: name}, nil

	case ast.IntegerType:
		integerNode := node.(*ast.IntegerNode)
		name := string(integerNode.GetValue().(int64))
		return &filterTreeYaml{Name: name}, nil

	case ast.MappingValueType:
		mappingValueNode := node.(*ast.MappingValueNode)
		if mappingValueNode.Key.Type() != ast.StringType {
			return nil, fmt.Errorf("Expected a string key but found: %v", mappingValueNode.Key.Type())
		}
		if mappingValueNode.Value.Type() != ast.SequenceType {
			return nil, fmt.Errorf("Expected a sequence of values but found: %v", mappingValueNode.Value.Type())
		}

		keyNode := mappingValueNode.Key.(*ast.StringNode)
		valueNode := mappingValueNode.Value.(*ast.SequenceNode)
		name := keyNode.Value

		var children []*filterTreeYaml

		for _, childNode := range valueNode.Values {
			child, childErr := yamlNodeToFilterTree(childNode)
			if childErr != nil {
				return nil, fmt.Errorf("%s: %w", name, childErr)
			}

			if child != nil {
				children = append(children, child)
			}
		}

		return &filterTreeYaml{Name: name, Children: children}, nil
	default:
		return nil, fmt.Errorf("Not supported key type: %v", node.Type())
	}
}

func deserializeFilterTreeYaml(raw string) (*filterTreeYaml, error) {
	tokens := lexer.Tokenize(raw)
	f, err := parser.Parse(tokens, 0)
	if err != nil {
		return nil, fmt.Errorf("YAML parsing failed: %w", err)
	}

	for _, doc := range f.Docs {
		return yamlNodeToFilterTree(doc.Body)
	}

	return nil, fmt.Errorf("No document found in YAML")
}

func (v *filterTreeYaml) UnmarshalYAML(raw []byte) error {
	s := string(raw)
	vv, err := deserializeFilterTreeYaml(s)
	if err != nil {
		return fmt.Errorf("deserializeFilterTreeYaml failed: %w", err)
	}

	*v = *vv

	return nil
}

type configurationYaml struct {
	Include []*filterTreeYaml `yaml:"include"`
	Exclude []*filterTreeYaml `yaml:"exclude"`
}

func filterTreeYamlArrayToFilterTreeArray(yaml []*filterTreeYaml) []*FilterTreeNode {
	var result []*FilterTreeNode

	for _, childNode := range yaml {
		result = append(result, childNode.ToFilterTree())
	}

	return result
}

func LoadConfiguration(content []byte) (*Configuration, error) {
	var config configurationYaml
	if err := yaml.Unmarshal([]byte(content), &config); err != nil {
		return nil, err
	}

	result := NewConfiguration(
		filterTreeYamlArrayToFilterTreeArray(config.Include),
		filterTreeYamlArrayToFilterTreeArray(config.Exclude),
	)

	return result, nil
}

func LoadConfigurationFile(path string) (*Configuration, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Can't load file %s: %w", path, err)
	}

	return LoadConfiguration(content)
}
