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
#
# the go version is embedded in the path, so changing Go's version or arch triggers rebuilds.
# other things can be added if necessary, but hopefully only go's formatting behavior matters?
# converts: "go version go1.19.5 darwin/arm64" -> "go1.19.5_darwin_arm64"
BUILD := .build/$(shell go version | cut -d' ' -f3- | sed 's/[^a-zA-Z0-9.]/_/g')
# tools that can be easily re-built on demand, and may be sensitive to dependency or go versions.
# currently this covers all needs.  if not, consider STABLE_BIN like github.com/uber/cadence has.
BIN := $(BUILD)/bin

# current (when committed) version of Go used in CI, and ideally also our docker images.
# this generally does not matter, but can impact goimports or formatting output.
# for maximum stability, make sure you use the same version as CI uses.
#
# this can _likely_ remain a major version, as fmt output does not tend to change in minor versions,
# which will allow findstring to match any minor version.
EXPECTED_GO_VERSION := go1.24
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

# note that vars that do not yet exist are empty, so stick to BUILD/BIN and probably nothing else.
$(BUILD)/lint: $(BUILD)/fmt # lint will fail if fmt (more generally, build) fails, so run it first
$(BUILD)/fmt: $(BUILD)/copyright # formatting must occur only after all other go-file-modifications are done
$(BUILD)/copyright: $(BUILD)/codegen # must add copyright to generated code
$(BUILD)/codegen: $(BUILD)/thrift $(BUILD)/generate
$(BUILD)/generate: $(BUILD)/thrift # go generate broadly requires compile-able code, which needs thrift
$(BUILD)/thrift: $(BUILD)/go_mod_check
$(BUILD)/go_mod_check: | $(BUILD) $(BIN)

# ====================================
# helper vars
# ====================================

PROJECT_ROOT = go.uber.org/cadence

# helper for executing bins that need other bins, just `$(BIN_PATH) the_command ...`
# I'd recommend not exporting this in general, to reduce the chance of accidentally using non-versioned tools.
BIN_PATH := PATH="$(abspath $(BIN)):$$PATH"

# default test args, easy to override
TEST_ARG ?= -v -race

# set a V=1 env var for verbose output. V=0 (or unset) disables.
# this is used to make two verbose flags:
# - $Q, to replace ALL @ use, so CI can be reliably verbose
# - $(verbose), to forward verbosity flags to commands via `$(if $(verbose),-v)` or similar
#
# SHELL='bash -x' is useful too, but can be more confusing to understand.
V ?= 0
ifneq (0,$(V))
verbose := 1
Q :=
else
verbose :=
Q := @
endif

# and enforce ^ that rule: grep the makefile for line-starting @ use, error if any exist.
# limit to one match because multiple look too weird.
_BAD_AT_USE=$(shell grep -n -m1 '^\s*@' $(MAKEFILE_LIST))
ifneq (,$(_BAD_AT_USE))
$(warning Makefile cannot use @ to silence commands, use $$Q instead:)
$(warning found on line $(_BAD_AT_USE))
$(error fix that line and try again)
endif

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

# builds a go-gettable tool, versioned by internal/tools/go.mod, and installs it into
# the build folder, named the same as the last portion of the URL or the second arg.
define go_build_tool
$Q echo "building $(or $(2), $(notdir $(1))) from internal/tools/go.mod..."
$Q go build -mod=readonly -modfile=internal/tools/go.mod -o $(BIN)/$(or $(2), $(notdir $(1))) $(1)
endef

# utility target.
# use as an order-only prerequisite for targets that do not implicitly create these folders.
$(BIN) $(BUILD):
	$Q mkdir -p $@

$(BIN)/thriftrw: internal/tools/go.mod
	$(call go_build_tool,go.uber.org/thriftrw)

$(BIN)/thriftrw-plugin-yarpc: internal/tools/go.mod
	$(call go_build_tool,go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc)

$(BIN)/goimports: internal/tools/go.mod
	$(call go_build_tool,golang.org/x/tools/cmd/goimports)

$(BIN)/revive: internal/tools/go.mod
	$(call go_build_tool,github.com/mgechev/revive)

$(BIN)/staticcheck: internal/tools/go.mod
	$(call go_build_tool,honnef.co/go/tools/cmd/staticcheck)

$(BIN)/errcheck: internal/tools/go.mod
	$(call go_build_tool,github.com/kisielk/errcheck)

$(BIN)/goveralls: internal/tools/go.mod
	$(call go_build_tool,github.com/mattn/goveralls)

$(BIN)/mockery: internal/tools/go.mod
	$(call go_build_tool,github.com/vektra/mockery/v2,mockery)

# copyright header checker/writer.  only requires stdlib, so no other dependencies are needed.
$(BIN)/copyright: internal/tools/licensegen.go
	go build -mod=readonly -o $@ ./internal/tools/licensegen.go

