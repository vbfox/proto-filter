package protofilter

import (
	"testing"

	"github.com/stretchr/testify/require"
	. "github.com/vbfox/proto-filter/testutils"
)

func runSimpleTest(t *testing.T, config string, input string, expected string) {
	assert := require.New(t)
	parsedConfig := ConfFromString(assert, config)
	inputDesc := DescriptorSetFromString(assert, "test.proto", input)
	actualDesc, err := FilterSet(inputDesc, parsedConfig)
	assert.NoError(err)
	assert.NotNil(actualDesc)
	assert.Len(actualDesc, 1)
	actual := FileDescriptorToString(assert, actualDesc[0])
	assert.Equal(expected, actual)
}

func TestIncludeMessage(t *testing.T) {
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
      - field_a_2
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

func TestMessageReferenceRepeated(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

message msg_a {
  repeated msg_b field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
		`syntax = "proto3";

message msg_a {
  repeated msg_b field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
	)
}

func TestMessageReferenceMap(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

message msg_a {
  map<string, msg_b> field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
		`syntax = "proto3";

message msg_a {
  map<string, msg_b> field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}
`,
	)
}

func TestOptionsAreKept(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

option java_package = "com.example.foo";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
		`syntax = "proto3";

option java_package = "com.example.foo";

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
	)
}

func TestPackagesAreKept(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - msg_a
`,
		`syntax = "proto3";

package pkg;

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
		`syntax = "proto3";

package pkg;

message msg_a {
  string field_a_1 = 1;

  string field_a_2 = 2;
}
`,
	)
}

func TestIncludeService(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - svc_a
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}

service svc_a {
  rpc method_a_1 ( msg_a ) returns ( msg_b );
}
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}

service svc_a {
  rpc method_a_1 ( msg_a ) returns ( msg_b );
}
`,
	)
}

func TestExcludeServiceMethod(t *testing.T) {
	runSimpleTest(
		t,
		`---
include:
  - test.proto:
    - svc_a
exclude:
  - test.proto
    - svc_a:
      - method_a_2
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}

message msg_c {
    string field_c_1 = 1;
  }

service svc_a {
  rpc method_a_1 ( msg_a ) returns ( msg_b );

  rpc method_a_2 ( msg_a ) returns ( msg_c );

  rpc method_a_3 ( msg_a ) returns ( msg_b );
}
`,
		`syntax = "proto3";

message msg_a {
  string field_a_1 = 1;
}

message msg_b {
  string field_b_1 = 1;
}

service svc_a {
  rpc method_a_1 ( msg_a ) returns ( msg_b );

  rpc method_a_3 ( msg_a ) returns ( msg_b );
}
`,
	)
}
