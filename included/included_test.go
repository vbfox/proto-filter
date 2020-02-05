package included

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vbfox/proto-filter/testutils"
)

func mapToString(m map[string]bool) string {
	var str strings.Builder
	for key, value := range m {
		if value {
			str.WriteString("+ ")
		} else {
			str.WriteString("- ")
		}

		str.WriteString(key)
		str.WriteString("\n")
	}

	return str.String()
}

func runIncludedTest(t *testing.T, config string, input string, expected string) {
	assert := require.New(t)
	parsedConfig := testutils.ConfFromString(assert, config)
	inputDesc := testutils.DescriptorSetFromString(assert, "test.proto", input)
	result := BuildIncluded(inputDesc, parsedConfig)
	actual := mapToString(result)
	assert.Equal(strings.Trim(expected, " \t\r\n"), strings.Trim(actual, " \t\r\n"))
}

func TestIncludeMessage(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_a/field_a_2
`,
	)
}

func TestExcludeField(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
- test.proto/msg_a/field_a_2
`,
	)
}

func TestPartialIncludeMessage(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
- test.proto/msg_b
- test.proto/msg_b/field_b_1
`,
	)
}

func TestPartialIncludeField(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
- test.proto/msg_a/field_a_2
`,
	)
}

func TestPartialNested(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
- test.proto/msg_a/field_a_1
+ test.proto/msg_a/msg_b
+ test.proto/msg_a/msg_b/field_b_1
- test.proto/msg_a/field_a_2
`,
	)
}

func TestMessageReference(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_b
+ test.proto/msg_b/field_b_1
`,
	)
}

func TestMessageReferenceRepeated(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_b
+ test.proto/msg_b/field_b_1
`,
	)
}

func TestMessageReferenceMap(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_b
+ test.proto/msg_b/field_b_1
`,
	)
}

func TestIncludeService(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_b
+ test.proto/msg_b/field_b_1
+ test.proto/msg_b
+ test.proto/svc_a/method_a_1
`,
	)
}

func TestExcludeServiceMethod(t *testing.T) {
	runIncludedTest(
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
		`
+ test.proto
+ test.proto/msg_a
+ test.proto/msg_a/field_a_1
+ test.proto/msg_b
+ test.proto/msg_b/field_b_1
+ test.proto/msg_b
+ test.proto/svc_a/method_a_1
- test.proto/svc_a/method_a_2
+ test.proto/svc_a/method_a_3
`,
	)
}
