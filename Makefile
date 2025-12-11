.PHONY: build install clean snapshot release test lint fmt check

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -s -w \
	-X saferay/cmd.version=$(VERSION) \
	-X saferay/cmd.commit=$(COMMIT) \
	-X saferay/cmd.date=$(DATE) \
	-X saferay/cmd.builtBy=make

# Format code
fmt:
	gofmt -w -s .
	@echo "✓ Code formatted"

# Run linter
lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...
	@echo "✓ Lint passed"

# Format + lint
check: fmt lint

# Build with lint check
build: check
	go build -ldflags "$(LDFLAGS)" -o saferay .
	@echo "✓ Build complete: saferay $(VERSION)"

# Build without checks (fast)
build-fast:
	go build -ldflags "$(LDFLAGS)" -o saferay .

install: build
	./saferay install

clean:
	rm -f saferay
	rm -rf dist/

test:
	./saferay check

# Local release test (no publish)
snapshot: check
	goreleaser release --snapshot --clean

# Real release (requires git tag + GITHUB_TOKEN)
release: check
	goreleaser release --clean
