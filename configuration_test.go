package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEmptyGetTopLevel(t *testing.T) {
	config := NewConfiguration(nil, nil)
	result := config.IsIncluded("foo.proto")
	assert.Equal(t, result, UnknownInclusion)
}

func TestEmptyGetChild(t *testing.T) {
	config := NewConfiguration(nil, nil)
	result := config.IsIncluded("foo.proto", "Bar")
	assert.Equal(t, result, UnknownInclusion)
}

func TestFileIncluded(t *testing.T) {
	include := []*FilterTreeNode{
		NewFilterTreeNode("foo.proto"),
	}

	config := NewConfiguration(include, nil)
	result := config.IsIncluded("foo.proto")
	assert.Equal(t, result, IncludedWithChildren)
}

func TestFileIncludedNested(t *testing.T) {
	include := []*FilterTreeNode{
		NewFilterTreeNode("foo.proto", NewFilterTreeNode("Bar")),
	}

	config := NewConfiguration(include, nil)
	result := config.IsIncluded("foo.proto", "Bar")
	assert.Equal(t, result, IncludedWithChildren)
}

func TestElementInFileIncluded(t *testing.T) {
	include := []*FilterTreeNode{
		NewFilterTreeNode("foo.proto"),
	}

	config := NewConfiguration(include, nil)
	result := config.IsIncluded("foo.proto", "bar")
	assert.Equal(t, result, UnknownInclusion)
}
