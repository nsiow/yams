.DEFAULT_GOAL = build

# --------------------------------------------------------------------------------
# Building
# --------------------------------------------------------------------------------

CLI ?= yams

GO_FILES = $(shell find . -type f -name '*.go')
GO_BUILDER ?= go build

.PHONY: build
build:
	go build ./...

.PHONY: build-cli
build-cli: $(CLI)

$(CLI): $(GO_FILES)
	go build ./cmd/...

.PHONY: clean
clean:
	rm -f $(CLI)
	rm -f coverage.*
	go clean -testcache

# --------------------------------------------------------------------------------
# Testing
# --------------------------------------------------------------------------------

GO_TEST_FLAGS ?=

# Track coverage of library; not helpers or codegen files
COVERAGE_OMIT ?= '(yams/cmd|yams/internal/testlib)'

.PHONY: format
format:
	go fmt ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	go test $(GO_TEST_FLAGS) ./...

.PHONY: testv
testv: GO_TEST_FLAGS+=-v
testv: test

.PHONY: testcount
testcount:
	@echo "Ran <$$(make testv | grep '=== RUN' | wc -l)> tests"

.PHONY: loc
loc:
	cloc --include-lang=Go --not-match-f '.*_test.go' .

.PHONY: cov
cov: coverage.out

.PHONY: cov-report
cov-report: coverage.out
	go tool cover -func=$<

.PHONY: cov-missing
cov-missing: coverage.out
	@go tool cover -func=$< | grep -v '100.0%' || echo '[✔] code coverage = 100.0'

.PHONY: cov-html
cov-html: coverage.html

coverage.out: $(GO_FILES)
	GO_TEST_FLAGS='-coverprofile=$@' make test
	grep -Ev $(COVERAGE_OMIT) $@ > $@.tmp
	mv $@.tmp $@

coverage.html: coverage.out
	go tool cover -html=$< -o $@

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

.PHONY: sar_v2
sar_v2: $(BUILD_DATA_DIR)/sar_v2.json.gz

$(BUILD_DATA_DIR)/sar_v2.json.gz: ./misc/sar_v2.py
	./$< $@

.PHONY: mp
mp: $(BUILD_DATA_DIR)/mp.json.gz

$(BUILD_DATA_DIR)/mp.json.gz: ./misc/mp.py
	./$< $@

.PHONY: clean-data
clean-data:
	rm -rf $(BUILD_DATA_DIR)/*.json.gz
