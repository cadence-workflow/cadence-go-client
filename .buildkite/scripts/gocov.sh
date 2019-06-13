#!/bin/sh

set -ex

# fetch codecov reporting tool
go get github.com/mattn/goveralls

# download cover files from all the tests
mkdir -p .build/coverage
buildkite-agent artifact download ".build/coverage/unit_test_cover.out" . --step "unit-test" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download ".build/coverage/integ_test_sticky_off_cover.out" . --step "integration-test-sticky-off" --build "$BUILDKITE_BUILD_ID"
buildkite-agent artifact download ".build/coverage/integ_test_sticky_on_cover.out" . --step "integration-test-sticky-on" --build "$BUILDKITE_BUILD_ID"

echo "download complete"

CURRDIR=`pwd`
echo $CURRDIR
ls -al .build
ls -al .build/coverage

#export BUILDKITE_REPO="https://github.com/uber-go/cadence-client"
# report coverage
make cover_ci

# cleanup
rm -rf .build
