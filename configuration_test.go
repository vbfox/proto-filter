package main

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestEmptyGetTopLevel(t *testing.T) {
	config := NewConfiguration(nil, nil)
	result := config.isIncluded("foo.proto")
	assert.False(t, result)
}

func TestEmptyGetChild(t *testing.T) {
	config := NewConfiguration(nil, nil)
	result := config.isIncluded("foo.proto", "Bar")
	assert.False(t, result)
}

func TestFileIncluded(t *testing.T) {
	include := []*FilterTreeNode{
		NewFilterTreeNode("foo.proto"),
	}

	config := NewConfiguration(include, nil)
	result := config.isIncluded("foo.proto")
	assert.True(t, result)
}

func TestElementInFileIncluded(t *testing.T) {
	include := []*FilterTreeNode{
		NewFilterTreeNode("foo.proto"),
	}

	config := NewConfiguration(include, nil)
	result := config.isIncluded("foo.proto", "bar")
	assert.True(t, result)
}
