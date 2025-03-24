# --------------------------------------------------------------------------------
# General
# --------------------------------------------------------------------------------

GO_TOOL_TARGET ?= ./...

# --------------------------------------------------------------------------------
# Building
# --------------------------------------------------------------------------------

CLI ?= yams

GO_FILES = $(shell find . -type f -name '*.go')
GO_BUILDER ?= go build

.PHONY: build
build:
	$(GO_BUILDER) $(GO_TOOL_TARGET)

.PHONY: build-cli
build-cli: $(CLI)

$(CLI): $(GO_FILES)
	$(GO_BUILDER) ./cmd/...

.PHONY: clean
clean:
	rm -f $(CLI)
	rm -f $(COVERAGE_FILE)
	rm -f $(COVERAGE_REPORT)
	rm -f $(TMPDIR)/yams.*.json

# --------------------------------------------------------------------------------
# Testing
# --------------------------------------------------------------------------------

GO_FORMATTER ?= go fmt

.PHONY: format
format:
	$(GO_FORMATTER) $(GO_TOOL_TARGET)

GO_LINTER      ?= golangci-lint
GO_LINTER_ARGS ?= run

.PHONY: lint
lint:
	$(GO_LINTER) $(GO_LINTER_ARGS)

GO_TEST_RUNNER ?= go test

GO_TEST_FLAGS ?=

.PHONY: test
test:
	$(GO_TEST_RUNNER) $(GO_TEST_FLAGS) $(GO_TOOL_TARGET)

.PHONY: testv
testv: GO_TEST_FLAGS+=-v
testv: test

COVERAGE_FILE   ?= coverage.out
COVERAGE_REPORT ?= coverage.html
GO_COVER_TOOL   ?= go tool cover
GO_COVER_FLAGS  ?= -html $(COVERAGE_FILE)

# Track coverage of library; not helpers or codegen files
COVERAGE_OMIT   ?= '(yams/cmd|yams/internal/testlib)'

.PHONY: cov
cov: $(COVERAGE_FILE)

$(COVERAGE_FILE): GO_TEST_FLAGS+=-coverprofile $(COVERAGE_FILE)
$(COVERAGE_FILE): test
	grep -Ev $(COVERAGE_OMIT) $(COVERAGE_FILE) > $(COVERAGE_FILE).tmp
	mv $(COVERAGE_FILE).tmp $(COVERAGE_FILE)

.PHONY: report
report: $(COVERAGE_FILE)
	$(GO_COVER_TOOL) $(GO_COVER_FLAGS)

$(COVERAGE_REPORT): $(COVERAGE_FILE)
	$(GO_COVER_TOOL) $(GO_COVER_FLAGS) -o $(COVERAGE_REPORT)

.PHONY: html
html: $(COVERAGE_REPORT)

# --------------------------------------------------------------------------------
# IAM data
# --------------------------------------------------------------------------------

BUILD_DATA_DIR ?= ./internal/assets

.PHONY: data
data: sar mp

.PHONY: sar
sar: $(BUILD_DATA_DIR)/sar.json.gz

$(BUILD_DATA_DIR)/sar.json.gz: ./misc/sar.py
	./$< $@

.PHONY: mp
mp: $(BUILD_DATA_DIR)/mp.json.gz

$(BUILD_DATA_DIR)/mp.json.gz: ./misc/mp.py
	./$< $@

.PHONY: clean-data
clean-data:
	rm -rf $(BUILD_DATA_DIR)/*.json.gz
