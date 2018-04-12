.PHONY: test bins clean cover cover_ci
PROJECT_ROOT = go.uber.org/cadence

export PATH := $(GOPATH)/bin:$(PATH)

THRIFT_GENDIR=.gen

# default target
default: test

# define the list of thrift files the service depends on
# (if you have some)
THRIFTRW_SRCS = \
  idl/github.com/uber/cadence/cadence.thrift \
  idl/github.com/uber/cadence/shared.thrift \

PROGS = cadence-client
TEST_ARG ?= -race -v -timeout 5m
BUILD := ./build

THRIFT_GEN=$(GOPATH)/bin/thrift-gen

define thriftrwrule
THRIFTRW_GEN_SRC += $(THRIFT_GENDIR)/go/$1/$1.go

$(THRIFT_GENDIR)/go/$1/$1.go:: $2
	@mkdir -p $(THRIFT_GENDIR)/go
	$(ECHO_V)thriftrw --plugin=yarpc --pkg-prefix=$(PROJECT_ROOT)/$(THRIFT_GENDIR)/go/ --out=$(THRIFT_GENDIR)/go $2
endef

$(foreach tsrc,$(THRIFTRW_SRCS),$(eval $(call \
	thriftrwrule,$(basename $(notdir \
	$(shell echo $(tsrc) | tr A-Z a-z))),$(tsrc))))

# Automatically gather all srcs
ALL_SRC := $(shell find . -name "*.go" | grep -v -e Godeps -e vendor \
	-e ".*/\..*" \
	-e ".*/_.*")

# Files that needs to run lint, exclude testify mock from lint
LINT_SRC := $(filter-out ./mock%,$(ALL_SRC))

# all directories with *_test.go files in them
TEST_DIRS := $(sort $(dir $(filter %_test.go,$(ALL_SRC))))

vendor:
	glide install

yarpc-install: vendor
	go get './vendor/go.uber.org/thriftrw'
	go get './vendor/go.uber.org/yarpc/encoding/thrift/thriftrw-plugin-yarpc'

clean_thrift:
	rm -rf .gen

thriftc: clean_thrift yarpc-install $(THRIFTRW_GEN_SRC)

copyright: ./internal/cmd/tools/copyright/licensegen.go
	go run ./internal/cmd/tools/copyright/licensegen.go --verifyOnly

vendor/glide.updated: glide.lock glide.yaml
	glide install
	touch vendor/glide.updated

dummy: vendor/glide.updated $(ALL_SRC)
	go build -i -o dummy internal/cmd/dummy/dummy.go

test: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(TEST_DIRS); do \
		go test -race -coverprofile=$@ "$$dir" | tee -a test.log; \
	done;

bins: thriftc copyright lint dummy

cover_profile: clean copyright lint vendor/glide.updated
	@mkdir -p $(BUILD)
	@echo "mode: atomic" > $(BUILD)/cover.out

	@echo Testing packages:
	@for dir in $(TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
		cat $(BUILD)/"$$dir"/coverage.out | grep -v "mode: atomic" >> $(BUILD)/cover.out; \
	done;

cover: cover_profile
	go tool cover -html=$(BUILD)/cover.out;

cover_ci: cover_profile
	goveralls -coverprofile=$(BUILD)/cover.out -service=travis-ci || echo -e "\x1b[31mCoveralls failed\x1b[m";

# golint fails to report many lint failures if it is only given a single file
# to work on at a time.  and we can't exclude files from its checks, so for
# best results we need to give it a whitelist of every file in every package
# that we want linted.
#
# so lint + this golint func works like:
# - iterate over all dirs (outputs "./folder/")
# - find .go files in a dir (via wildcard, so not recursively)
# - filter to only files in LINT_SRC
# - if it's not empty, run golint against the list
define lint_if_present
test -n "$1" && golint -set_exit_status $1
endef

lint:
	@$(foreach pkg,\
		$(sort $(dir $(LINT_SRC))), \
		$(call lint_if_present,$(filter $(wildcard $(pkg)*.go),$(LINT_SRC))) || ERR=1; \
	) test -z "$$ERR" || exit 1
	@OUTPUT=`gofmt -l $(ALL_SRC) 2>&1`; \
	if [ "$$OUTPUT" ]; then \
		echo "Run 'make fmt'. gofmt must be run on the following files:"; \
		echo "$$OUTPUT"; \
		exit 1; \
	fi

fmt:
	@gofmt -w $(ALL_SRC)

clean:
	rm -rf cadence-client
	rm -Rf $(BUILD)
	rm -f dummy
