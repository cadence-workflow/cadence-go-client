#!/bin/sh

# intentionally NOT `set -e` to simplify checks before exiting,
# make changes with care

# refresh deps
make tidy || exit 1
# regenerate, format, and make sure everything builds
make build || exit 1

# intentionally capture stderr, so status-errors are also PR-failing.
# in particular this catches "dubious ownership" failures, which otherwise
# do not fail this check and the $() hides the exit code.
if [ -n "$(git status --porcelain  2>&1)" ]; then
  echo "There file changes after applying your diff and performing a build."
  echo "Please run this command and commit the changes:"
  echo "\tmake tidy && make build"
  git status --porcelain
  git --no-pager diff
  exit 1
fi
