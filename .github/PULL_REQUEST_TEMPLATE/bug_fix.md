## Linked Issue

Closes #<!-- issue number -->

## Root Cause Analysis

**Category:** <!-- pick one: logic-error | missing-validation | race-condition | missing-edge-case | data-migration | config-error | dependency-change | performance | security | integration -->

**Detail:**
<!-- Explain what exactly went wrong technically. Example:
"The order total was recalculated after discount, but the tax was applied
to the pre-discount amount because service A cached the base price." -->

## Fix Summary

**Changed files:**
<!-- List key files and what was changed -->
- `main/app/service/xxx.go:123` — description

**How to verify:**
<!-- Steps for tester to confirm the fix -->
1. Repeat the reproduction steps from the issue
2. Verify that [expected behavior] now occurs
3. Also check [edge case]

## Test Coverage

- [ ] Added integration test: <!-- test file path or "N/A" -->
- [ ] Updated existing test: <!-- test name or "N/A" -->
- [ ] All tests pass (`cd main && go test ./...`)

## Prevention

<!-- What prevents this class of bug from recurring? -->
- [ ] Added validation
- [ ] Added test coverage
- [ ] Updated documentation
- [ ] Other: ___

## Checklist

- [ ] `go fmt ./...` passed
- [ ] `go vet ./...` passed
- [ ] Root cause analysis filled
- [ ] Test coverage section filled
- [ ] Issue linked with `Closes #N`
