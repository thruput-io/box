**1. The "One True Command" for Coverage**
The most significant cause of agent confusion is looking at per-package coverage. In this project, handlers are often exercised by integration-style tests in other packages.
- **Action:** Always use the global coverage script: `bash scripts/coverage.sh`.
- **Why:** This uses `-coverpkg=./...`, which correctly attributes coverage across all packages (currently ~88.7% total).

**2. Fix the Base (Break the Failure Loop)**
The coverage script fails if *any* test fails. Agents get stuck trying to increase coverage while the baseline is broken.
- **Immediate Task:** Fix the two known failures:
    - `TestTokenHelpers_ResolveClientFromForm_WithSecretValidation`: Fix the test fixture (which currently uses an empty secret, bypassing validation logic).
JOHAN: Introduce two domain types ClientWithSecret and ClientWithoutSecret. Make both types implement the `domain.Client` interface. ClientWithSecret can have a method for validating it. Secret should be a non empty string.
  
  - `TestCorsMiddleware_AddsHeaders`: Fix the expectation (the middleware correctly short-circuits `OPTIONS` requests; the test should expect the inner handler *not* to be called).

**3. Coverage via Code Reduction**
Instead of adding more tests, increase coverage percentage by deleting dead branches.
- **Action:** Identify "impossible" error handling in handlers (e.g., checking if a `domain.TenantID` is valid after it has already been constructed).
- **Goal:** Trust the strict domain types and remove the redundant branches. This reduces the denominator and removes uncovered paths.

**4. Modularize the App Dependency**
To avoid "building the world" in every test:
- **Action:** Refactor handlers to accept specific domain interfaces or smaller configuration slices instead of the monolithic `*app.App`.
- **Benefit:** This allows unit tests to be small, focused, and free of complex setup boilerplate.

**5. Final Gatekeeper**
- **Process:** No contribution is valid unless `make all` passes. This ensures that new tests are not only functional but also compliant with all quality and linting standards.

By following this plan, we move from "chasing a percentage" to "improving the architecture," which naturally results in high, meaningful coverage.