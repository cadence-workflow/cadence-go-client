# get rid of default behaviors, they're just noise
MAKEFLAGS += --no-builtin-rules
.SUFFIXES:

default: help

# ###########################################
#                TL;DR DOCS:
# ###########################################
# - Targets should never, EVER be *actual source files*.
#   Always use book-keeping files in $(BUILD).
#   Otherwise e.g. changing git branches could confuse Make about what it needs to do.
# - Similarly, prerequisites should be those book-keeping files,
#   not source files that are prerequisites for book-keeping.
#   e.g. depend on .build/fmt, not $(ALL_SRC), and not both.
# - Be strict and explicit about prerequisites / order of execution / etc.
# - Test your changes with `-j 27 --output-sync` or something!
# - Test your changes with `make -d ...`!  It should be reasonable!

# temporary build products and book-keeping targets that are always good to / safe to clean.
BUILD := .build
# less-than-temporary build products, e.g. tools.
# usually unnecessary to clean, and may require downloads to restore, so this folder is not automatically cleaned.
BIN := .bin

# current (when committed) version of Go used in CI, and ideally also our docker images.
# this generally does not matter, but can impact goimports or formatting output.
# for maximum stability, make sure you use the same version as CI uses.
#
# this can _likely_ remain a major version, as fmt output does not tend to change in minor versions,
# which will allow findstring to match any minor version.
EXPECTED_GO_VERSION := go1.16
CURRENT_GO_VERSION := $(shell go version)
ifeq (,$(findstring $(EXPECTED_GO_VERSION),$(CURRENT_GO_VERSION)))
# if you are seeing this warning: consider using https://github.com/travis-ci/gimme to pin your version
$(warning Caution: you are not using CI's go version. Expected: $(EXPECTED_GO_VERSION), current: $(CURRENT_GO_VERSION))
endif

# ====================================
# book-keeping files that are used to control sequencing.
#
# you should use these as prerequisites in almost all cases, not the source files themselves.
# these are defined in roughly the reverse order that they are executed, for easier reading.
#
# recipes and any other prerequisites are defined only once, further below.
# ====================================

# all bins depend on: $(BUILD)/lint
# note that vars that do not yet exist are empty, so any prerequisites defined below are ineffective here.
$(BUILD)/lint: $(BUILD)/fmt $(BUILD)/dummy # lint will fail if fmt or dummy fails, so run them first
$(BUILD)/dummy: $(BUILD)/fmt # do a build after fmt-ing
$(BUILD)/fmt: $(BUILD)/copyright # formatting must occur only after all other go-file-modifications are done
$(BUILD)/copyright: $(BUILD)/codegen # must add copyright to generated code
$(BUILD)/codegen: $(BUILD)/thrift # thrift is currently the only codegen, but this way it's easier to extend
$(BUILD)/thrift:

# ====================================
# helper vars
# ====================================

PROJECT_ROOT = go.uber.org/cadence

# helper for executing bins that need other bins, just `$(BIN_PATH) the_command ...`
# I'd recommend not exporting this in general, to reduce the chance of accidentally using non-versioned tools.
BIN_PATH := PATH="$(abspath $(BIN)):$$PATH"

# default test args, easy to override
TEST_ARG ?= -v -race

# automatically gather all source files that currently exist.
# works by ignoring everything in the parens (and does not descend into matching folders) due to `-prune`,
# and everything else goes to the other side of the `-o` branch, which is `-print`ed.
# this is dramatically faster than a `find . | grep -v vendor` pipeline, and scales far better.
FRESH_ALL_SRC = $(shell \
	find . \
	\( \
		-path './vendor/*' \
		-o -path './idls/*' \
		-o -path './$(BUILD)/*' \
		-o -path './$(BIN)/*' \
	\) \
	-prune \
	-o -name '*.go' -print \
)
# most things can use a cached copy, e.g. all dependencies.
# this will not include any files that are created during a `make` run, e.g. via protoc,
# but that generally should not matter (e.g. dependencies are computed at parse time, so it
# won't affect behavior either way - choose the fast option).
#
# if you require a fully up-to-date list, e.g. for shell commands, use FRESH_ALL_SRC instead.
ALL_SRC := $(FRESH_ALL_SRC)
# as lint ignores generated code, it can use the cached copy in all cases.
LINT_SRC := $(filter-out %_test.go ./.gen/% ./mock% ./tools.go ./internal/compatibility/%, $(ALL_SRC))

# ====================================
# $(BIN) targets
# ====================================

# utility target.
# use as an order-only prerequisite for targets that do not implicitly create these folders.
$(BIN) $(BUILD):
	@mkdir -p $@

$(BIN)/thriftrw: go.mod
	go build -mod=readonly -o $@ go.uber.org/thriftrw

$(BIN)/thriftrw-plugin-yarpc: go.mod
	go build -mod=readonly -o $@ go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc

$(BIN)/golint: go.mod
	go build -mod=readonly -o $@ golang.org/x/lint/golint

$(BIN)/staticcheck: go.mod
	go build -mod=readonly -o $@ honnef.co/go/tools/cmd/staticcheck

$(BIN)/errcheck: go.mod
	go build -mod=readonly -o $@ github.com/kisielk/errcheck

# copyright header checker/writer.  only requires stdlib, so no other dependencies are needed.
$(BIN)/copyright: internal/cmd/tools/copyright/licensegen.go go.mod
	go build -mod=readonly -o $@ ./internal/cmd/tools/copyright/licensegen.go

# dummy binary that ensures most/all packages build, without needing to wait for tests.
$(BUILD)/dummy: $(ALL_SRC) go.mod
	go build -o $@ internal/cmd/dummy/dummy.go

# ====================================
# Codegen targets
# ====================================

# IDL submodule must be populated, or files will not exist -> prerequisites will be wrong -> build will fail.
# Because it must exist before the makefile is parsed, this cannot be done automatically as part of a build.
# Instead: call this func in targets that require the submodule to exist, so that target will not be built.
#
# THRIFT_FILES is just an easy identifier for "the submodule has files", others would work fine as well.
define ensure_idl_submodule
$(if $(wildcard THRIFT_FILES),,$(error idls/ submodule must exist, or build will fail.  Run `git submodule update --init` and try again))
endef

# codegen is done when thrift is done (it's just a naming-convenience, $(BUILD)/thrift would be fine too)
$(BUILD)/codegen: $(BUILD)/thrift | $(BUILD)
	@touch $@

THRIFT_FILES := idls/thrift/cadence.thrift idls/thrift/shadower.thrift
# book-keeping targets to build.  one per thrift file.
# idls/thrift/thing.thrift -> .build/thing.thrift
# the reverse is done in the recipe.
THRIFT_GEN := $(subst idls/thrift/,.build/,$(THRIFT_FILES))

# dummy targets to detect when the idls submodule does not exist, to provide a better error message
$(THRIFT_FILES):
	$(call ensure_idl_submodule)

# thrift is done when all sub-thrifts are done.
$(BUILD)/thrift: $(THRIFT_GEN) | $(BUILD)
	@touch $@

# how to generate each thrift book-keeping file.
#
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
# ideally this would --no-recurse like the server does, but currently that produces a new output file, and parallel
# compiling appears to work fine.  seems likely it only risks rare flaky builds.
$(THRIFT_GEN): $(THRIFT_FILES) $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc | $(BUILD)
	@echo 'thriftrw for $(subst .build/,idls/thrift/,$@)...'
	@$(BIN_PATH) $(BIN)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(PROJECT_ROOT)/.gen/go \
		--out=.gen/go \
		$(subst .build/,idls/thrift/,$@)
	@touch $@

# ====================================
# other intermediates
# ====================================

# golint fails to report many lint failures if it is only given a single file
# to work on at a time, and it can't handle multiple packages at once, *and*
# we can't exclude files from its checks, so for best results we need to give
# it a whitelist of every file in every package that we want linted, per package.
#
# so lint + this golint func works like:
# - iterate over all lintable dirs (outputs "./folder/")
# - find .go files in a dir (via wildcard, so not recursively)
# - filter to only files in LINT_SRC
# - if it's not empty, run golint against the list
define lint_if_present
test -n "$1" && $(BIN)/golint -set_exit_status $1
endef

# TODO: replace this with revive, like the server.
# keep in sync with `lint`
$(BUILD)/lint: $(BIN)/golint $(ALL_SRC)
	@$(foreach pkg, \
		$(sort $(dir $(LINT_SRC))), \
		$(call lint_if_present,$(filter $(wildcard $(pkg)*.go),$(LINT_SRC))) || ERR=1; \
	) test -z "$$ERR"; touch $@; exit $$ERR

# fmt and copyright are mutually cyclic with their inputs, so if a copyright header is modified:
# - copyright -> makes changes
# - fmt sees changes -> makes changes
# - now copyright thinks it needs to run again (but does nothing)
# - which means fmt needs to run again (but does nothing)
# and now after two passes it's finally stable, because they stopped making changes.
#
# this is not fatal, we can just run 2x.
# to be fancier though, we can detect when *both* are run, and re-touch the book-keeping files to prevent the second run.
# this STRICTLY REQUIRES that `copyright` and `fmt` are mutually stable, and that copyright runs before fmt.
# if either changes, this will need to change.
MAYBE_TOUCH_COPYRIGHT=

# TODO: switch to goimports, so we can pin the version
$(BUILD)/fmt: $(ALL_SRC) | $(BUILD)
	@echo "gofmt..."
	@# use FRESH_ALL_SRC so it won't miss any generated files produced earlier
	@gofmt -w $(ALL_SRC)
	@# ideally, mimic server: $(BIN)/goimports -local "go.uber.org/cadence" -w $(FRESH_ALL_SRC)
	@touch $@
	@$(MAYBE_TOUCH_COPYRIGHT)

$(BUILD)/copyright: $(ALL_SRC) $(BIN)/copyright | $(BUILD)
	$(BIN)/copyright --verifyOnly
	@$(eval MAYBE_TOUCH_COPYRIGHT=touch $@)
	@touch $@

$(BUILD)/dummy: $(ALL_SRC) go.mod

# ====================================
# developer-oriented targets
#
# many of these share logic with other intermediates, but are useful to make .PHONY for output on demand.
# as the Makefile is fast, it's reasonable to just delete the book-keeping file recursively make.
# this way the effort is shared with future `make` runs.
# ====================================

.PHONY: build
build: $(BUILD)/dummy ## ensure all packages build

.PHONY: lint
# useful to actually re-run to get output again, as the intermediate will not be run unless necessary.
# dummy is used only because it occurs before $(BUILD)/lint, fmt would likely be sufficient too.
# keep in sync with `$(BUILD)/lint`
lint: $(BUILD)/dummy ## (re)run golint
	@$(foreach pkg, \
		$(sort $(dir $(LINT_SRC))), \
		$(call lint_if_present,$(filter $(wildcard $(pkg)*.go),$(LINT_SRC))) || ERR=1; \
	) test -z "$$ERR"; touch $(BUILD)/lint; exit $$ERR

.PHONY: fmt
# intentionally not re-making, gofmt is slow and it's clear when it's unnecessary
fmt: $(BUILD)/fmt ## run gofmt

.PHONY: copyright
# not identical to the intermediate target, but does provide the same codegen (or more).
copyright: $(BIN)/copyright ## update copyright headers
	$(BIN)/copyright
	@touch $(BUILD)/copyright

.PHONY: staticcheck
staticcheck: $(BIN)/staticcheck $(BUILD)/fmt ## (re)run staticcheck
	$(BIN)/staticcheck ./...

.PHONY: errcheck
errcheck: $(BIN)/errcheck $(BUILD)/fmt ## (re)run errcheck
	$(BIN)/errcheck ./...

.PHONY: all
all: $(BUILD)/lint ## refresh codegen, lint, and ensure the dummy binary builds, if necessary

.PHONY: clean
clean:
	rm -Rf $(BUILD) .gen
	@# remove old things (no longer in use).  this can be removed "eventually", when we feel like they're unlikely to exist.
	rm -Rf .bins dummy

# broken up into multiple += so I can interleave comments.
# this all becomes a single line of output.
# you must not use single-quotes within the string in this var.
JQ_DEPS_AGE = jq '
# only deal with things with updates
JQ_DEPS_AGE += select(.Update)
# allow additional filtering, e.g. DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)'
JQ_DEPS_AGE += $(DEPS_FILTER)
# add "days between current version and latest version"
JQ_DEPS_AGE += | . + {Age:(((.Update.Time | fromdate) - (.Time | fromdate))/60/60/24 | floor)}
# add "days between latest version and now"
JQ_DEPS_AGE += | . + {Available:((now - (.Update.Time | fromdate))/60/60/24 | floor)}
# 123 days: library 	old_version -> new_version
JQ_DEPS_AGE += | ([.Age, .Available] | max | tostring) + " days: " + .Path + "  \t" + .Version + " -> " + .Update.Version
JQ_DEPS_AGE += '
# remove surrounding quotes from output
JQ_DEPS_AGE += --raw-output

# exclude `"Indirect": true` dependencies.  direct ones have no "Indirect" key at all.
JQ_DEPS_ONLY_DIRECT = | select(has("Indirect") | not)

.PHONY: deps
deps: ## Check for dependency updates, for things that are directly imported
	@make --no-print-directory DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)' deps-all

