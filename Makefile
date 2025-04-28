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

ifdef run
	GO_TEST_FLAGS += "-run=$(run)"
endif

# Track coverage of library; not test helpers
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
	@echo "\n-> Source code (tests excluded)"
	@cloc --quiet --include-lang=Go --not-match-f '.*_test.go' . | grep -v cloc
	@echo "\n-> Test files"
	@cloc --quiet --include-lang=Go --match-f '.*_test.go' . | grep -v cloc

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
# Debugging
# --------------------------------------------------------------------------------

ifdef pkg
	GO_DEBUG_TARGETS += "$(pkg)"
endif

ifdef run
	GO_DEBUG_FLAGS += "-test.run=$(run)"
endif

.PHONY: debug
debug:
	dlv test $(GO_DEBUG_TARGETS) -- $(GO_DEBUG_FLAGS)

# --------------------------------------------------------------------------------
# IAM data
# --------------------------------------------------------------------------------

BUILD_DATA_DIR ?= ./internal/assets

.PHONY: data
data: sar mp

.PHONY: sar
sar: $(BUILD_DATA_DIR)/sar.json.gz

$(BUILD_DATA_DIR)/sar.json.gz: ./misc/sar_v2.py
	./$< $@

.PHONY: mp
mp: $(BUILD_DATA_DIR)/mp.json.gz

$(BUILD_DATA_DIR)/mp.json.gz: ./misc/mp.py
	./$< $@

.PHONY: clean-data
clean-data:
	rm -rf $(BUILD_DATA_DIR)/*.json.gz

# --------------------------------------------------------------------------------
# CloudFormation
# --------------------------------------------------------------------------------

CF_STACK_NAME ?= yams-test-infra

.PHONY: cf
cf: cf-account-0 cf-account-1 cf-account-2

.PHONY: cf-account-0
cf-account-0: misc/cf-templates/account-0.template.yaml
	aws cloudformation deploy \
		--profile yams0 \
		--region us-east-1 \
		--template-file $< \
		--stack-name $(CF_STACK_NAME) \
		--capabilities CAPABILITY_NAMED_IAM \
		--no-fail-on-empty-changeset \
		--disable-rollback
