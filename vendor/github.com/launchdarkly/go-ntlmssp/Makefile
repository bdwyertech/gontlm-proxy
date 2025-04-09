GOLANGCI_LINT_VERSION=v1.64.5

LINTER=./bin/golangci-lint
LINTER_VERSION_FILE=./bin/.golangci-lint-version-$(GOLANGCI_LINT_VERSION)

OUTPUT_DIR=./build

ALL_SOURCES := $(shell find * -type f -name "*.go")

COVERAGE_PROFILE_RAW=./build/coverage_raw.out
COVERAGE_PROFILE_RAW_HTML=./build/coverage_raw.html
COVERAGE_PROFILE_FILTERED=./build/coverage.out
COVERAGE_PROFILE_FILTERED_HTML=./build/coverage.html
COVERAGE_ENFORCER_FLAGS=-package github.com/launchdarkly/go-ntlmssp \
	-skipfiles '(internal/sharedtest/$(COVERAGE_ENFORCER_SKIP_FILES_EXTRA))' \
	-skipcode "// COVERAGE" \
	-packagestats -filestats -showcode

.PHONY: all build clean test test-coverage benchmarks benchmark-allocs lint

bump-min-go-version:
	go mod edit -go=$(MIN_GO_VERSION) go.mod
	cd ldotel && go mod edit -go=$(MIN_GO_VERSION) go.mod
	cd ldai && go mod edit -go=$(MIN_GO_VERSION) go.mod
	cd testservice && go mod edit -go=$(MIN_GO_VERSION) go.mod
	cd ./.github/variables && sed -i.bak "s#min=[^ ]*#min=$(MIN_GO_VERSION)#g" go-versions.env && rm go-versions.env.bak

clean:
	rm -rf ./bin/

build:
	go build ./...

test:
	go test -v -race ./...

lint:
	$(LINTER) run ./...

test-coverage: $(COVERAGE_PROFILE_RAW)
	go run github.com/launchdarkly-labs/go-coverage-enforcer@latest $(COVERAGE_ENFORCER_FLAGS) -outprofile $(COVERAGE_PROFILE_FILTERED) $(COVERAGE_PROFILE_RAW)
	go tool cover -html $(COVERAGE_PROFILE_FILTERED) -o $(COVERAGE_PROFILE_FILTERED_HTML)
	go tool cover -html $(COVERAGE_PROFILE_RAW) -o $(COVERAGE_PROFILE_RAW_HTML)

$(COVERAGE_PROFILE_RAW): $(ALL_SOURCES)
	@mkdir -p ./build
	go test -coverprofile $(COVERAGE_PROFILE_RAW) -coverpkg=./... ./... >/dev/null
	# note that -coverpkg=./... is necessary so it aggregates coverage to include inter-package references

$(LINTER_VERSION_FILE):
	rm -f $(LINTER)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s $(GOLANGCI_LINT_VERSION)
	touch $(LINTER_VERSION_FILE)

lint: $(LINTER_VERSION_FILE)
