BINARY := cpanel-mcp-ctx
MODULE := github.com/nniksa91/cpanel-mcp-ctx
VERSION ?= 0.1.0-dev

.PHONY: build test lint clean release-local install

build:
	go build -ldflags "-s -w -X main.version=$(VERSION)" -o bin/$(BINARY) ./cmd/cpanel-mcp-ctx

install: build
	install -d $(HOME)/.local/bin
	install -m 755 bin/$(BINARY) $(HOME)/.local/bin/$(BINARY)

test:
	go test ./...

lint:
	go vet ./...
	@command -v golangci-lint >/dev/null && golangci-lint run ./... || echo "install golangci-lint for full lint"

clean:
	rm -rf bin/ dist/

release-local:
	GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/$(BINARY)-linux-amd64 ./cmd/cpanel-mcp-ctx
	GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X main.version=$(VERSION)" -o dist/$(BINARY)-darwin-arm64 ./cmd/cpanel-mcp-ctx
