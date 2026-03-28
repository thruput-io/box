.PHONY: all fmt lint lint-fix test build

GOPATH := $(shell go env GOPATH)
GOFUMPT := $(GOPATH)/bin/gofumpt
GOIMPORTS := $(GOPATH)/bin/goimports

all: fmt lint-fix test build

fmt:
	$(GOFUMPT) -l -w .
	$(GOIMPORTS) -l -w .

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

test:
	go test ./...

build:
	go build ./...
