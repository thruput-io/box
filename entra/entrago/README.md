
The Entra service provides a mock implementation of Entra ID (formerly Azure AD).

## Coverage Check

To run the coverage check and ensure it meets the minimum threshold of 80%, execute the following script from the `entra` directory:

```bash
./scripts/coverage.sh
```

The script performs the following actions:
1.  Runs `go test -v -coverprofile=coverage.out ./...` to generate a coverage profile.
2.  Uses `go tool cover -func=coverage.out` to calculate total coverage.
3.  Exits with a non-zero status code if the total coverage is below 80%.