# ensures mod files are in sync for critical packages
$(BUILD)/go_mod_check: go.mod internal/tools/go.mod
	$Q # ensure both have the same apache/thrift replacement
	$Q ./scripts/check-gomod-version.sh go.uber.org/thriftrw $(if $(verbose),-v)
	$Q touch $@

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
$(BUILD)/codegen: | $(BUILD)
	$Q touch $@

THRIFT_FILES := idls/thrift/cadence.thrift idls/thrift/shadower.thrift
# book-keeping targets to build.  one per thrift file.
# idls/thrift/thing.thrift -> .build/go_version/thing.thrift
# the reverse is done in the recipe.
THRIFT_GEN := $(subst idls/thrift,$(BUILD),$(THRIFT_FILES))

# dummy targets to detect when the idls submodule does not exist, to provide a better error message
$(THRIFT_FILES):
	$(call ensure_idl_submodule)

# thrift is done when all sub-thrifts are done.
$(BUILD)/thrift: $(THRIFT_GEN)
	$Q touch $@

# how to generate each thrift book-keeping file.
#
# note that each generated file depends on ALL thrift files - this is necessary because they can import each other.
# ideally this would --no-recurse like the server does, but currently that produces a new output file, and parallel
# compiling appears to work fine.  seems likely it only risks rare flaky builds.
$(THRIFT_GEN): $(THRIFT_FILES) $(BIN)/thriftrw $(BIN)/thriftrw-plugin-yarpc
	$Q echo 'thriftrw for $(subst $(BUILD),idls/thrift,$@)...'
	$Q $(BIN_PATH) $(BIN)/thriftrw \
		--plugin=yarpc \
		--pkg-prefix=$(PROJECT_ROOT)/.gen/go \
		--out=.gen/go \
		$(subst $(BUILD),idls/thrift,$@)
	$Q touch $@

# mockery is quite noisy so it's worth being kinda precise with the files.
# as long as the //go:generate line is in the file that defines the thing to mock, this will auto-discover it.
# if we build any fancier generators, like the server's wrappers, this might need adjusting / switch to completely manual.
$(BUILD)/generate: $(shell grep --files-with-matches -E '^//go:generate' $(ALL_SRC)) $(BIN)/mockery
	$Q $(BIN_PATH) go generate ./...
	$Q touch $@

# ====================================
# other intermediates
# ====================================

# note that LINT_SRC is fairly fake as a prerequisite.
# it's a coarse "you probably don't need to re-lint" filter, nothing more.
$(BUILD)/lint: $(LINT_SRC) $(BIN)/revive | $(BUILD)
	$Q $(BIN)/revive -config revive.toml -exclude './vendor/...' -exclude './.gen/...' -formatter stylish ./...
	$Q touch $@

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
$(BUILD)/fmt: $(ALL_SRC) $(BIN)/goimports
	$Q echo "goimports..."
	$Q # use FRESH_ALL_SRC so it won't miss any generated files produced earlier
	$Q $(BIN)/goimports -local "go.uber.org/cadence" -w $(FRESH_ALL_SRC)
	$Q touch $@
	$Q $(MAYBE_TOUCH_COPYRIGHT)

$(BUILD)/copyright: $(ALL_SRC) $(BIN)/copyright
	$(BIN)/copyright --verifyOnly
	$Q $(eval MAYBE_TOUCH_COPYRIGHT=touch $@)
	$Q touch $@

# ====================================
# developer-oriented targets
#
# many of these share logic with other intermediates, but are useful to make .PHONY for output on demand.
# as the Makefile is fast, it's reasonable to just delete the book-keeping file recursively make.
# this way the effort is shared with future `make` runs.
# ====================================

# "re-make" a target by deleting and re-building book-keeping target(s).
# the + is necessary for parallelism flags to be propagated
define remake
$Q rm -f $(addprefix $(BUILD)/,$(1))
$Q +$(MAKE) --no-print-directory $(addprefix $(BUILD)/,$(1))
endef

.PHONY: build
build: $(BUILD)/fmt ## ensure all packages build
	go build ./...
	$Q # caution: some errors are reported on stdout for some reason
	go test -exec true ./... >/dev/null

.PHONY: lint
# useful to actually re-run to get output again.
# reuse the intermediates for simplicity and consistency.
lint: ## (re)run the linter
	$(call remake,lint)

.PHONY: fmt
# intentionally not re-making, it's clear when it's unnecessary
fmt: $(BUILD)/fmt ## run goimports

.PHONY: copyright
# not identical to the intermediate target, but does provide the same codegen (or more).
copyright: $(BIN)/copyright ## update copyright headers
	$(BIN)/copyright
	$Q touch $(BUILD)/copyright

