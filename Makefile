# --------------------------------------------------------------------------------
# General
# --------------------------------------------------------------------------------

GO_TOOL_TARGET ?= ./...

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

GO_FORMATTER ?= go fmt

.PHONY: format
format:
	$(GO_FORMATTER) $(GO_TOOL_TARGET)

GO_LINTER ?= go vet

.PHONY: lint
lint:
	$(GO_LINTER) $(GO_TOOL_TARGET)

GO_TEST_RUNNER ?= go test

.PHONY: test
test:
	$(GO_TEST_RUNNER) $(GO_TOOL_TARGET)

# --------------------------------------------------------------------------------
# Codegen: generating code
# --------------------------------------------------------------------------------

GO_GENERATOR ?= go generate

.PHONY: codegen
codegen: clean-codegen
	$(GO_GENERATOR) $(GO_TOOL_TARGET)
	$(MAKE) format
	@echo 'code generation complete'

.PHONY: clean-codegen
clean-codegen:
	rm -f ./pkg/aws/managedpolicies/zzz_*.go

# --------------------------------------------------------------------------------
# Codegen: fetching data
# --------------------------------------------------------------------------------

BUILD_DATA_DIR        ?= ./builddata
REPO_CLONE_URL        ?= https://github.com/iann0036/iam-dataset.git
REPO_LOCAL_PATH       ?= $(BUILD_DATA_DIR)/clone/iam-dataset
DATA_IAM_DEFINITION   ?= $(BUILD_DATA_DIR)/iam_definition.json
DATA_MANAGED_POLICIES ?= $(BUILD_DATA_DIR)/managed_policies.json

.PHONY: data
data: $(REPO_LOCAL_PATH) $(DATA_IAM_DEFINITION) $(DATA_MANAGED_POLICIES)

$(REPO_LOCAL_PATH):
	git clone --single-branch --depth 1 $(REPO_CLONE_URL) $@

$(DATA_IAM_DEFINITION):
	@echo 'Generating IAM permission dataset'
	@cp $(REPO_LOCAL_PATH)/aws/iam_definition.json $@

$(DATA_MANAGED_POLICIES):
	@echo 'Generating managed policy dataset'
	@cat $(REPO_LOCAL_PATH)/aws/managedpolicies/*.json \
		| jq '. | select(.arn != null) | {arn, document}' \
		| jq -s '.' \
		> $@

.PHONY: clean-data
clean-data:
	rm -rf $(BUILD_DATA_DIR)
