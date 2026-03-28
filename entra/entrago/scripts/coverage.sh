#!/bin/bash
set -euo pipefail

MIN_COVERAGE=85

mkdir -p .tmp
COVERAGE_OUT=.tmp/coverage.out

# Run go test with coverage profile across all packages
go test -v -coverpkg=./... -coverprofile="${COVERAGE_OUT}" ./...

# Check total coverage
TOTAL_COVERAGE=$(go tool cover -func="${COVERAGE_OUT}" | grep total | grep -Eo '[0-9]+\.[0-9]+')

echo "Total coverage: ${TOTAL_COVERAGE}%"

# Check if coverage is below threshold
if awk "BEGIN {exit !(${TOTAL_COVERAGE} < ${MIN_COVERAGE})}"; then
    echo "Error: Total coverage is below ${MIN_COVERAGE}%"
    exit 1
fi

echo "Coverage check passed!"
