BINARY_NAME=gh-deps
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildTime=${BUILD_TIME}"

.PHONY: build
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} cmd/gh-deps/main.go

.PHONY: install
install:
	go install ${LDFLAGS} ./cmd/gh-deps

.PHONY: test
test:
	go test -v ./...

.PHONY: test-coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -f ${BINARY_NAME}
	rm -f coverage.out

.PHONY: deps
deps:
	go mod download
	go mod tidy

.PHONY: gh-install
gh-install: build
	@echo "Installing as gh extension..."
	@echo "Note: Make sure this directory is registered with 'gh extension install .'"

.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  install        - Install to GOPATH"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  lint           - Run linter"
	@echo "  clean          - Remove build artifacts"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  gh-install     - Build and prepare for gh extension installation"
	@echo "  help           - Show this help message"
