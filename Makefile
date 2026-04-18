VERSION    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
BUILD_DATE ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS    := -ldflags "-X main.buildVersion=$(VERSION) -X main.buildDate=$(BUILD_DATE)"
BIN        := circleci

.PHONY: build test lint clean update-golden

build:
	go build $(LDFLAGS) -o bin/$(BIN) ./cmd/circleci

test:
	go test ./...

lint:
	go vet ./...

# Regenerate all golden files after intentional help-text changes.
update-golden:
	UPDATE_GOLDEN=1 go test ./...

clean:
	rm -rf bin/

# Quick smoke test: build + run --help.
smoke: build
	./bin/$(BIN) --help
	@echo "---"
	NO_COLOR=1 ./bin/$(BIN) --help
	@echo "---"
	CI=true ./bin/$(BIN) --help