.PHONY: deps-all
deps-all: ## Check for all dependency updates
	@go list -u -m -json all \
		| $(JQ_DEPS_AGE) \
		| sort -n

.PHONY: help
help:
	@# print help first, so it's visible
	@printf "\033[36m%-20s\033[0m %s\n" 'help' 'Prints a help message showing any specially-commented targets'
	@# then everything matching "target: ## magic comments"
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*:.* ## .*" | awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

# v==================== not yet cleaned up =======================v

INTEG_TEST_ROOT := ./test
COVER_ROOT := $(BUILD)/coverage
UT_COVER_FILE := $(COVER_ROOT)/unit_test_cover.out
INTEG_STICKY_OFF_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_off_cover.out
INTEG_STICKY_ON_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_on_cover.out
INTEG_GRPC_COVER_FILE := $(COVER_ROOT)/integ_test_grpc_cover.out

UT_DIRS := $(filter-out $(INTEG_TEST_ROOT)%, $(sort $(dir $(filter %_test.go,$(ALL_SRC)))))

.PHONY: unit_test integ_test_sticky_off integ_test_sticky_on integ_test_grpc cover cover_ci
test: unit_test integ_test_sticky_off integ_test_sticky_on ## run all tests (requires a running cadence instance)

unit_test: $(ALL_SRC) ## run all unit tests
	@mkdir -p $(COVER_ROOT)
	@echo "mode: atomic" > $(UT_COVER_FILE)
	@failed=0; \
	for dir in $(UT_DIRS); do \
		mkdir -p $(COVER_ROOT)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(COVER_ROOT)/"$$dir"/cover.out || failed=1; \
		cat $(COVER_ROOT)/"$$dir"/cover.out | grep -v "mode: atomic" >> $(UT_COVER_FILE); \
	done; \
	exit $$failed