.PHONY: staticcheck
staticcheck: $(BIN)/staticcheck $(BUILD)/fmt ## (re)run staticcheck
	$(BIN)/staticcheck ./...

.PHONY: errcheck
errcheck: $(BIN)/errcheck $(BUILD)/fmt ## (re)run errcheck
	$(BIN)/errcheck ./...

.PHONY: generate
generate: ## run go-generate (build, lint, fmt, tests all do this for you if needed)
	$(call remake,generate)

.PHONY: tidy
tidy:
	go mod tidy
	cd internal/tools; go mod tidy

.PHONY: all
all: $(BUILD)/lint ## refresh codegen, lint, and ensure everything builds, whatever is necessary

.PHONY: clean
clean:
	$Q # intentionally not using $(BUILD) as that covers only a single version
	rm -Rf .build .gen
	$Q # remove old things (no longer in use).  this can be removed "eventually", when we feel like they're unlikely to exist.
	rm -Rf .bin

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
	$Q make --no-print-directory DEPS_FILTER='$(JQ_DEPS_ONLY_DIRECT)' deps-all

.PHONY: deps-all
deps-all: ## Check for all dependency updates
	$Q go list -u -m -json all \
		| $(JQ_DEPS_AGE) \
		| sort -n

.PHONY: help
help:
	$Q # print help first, so it's visible
	$Q printf "\033[36m%-20s\033[0m %s\n" 'help' 'Prints a help message showing any specially-commented targets'
	$Q # then everything matching "target: ## magic comments"
	$Q cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*:.* ## .*" | awk 'BEGIN {FS = ":.*? ## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' | sort

# v==================== not yet cleaned up =======================v

INTEG_TEST_ROOT := ./test
COVER_ROOT := $(BUILD)/coverage
UT_COVER_FILE := $(COVER_ROOT)/unit_test_cover.out
INTEG_STICKY_OFF_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_off_cover.out
INTEG_STICKY_ON_COVER_FILE := $(COVER_ROOT)/integ_test_sticky_on_cover.out
INTEG_GRPC_COVER_FILE := $(COVER_ROOT)/integ_test_grpc_cover.out

UT_DIRS := $(filter-out $(INTEG_TEST_ROOT)%, $(sort $(dir $(filter %_test.go,$(ALL_SRC)))))

.PHONY: unit_test integ_test_sticky_off integ_test_sticky_on integ_test_grpc cover
test: unit_test integ_test_sticky_off integ_test_sticky_on ## run all tests (requires a running cadence instance)

unit_test: $(BUILD)/fmt ## run all unit tests
	$Q mkdir -p $(COVER_ROOT)
	$Q echo "mode: atomic" > $(UT_COVER_FILE)
	$Q FAIL=""; \
	for dir in $(UT_DIRS); do \
		mkdir -p $(COVER_ROOT)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(COVER_ROOT)/"$$dir"/cover.out || FAIL="$$FAIL $$dir"; \
		cat $(COVER_ROOT)/"$$dir"/cover.out | grep -v "mode: atomic" >> $(UT_COVER_FILE); \
	done; test -z "$$FAIL" || (echo "Failed packages; $$FAIL"; exit 1)
	cat $(UT_COVER_FILE) > .build/cover.out;

integ_test_sticky_off: $(BUILD)/fmt
	$Q mkdir -p $(COVER_ROOT)
	STICKY_OFF=true go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_OFF_COVER_FILE) -coverpkg=./...

integ_test_sticky_on: $(BUILD)/fmt
	$Q mkdir -p $(COVER_ROOT)
	STICKY_OFF=false go test $(TEST_ARG) ./test -coverprofile=$(INTEG_STICKY_ON_COVER_FILE) -coverpkg=./...

integ_test_grpc: $(BUILD)/fmt
	$Q mkdir -p $(COVER_ROOT)
	STICKY_OFF=false go test $(TEST_ARG) ./test -coverprofile=$(INTEG_GRPC_COVER_FILE) -coverpkg=./...

# intermediate product, ci needs a stable output, so use coverage_report.
# running this target requires coverage files to have already been created, e.g. run ^ the above by hand, which happens in ci.
$(COVER_ROOT)/cover.out: $(UT_COVER_FILE) $(INTEG_STICKY_OFF_COVER_FILE) $(INTEG_STICKY_ON_COVER_FILE) $(INTEG_GRPC_COVER_FILE)
	$Q echo "mode: atomic" > $(COVER_ROOT)/cover.out
	cat $(UT_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_OFF_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_STICKY_ON_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out
	cat $(INTEG_GRPC_COVER_FILE) | grep -v "mode: atomic" | grep -v ".gen" >> $(COVER_ROOT)/cover.out

coverage_report: $(COVER_ROOT)/cover.out
	cp $< $@

cover: $(COVER_ROOT)/cover.out
	go tool cover -html=$(COVER_ROOT)/cover.out;
