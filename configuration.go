package main

type FilterTreeNode struct {
	name     string
	children []*FilterTreeNode
}

func NewFilterTreeNode(name string, children ...*FilterTreeNode) *FilterTreeNode {
	if children == nil {
		children = []*FilterTreeNode{}
	}
	return &FilterTreeNode{
		name:     name,
		children: children,
	}
}

type Configuration struct {
	include []*FilterTreeNode
	exclude []*FilterTreeNode
}

func NewConfiguration(include []*FilterTreeNode, exclude []*FilterTreeNode) *Configuration {
	if include == nil {
		include = []*FilterTreeNode{}
	}
	if exclude == nil {
		exclude = []*FilterTreeNode{}
	}
	return &Configuration{
		include: include,
		exclude: exclude,
	}
}

func (node *FilterTreeNode) isLeaf() bool {
	return len(node.children) == 0
}

func findTreeNode(nodes []*FilterTreeNode, name string) *FilterTreeNode {
	for _, n := range nodes {
		if n.name == name {
			return n
		}
	}
	return nil
}

func isIncludedCore(include []*FilterTreeNode, exclude []*FilterTreeNode, path []string) bool {
	pathElement := path[0]
	includeElement := findTreeNode(include, pathElement)
	excludeElement := findTreeNode(exclude, pathElement)

	if len(path) == 1 {
		if includeElement != nil && includeElement.isLeaf() {
			// Element explicitely in the include list
			return true
		}
		if excludeElement != nil && excludeElement.isLeaf() {
			// Element explicitely in the exclude list
			return true
		}
		return includeElement != nil
	} else {
		return false
	}
}

func (config *Configuration) isIncluded(path ...string) bool {
	return isIncludedCore(config.include, config.exclude, path)
}
