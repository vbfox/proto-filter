package protofilter

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/stretchr/testify/require"
	"github.com/vbfox/proto-filter/configuration"
)

type ReaderNoCloser struct {
	Reader io.Reader
}

func (x ReaderNoCloser) Read(p []byte) (n int, err error) {
	return x.Reader.Read(p)
}

func (x ReaderNoCloser) Close() error {
	return nil
}

func confFromString(assert *require.Assertions, yml string) *configuration.Configuration {
	result, err := configuration.LoadConfiguration([]byte(yml))
	assert.NoError(err)
	assert.NotNil(result)
	return result
}

func descriptorSetFromString(assert *require.Assertions, path string, content string) []*desc.FileDescriptor {
	buf := bytes.NewBufferString(content)
	parser := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			if filename != path {
				return nil, fmt.Errorf("Found file %s but expecting %s", filename, path)
			}
			reader := ReaderNoCloser{Reader: buf}
			return reader, nil
		},
	}
	desc, err := parser.ParseFiles(path)
	assert.NoError(err)
	assert.NotNil(desc)
	assert.Len(desc, 1)
	return desc
}

type WriterNoCloser struct {
	Writer io.Writer
}

func (x WriterNoCloser) Write(p []byte) (n int, err error) {
	return x.Writer.Write(p)
}

func (x WriterNoCloser) Close() error {
	return nil
}

func fileDescriptorToString(assert *require.Assertions, descriptor *desc.FileDescriptor) string {
	fds := []*desc.FileDescriptor{descriptor}
	p := protoprint.Printer{}
	buf := new(bytes.Buffer)
	firstFile := true
	err := p.PrintProtoFiles(fds, func(name string) (io.WriteCloser, error) {
		if !firstFile {
			return nil, fmt.Errorf("Found file %s but another file has been found before", name)
		}
		firstFile = false
		writer := WriterNoCloser{Writer: buf}
		return writer, nil
	})
	assert.NoError(err)
	return buf.String()
}

func runSimpleTest(t *testing.T, config string, input string, expected string) {
	assert := require.New(t)
	parsedConfig := confFromString(assert, config)
	inputDesc := descriptorSetFromString(assert, "test.proto", input)
	actualDesc, err := FilterSet(inputDesc, parsedConfig)
	assert.NoError(err)
	assert.NotNil(actualDesc)
	assert.Len(actualDesc, 1)
	actual := fileDescriptorToString(assert, actualDesc[0])
	assert.Equal(expected, actual)
}

func TestIncludeEverything(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
	)
}

func TestExcludeField(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
exclude:
  - test.proto:
    - msg_a:
      - field_2
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}
`,
	)
}

func TestPartialIncludeMessage(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}

message msg_b {
    string field_b_1 = 1;
  }
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}
`,
	)
}

func TestPartialIncludeField(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a:
      - field_a_1
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}
`,
	)
}

func TestPartialNested(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a:
      - msg_b
`,
		`syntax = "proto3";

message msg_a {
  message msg_b {
    string field_b_1 = 1;
  }

  string field_a_1 = 1;
}
`,
		`syntax = "proto3";

message msg_a {
  message msg_b {
    string field_b_1 = 1;
  }
}
`,
	)
}

func TestMessageReference(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

message msg_a {
  msg_b field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
		`syntax = "proto3";

message msg_a {
  msg_b field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
	)
}
