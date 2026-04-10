.PHONY: all fmt lint lint-fix deadcode test coverage build ci

GOPATH := $(shell go env GOPATH)
GOFUMPT := $(GOPATH)/bin/gofumpt
GOIMPORTS := $(GOPATH)/bin/goimports
DEADCODE := $(GOPATH)/bin/deadcode

all: fmt lint-fix lint deadcode test build

fmt:
	$(GOFUMPT) -l -w .
	$(GOIMPORTS) -l -w .

lint:
	golangci-lint run ./...

lint-fix:
	golangci-lint run --fix ./...

deadcode:
	go run scripts/deadcode_fix.go

test:
	go test ./...

coverage:
	bash scripts/coverage.sh

build:
	mkdir -p bin
	go build -o bin/server .
	go build ./...

ci: lint deadcode test coverage build
