.PHONY: test bins clean cover cover_ci
PROJECT_ROOT = github.com/uber-go/cadence-client

export PATH := $(GOPATH)/bin:$(PATH)

THRIFT_GENDIR=.gen

# default target
default: test

# define the list of thrift files the service depends on
# (if you have some)
THRIFT_SRCS = idl/github.com/uber/cadence/cadence.thrift \
	idl/github.com/uber/cadence/shared.thrift \

PROGS = cadence-client
TEST_ARG ?= -race -v -timeout 5m
BUILD := ./build

LIBRARY_VERSION=v0.1.0
GIT_SHA=`git rev-parse HEAD`
OUT_VERSION_FILE=./client/cadence/version.go

export PATH := $(GOPATH)/bin:$(PATH)

THRIFT_GEN=$(GOPATH)/bin/thrift-gen


define thriftrule
THRIFT_GEN_SRC += $(THRIFT_GENDIR)/go/$1/tchan-$1.go

$(THRIFT_GENDIR)/go/$1/tchan-$1.go:: $2 $(THRIFT_GEN)
	@mkdir -p $(THRIFT_GENDIR)/go
	$(ECHO_V)$(THRIFT_GEN) --generateThrift --packagePrefix $(PROJECT_ROOT)/$(THRIFT_GENDIR)/go/ --inputFile $2 --outputDir $(THRIFT_GENDIR)/go \
		$(foreach template,$(THRIFT_TEMPLATES), --template $(template))
endef

$(foreach tsrc,$(THRIFT_SRCS),$(eval $(call \
	thriftrule,$(basename $(notdir \
	$(shell echo $(tsrc) | tr A-Z a-z))),$(tsrc))))

# Automatically gather all srcs
ALL_SRC := $(shell find . -name "*.go" | grep -v -e Godeps -e vendor \
	-e ".*/\..*" \
	-e ".*/_.*" \
	-e ".*/mocks.*")

# all directories with *_test.go files in them
TEST_DIRS := $(sort $(dir $(filter %_test.go,$(ALL_SRC))))

glide:
	glide install

clean_thrift:
	rm -rf .gen

thriftc: clean_thrift glide $(THRIFT_GEN_SRC)

libversion_gen: ./cmd/tools/libversiongen.go
    # auto-generate const for libversion and git-sha
	go run ./cmd/tools/libversiongen.go -v=$(LIBRARY_VERSION) -s=$(GIT_SHA) -o=$(OUT_VERSION_FILE)

bins: libversion_gen thriftc
	go build -i -o cadence-client main.go

bins_nothrift: libversion_gen glide
	go build -i -o cadence-client main.go

test: bins
	@rm -f test
	@rm -f test.log
	@for dir in $(TEST_DIRS); do \
		go test -race -coverprofile=$@ "$$dir" | tee -a test.log; \
	done;

cover_profile: clean bins_nothrift
	@echo Testing packages:
	@for dir in $(TEST_DIRS); do \
		mkdir -p $(BUILD)/"$$dir"; \
		go test "$$dir" $(TEST_ARG) -coverprofile=$(BUILD)/"$$dir"/coverage.out || exit 1; \
	done;

cover: cover_profile
	@for dir in $(TEST_DIRS); do \
		go tool cover -html=$(BUILD)/"$$dir"/coverage.out; \
	done

cover_ci: cover_profile
	@for dir in $(TEST_DIRS); do \
		goveralls -coverprofile=$(BUILD)/"$$dir"/coverage.out -service=travis-ci || echo -e "\x1b[31mCoveralls failed\x1b[m"; \
	done

clean:
	rm -rf cadence-client
