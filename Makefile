.DEFAULT_GOAL = cli

# --------------------------------------------------------------------------------
# Building
# --------------------------------------------------------------------------------

CLI              ?= yams
YAMS_INSTALL_DIR ?= /usr/local/bin/

GO_FILES = $(shell find . -type f -name '*.go')
GO_BUILDER ?= go build

.PHONY: build
build:
	go build ./...

.PHONY: cli
cli: $(CLI)

$(CLI): $(GO_FILES)
	go build ./cmd/yams

.PHONY: install
install: $(CLI)
	cp $< $(YAMS_INSTALL_DIR)

.PHONY: clean
clean:
	rm -f $(CLI)
	rm -f coverage.*
	rm -f *.cov
	go clean -testcache

# --------------------------------------------------------------------------------
# Testing
# --------------------------------------------------------------------------------

GO_TEST_FLAGS ?=

ifdef run
	GO_TEST_FLAGS += "-run=$(run)"
endif

# Track coverage of library and server; not CLI / test helpers
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

.PHONY: precommit
precommit: clean test lint cov-report ui-lint ui-test

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
	uv run $< $@

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
		--tags is-yams-test-resource=true \

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
AWS_CONFIG_FILTER     += OR (resourceType='AWS::IAM::Policy' AND resourceName LIKE 'yams-%')

REAL_WORLD_DATA_FILE = testdata/real-world/awsconfig.jsonl

.PHONY: real-world-data
real-world-data:
	mkdir -p testdata/real-world/
	@echo -n 'Real-world data file BEFORE: '
	@test -f $(REAL_WORLD_DATA_FILE) && md5 $(REAL_WORLD_DATA_FILE) || echo 'no file yet'
	@aws configservice select-aggregate-resource-config \
		--region us-east-1 \
		--profile boringcloud \
		--configuration-aggregator-name $(AWS_CONFIG_AGGREGATOR) \
		--expression "SELECT $(AWS_CONFIG_FIELDS) WHERE $(AWS_CONFIG_FILTER)" \
	| jq -c '.Results[] | fromjson' \
	> $(REAL_WORLD_DATA_FILE)
	@echo -n 'Real-world data file AFTER: '
	@md5 $(REAL_WORLD_DATA_FILE)
	
.PHONY: real-world-org-data
real-world-org-data:
	make cli && ./yams dump -target org -out /tmp/org.json && cat /tmp/org.json | jq -c '.[]' > testdata/real-world/org.jsonl

# --------------------------------------------------------------------------------
# UI Development
# --------------------------------------------------------------------------------

UI_DIR = ./ui

.PHONY: ui-deps
ui-deps:
	cd $(UI_DIR) && npm install

.PHONY: ui-dev
ui-dev: ui-clean ui-deps
	cd $(UI_DIR) && npm run dev

.PHONY: ui-build
ui-build:
	cd $(UI_DIR) && npm run build

.PHONY: ui-lint
ui-lint:
	cd $(UI_DIR) && npm run lint

.PHONY: ui-test
ui-test:
	cd $(UI_DIR) && npm test

.PHONY: ui-preview
ui-preview:
	cd $(UI_DIR) && npm run preview

.PHONY: ui-clean
ui-clean:
	rm -rf $(UI_DIR)/node_modules
	rm -rf $(UI_DIR)/dist
