#!/bin/bash
set -e

# Run go test with coverage profile
go test -v -coverprofile=coverage.out ./...

# Check total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+')

echo "Total coverage: ${TOTAL_COVERAGE}%"

# Check if coverage is below 90%
# Using bc for floating point comparison if available, otherwise integer comparison
if (( $(echo "$TOTAL_COVERAGE < 85" | bc -l) )); then
    echo "Error: Total coverage is below 90%"
    exit 1
fi

echo "Coverage check passed!"
