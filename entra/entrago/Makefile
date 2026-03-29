ma.PHONY: all fmt lint lint-fix test coverage build ci

GOPATH := $(shell go env GOPATH)
GOFUMPT := $(GOPATH)/bin/gofumpt
GOIMPORTS := $(GOPATH)/bin/goimports

all: fmt lint-fix lint test build

fmt:
	$(GOFUMPT) -l -w .
	$(GOIMPORTS) -l -w .

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

test:
	go test ./...

coverage:
	bash scripts/coverage.sh

build:
	mkdir -p bin
	go build -o bin/server .
	go build ./...

ci: lint test coverage build