integ_test_sticky_off: $(ALL_SRC)
	@mkdir -p $(COVER_ROOT)
	STICKY_OFF=true go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_OFF_COVER_FILE) -coverpkg=./...

integ_test_sticky_on: $(ALL_SRC)
	@mkdir -p $(COVER_ROOT)
	STICKY_OFF=false go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_ON_COVER_FILE) -coverpkg=./...

integ_test_grpc: $(ALL_SRC)
	@mkdir -p $(COVER_ROOT)
	STICKY_OFF=false go test $(TEST_ARG) ./test -coverprofile=$(INTEG_GRPC_COVER_FILE) -coverpkg=./...

$(COVER_ROOT)/cover.out: $(UT_COVER_FILE) $(INTEG_STICKY_OFF_COVER_FILE) $(INTEG_STICKY_ON_COVER_FILE) $(INTEG_GRPC_COVER_FILE)
	@echo "mode: atomic" > $(COVER_ROOT)/cover.out
	cat $(UT_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_OFF_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_ON_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_GRPC_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out

cover: $(COVER_ROOT)/cover.out
	go tool cover -html=$(COVER_ROOT)/cover.out;

cover_ci: $(COVER_ROOT)/cover.out
	goveralls -coverprofile=$(COVER_ROOT)/cover.out -service=buildkite || echo -e "\x1b[31mCoveralls failed\x1b[m";
