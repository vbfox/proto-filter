package testutils

import (
	"bytes"
	"fmt"
	"io"

	"github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/jhump/protoreflect/desc/protoprint"
	"github.com/stretchr/testify/require"
	"github.com/vbfox/proto-filter/configuration"
)

type readerNoCloser struct {
	Reader io.Reader
}

func (x readerNoCloser) Read(p []byte) (n int, err error) {
	return x.Reader.Read(p)
}

func (x readerNoCloser) Close() error {
	return nil
}

func ConfFromString(assert *require.Assertions, yml string) *configuration.Configuration {
	result, err := configuration.LoadConfiguration([]byte(yml))
	assert.NoError(err)
	assert.NotNil(result)
	return result
}

func DescriptorSetFromString(assert *require.Assertions, path string, content string) []*desc.FileDescriptor {
	buf := bytes.NewBufferString(content)
	parser := protoparse.Parser{
		Accessor: func(filename string) (io.ReadCloser, error) {
			if filename != path {
				return nil, fmt.Errorf("Found file %s but expecting %s", filename, path)
			}
			reader := readerNoCloser{Reader: buf}
			return reader, nil
		},
	}
	desc, err := parser.ParseFiles(path)
	assert.NoError(err)
	assert.NotNil(desc)
	assert.Len(desc, 1)
	return desc
}

type writerNoCloser struct {
	Writer io.Writer
}

func (x writerNoCloser) Write(p []byte) (n int, err error) {
	return x.Writer.Write(p)
}

func (x writerNoCloser) Close() error {
	return nil
}

func FileDescriptorToString(assert *require.Assertions, descriptor *desc.FileDescriptor) string {
	fds := []*desc.FileDescriptor{descriptor}
	p := protoprint.Printer{}
	buf := new(bytes.Buffer)
	firstFile := true
	err := p.PrintProtoFiles(fds, func(name string) (io.WriteCloser, error) {
		if !firstFile {
			return nil, fmt.Errorf("Found file %s but another file has been found before", name)
		}
		firstFile = false
		writer := writerNoCloser{Writer: buf}
		return writer, nil
	})
	assert.NoError(err)
	return buf.String()
}
