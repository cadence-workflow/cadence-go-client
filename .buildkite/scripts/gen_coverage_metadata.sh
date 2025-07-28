#!/bin/sh

set -ex

# This script generates coverage metadata for the coverage report.
# Output is used by SonarQube integration in Uber and not used by OS repo coverage tool itself.

# Example output:
#   commit-sha: 6953daa563e8e44512bc349c9608484cfd4ec4ff
#   timestamp: 2024-03-04T19:29:16Z

# Required env variables:
# - BUILDKITE_BRANCH
# - BUILDKITE_COMMIT

output_path="$1"

if [ "$BUILDKITE_BRANCH" != "master" ] && [ "$BUILDKITE_BRANCH" != "origin/master" ]; then
  echo "Coverage metadata is only generated for master branch. Current branch: $BUILDKITE_BRANCH"
  exit 0
fi

if [ -z "$BUILDKITE_COMMIT" ]; then
  echo "BUILDKITE_COMMIT is not set"
  exit 1
fi

echo "commit-sha: $BUILDKITE_COMMIT" > "$output_path"
echo "timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)" >> "$output_path"

echo "Coverage metadata written to $output_path"

# TODO: delete this file after migration to github actions