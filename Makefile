default: test

SRCS    := $(shell find . -type f -name '*.go')
LDFLAGS := -ldflags="-s -w"

build: $(SRCS)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS)

test:
	go test -v -parallel 4 -race ./...

.PHONY: build test
