.PHONY: build install clean snapshot release test

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS = -s -w \
	-X saferay/cmd.version=$(VERSION) \
	-X saferay/cmd.commit=$(COMMIT) \
	-X saferay/cmd.date=$(DATE) \
	-X saferay/cmd.builtBy=make

build:
	go build -ldflags "$(LDFLAGS)" -o saferay .

install: build
	./saferay install

clean:
	rm -f saferay
	rm -rf dist/

test:
	./saferay check

# Local release test (no publish)
snapshot:
	goreleaser release --snapshot --clean

# Real release (requires git tag + GITHUB_TOKEN)
release:
	goreleaser release --clean
