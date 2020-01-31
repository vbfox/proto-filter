.DEFAULT_GOAL := all
.PHONY: clean test build all

SRCS := $(wildcard *.go)

clean:
	rm proto-filter
	rm -rf out

test: ${SRCS}
	go test ./...

proto-filter: ${SRCS}
	go build

build: proto-filter

all: test build
