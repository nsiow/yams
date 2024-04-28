# --------------------------------------------------------------------------------
# Building
# --------------------------------------------------------------------------------

CLI ?= yams

GO_BUILDER ?= go build

.PHONY: build-cli
build-cli: $(CLI)

$(CLI): $(shell find . -type f -name '*.go')
	$(GO_BUILDER) ./cmd/...

.PHONY: clean
clean:
	rm -f $(CLI)

# --------------------------------------------------------------------------------
# Testing
# --------------------------------------------------------------------------------

GO_LINTER ?= go vet

.PHONY: lint
lint:
	$(GO_LINTER) ./...

GO_TEST_RUNNER ?= go test

.PHONY: test
test:
	$(GO_TEST_RUNNER) ./...
