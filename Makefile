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
# Testing Infrastructure (CloudFormation)
# --------------------------------------------------------------------------------

CF_STACK_NAME   ?= yams-test-infra
CF_STACK_REGION ?= us-east-1

CF_DEPLOY = aws cloudformation deploy \
		--parameter-overrides \
		    AccountId0=$(YAMS_TEST_ACCOUNT_ID_0) \
		    AccountId1=$(YAMS_TEST_ACCOUNT_ID_1) \
		    AccountId2=$(YAMS_TEST_ACCOUNT_ID_2) \
		--region $(CF_STACK_REGION) \
		--stack-name $(CF_STACK_NAME) \
		--capabilities CAPABILITY_NAMED_IAM \
		--no-fail-on-empty-changeset \
		--disable-rollback \
		--tags is-yams-test-resource=true

.PHONY: cf
cf: cf-account-0 cf-account-1 cf-account-2

.PHONY: cf-account-0
cf-account-0: misc/cf-templates/account-0.template.yaml
	$(CF_DEPLOY) --profile yams0 --template-file $<

.PHONY: cf-account-1
cf-account-1: misc/cf-templates/account-1.template.yaml
	$(CF_DEPLOY) --profile yams1 --template-file $<

.PHONY: cf-account-2
cf-account-2: misc/cf-templates/account-2.template.yaml
	$(CF_DEPLOY) --profile yams2 --template-file $<

# --------------------------------------------------------------------------------
# Testing Infrastructure (Config)
# --------------------------------------------------------------------------------

AWS_CONFIG_AGGREGATOR ?= boringcloud-awsconfig-aggregator 
AWS_CONFIG_FIELDS     ?= *, configuration, supplementaryConfiguration, tags
AWS_CONFIG_FILTER     ?= tags.tag='is-yams-test-resource=true'

.PHONY: real-world-data
real-world-data:
	mkdir -p testdata/real-world/
	aws configservice select-aggregate-resource-config \
		--region us-east-1 \
		--profile boringcloud \
		--configuration-aggregator-name $(AWS_CONFIG_AGGREGATOR) \
		--expression "SELECT $(AWS_CONFIG_FIELDS) WHERE $(AWS_CONFIG_FILTER)" \
	| jq -c '.Results[] | fromjson' \
	> testdata/real-world/awsconfig.jsonl
