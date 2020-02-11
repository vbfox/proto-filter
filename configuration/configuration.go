package configuration

import "fmt"

type FilterTreeNode struct {
	Name     string
	Children []*FilterTreeNode
}

func NewFilterTreeNode(name string, children ...*FilterTreeNode) *FilterTreeNode {
	if children == nil {
		children = []*FilterTreeNode{}
	}
	return &FilterTreeNode{
		Name:     name,
		Children: children,
	}
}

type Configuration struct {
	Include []*FilterTreeNode
	Exclude []*FilterTreeNode
}

func NewConfiguration(include []*FilterTreeNode, exclude []*FilterTreeNode) *Configuration {
	if include == nil {
		include = []*FilterTreeNode{}
	}
	if exclude == nil {
		exclude = []*FilterTreeNode{}
	}
	return &Configuration{
		Include: include,
		Exclude: exclude,
	}
}

type InclusionResult int

const (
	UnknownInclusion        InclusionResult = 0
	IncludedWithoutChildren InclusionResult = 1
	IncludedWithChildren    InclusionResult = 2
	Excluded                InclusionResult = 3
)

func (node *FilterTreeNode) isLeaf() bool {
	return len(node.Children) == 0
}

func findTreeNode(nodes []*FilterTreeNode, name string) *FilterTreeNode {
	if nodes == nil {
		return nil
	}

	for _, n := range nodes {
		if n.Name == name {
			return n
		}
	}
	return nil
}

func isIncludedCore(include []*FilterTreeNode, exclude []*FilterTreeNode, path []string) InclusionResult {
	pathElement := path[0]
    includeElement := findTreeNode(include, pathElement)
    excludeElement := findTreeNode(exclude, pathElement)
    fmt.Printf("IsIncluded %v, Included=%v, Excluded=%v\n", path, includeElement, excludeElement)

	if len(path) == 1 {
		if includeElement != nil && includeElement.isLeaf() {
			// Element explicitely in the include list
			return IncludedWithChildren
		}
		if excludeElement != nil && excludeElement.isLeaf() {
			// Element explicitely in the exclude list
			return Excluded
		}
		if includeElement != nil {
			return IncludedWithoutChildren
		}
		return UnknownInclusion
	}

	if includeElement == nil && excludeElement == nil {
		return UnknownInclusion
	}

	childrenPath := path[1:]
	var childrenInclude []*FilterTreeNode = nil
	if includeElement != nil {
		childrenInclude = includeElement.Children
	}
	var childrenExclude []*FilterTreeNode = nil
	if excludeElement != nil {
		childrenExclude = excludeElement.Children
	}

	return isIncludedCore(childrenInclude, childrenExclude, childrenPath)
}

func (config *Configuration) IsIncluded(path ...string) InclusionResult {
	return isIncludedCore(config.Include, config.Exclude, path)
}
