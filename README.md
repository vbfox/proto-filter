# proto-filter

## Configuration

All samples here assume the following content in simple.proto :

```protobuf
syntax = "proto3";

message SearchRequest {
  string query = 1;
  int32 page_number = 2;
  int32 result_per_page = 3;
}

message SearchResponse {
  repeated Result results = 1;
  SearchRequest request = 2;
}

message Result {
  string url = 1;
  string title = 2;
  repeated string snippets = 3;
}
```

Configuration contain 2 sections :

```yaml
include:
    - simple.proto:
        - SearchResponse
exclude:
    - simple.proto:
        - SearchResponse:
            - useless
        - SearchRequest:
            - page_number
        - Result:
            - snippets
```

* The default is to include nothing (An empty file is generated)
* Leafs specified in `include` are included for example the `simple.proto/SearchResponse` message will be included.
* Parents of included elements are included (A `simple.proto` file will be created) but not their content
* Anything inside an included element is implicitely included (Fields & Sub messages)
* Any message referenced by an included element is included implicitely (Recursively)


### The file hierarchy

Files are specified by their file names and path from the root of the proto path.

### Specifying fields

Fields can be specified in multiple ways:

* Designating all fields with a start (`*`)
* By field name (`result_per_page`)
* By field number (`3`)
* Designating all fields with a start (`*`)

### Selecting the fields of an implicitly included message

This specific configuration:

```yaml
include:
    - simple.proto:
        - SearchRequest:
            - query
```

Will produce :

```protobuf
syntax = "proto3";

message SearchRequest {
  string query = 1;
}
```

But 

```yaml
include:
    - simple.proto:
        - SearchResponse
        - SearchRequest:
            - query
```

Will keep all fields of `SearchRequest` because `SearchResponse` include a field of this type so it's implicitely including all the fields.

To keep only `query` all fields need to be explicitely excluded:

```yaml
include:
    - simple.proto:
        - SearchResponse
        - SearchRequest:
            - query
exclude:
    - simple.proto:
        - SearchRequest:
            - *
```
