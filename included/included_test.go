package included

import (
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vbfox/proto-filter/testutils"
)

type keyValuePair struct {
	Key   string
	Value bool
}

type ListOfPairs []keyValuePair

func (l ListOfPairs) Len() int {
	return len([]keyValuePair(l))
}

func (l ListOfPairs) Less(i, j int) bool {
	pi := l[i]
	pj := l[j]

	return strings.Compare(pi.Key, pj.Key) < 0
}

func (l ListOfPairs) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

func mkListOfPairs(m map[string]bool) ListOfPairs {
	result := make(ListOfPairs, len(m))
	i := 0
	for key, value := range m {
		result[i] = keyValuePair{Key: key, Value: value}
		i = i + 1
	}

	return result
}

func mapToString(m map[string]bool) string {
	pairs := mkListOfPairs(m)
	sort.Sort(pairs)

	var str strings.Builder
	for _, pair := range pairs {
		if pair.Value {
			str.WriteString("+ ")
		} else {
			str.WriteString("- ")
		}

		str.WriteString(pair.Key)
		str.WriteString("\n")
	}

	return str.String()
}

func runIncludedTest(t *testing.T, config string, input string, expected string) {
	assert := require.New(t)
	parsedConfig := testutils.ConfFromString(assert, config)
	inputDesc := testutils.DescriptorSetFromString(assert, "test.proto", input)
	result, err := BuildIncluded(inputDesc, parsedConfig)
	assert.NoError(err)
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
