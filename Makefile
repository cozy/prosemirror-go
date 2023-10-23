# # Some interesting links on Makefiles:
# https://danishpraka.sh/2019/12/07/using-makefiles-for-go.html
# https://tech.davis-hansson.com/p/make/

MAKEFLAGS += --warn-undefined-variables
MAKEFLAGS += --no-builtin-rules
SHELL := bash

## tests: run the tests
tests:
	@go test ./...
.PHONY: tests

## lint: enforce a consistent code style and detect code smells
lint: bin/golangci-lint
	@bin/golangci-lint run -E gofmt -E unconvert -E misspell -E whitespace -E exportloopref -E bidichk
.PHONY: lint

bin/golangci-lint: Makefile
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- v1.55.0

## clean: clean the generated files and directories
clean:
	@rm -rf bin
.PHONY: clean

## help: print this help message
help:
	@echo "Usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
.PHONY: help
