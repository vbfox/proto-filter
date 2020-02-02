package configuration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEmptyFile(t *testing.T) {
	assert := require.New(t)

	yml := ``
	result, err := LoadConfiguration(yml)
	assert.NoError(err)
	assert.NotNil(result)
	assert.NotNil(result.Include)
	assert.Empty(result.Include)
	assert.NotNil(result.Exclude)
	assert.Empty(result.Exclude)
}

func TestSimpleFile(t *testing.T) {
	assert := require.New(t)

	yml := `
include:
    - a
exclude:
    - b
`
	result, err := LoadConfiguration(yml)
	assert.NoError(err)
	assert.NotNil(result)
	assert.Len(result.Include, 1)

	include1 := result.Include[0]
	assert.Equal(include1.Name, "a")
	assert.True(include1.isLeaf())

	assert.Len(result.Exclude, 1)
	exclude1 := result.Exclude[0]
	assert.Equal(exclude1.Name, "b")
	assert.True(exclude1.isLeaf())
}

func TestSampleFile(t *testing.T) {
	assert := require.New(t)

	yml := `
include:
    - simple.proto:
        - SearchResponse
        - SearchRequest:
            - query
exclude:
    - simple.proto:
        - SearchRequest:
            - "*"
`
	result, err := LoadConfiguration(yml)
	assert.NoError(err)
	assert.NotNil(result)
	assert.Len(result.Include, 1)

	include1 := result.Include[0]
	assert.Equal(include1.Name, "simple.proto")
	assert.Len(include1.Children, 2)
	assert.Equal(include1.Children[0].Name, "SearchResponse")
	assert.True(include1.Children[0].isLeaf())
	assert.Equal(include1.Children[1].Name, "SearchRequest")
	assert.Len(include1.Children[1].Children, 1)
	assert.Equal(include1.Children[1].Children[0].Name, "query")
	assert.True(include1.Children[1].Children[0].isLeaf())

	assert.Len(result.Exclude, 1)
	exclude1 := result.Exclude[0]
	assert.Equal(exclude1.Name, "simple.proto")
	assert.Len(exclude1.Children, 1)
	assert.Equal(exclude1.Children[0].Name, "SearchRequest")
	assert.Len(exclude1.Children[0].Children, 1)
	assert.Equal(exclude1.Children[0].Children[0].Name, "*")
	assert.True(exclude1.Children[0].Children[0].isLeaf())
}
