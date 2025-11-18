# Pull Request

## Description

Provide a clear and concise description of what this PR does.

Fixes #(issue number)

## Type of Change

- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update
- [ ] Performance improvement
- [ ] Code refactoring
- [ ] Test improvement
- [ ] CI/CD improvement

## Changes Made

Describe the changes in detail:

-
-
-

## Testing

### Test Environment

- **Go version:** [e.g., 1.23.0]
- **OS:** [e.g., Ubuntu 22.04, Windows 11, macOS 14]
- **Device tested:** [e.g., PSoC 4, PSoC 5LP, mock device]

### Tests Performed

- [ ] All existing tests pass (`go test -v -race ./...`)
- [ ] Added new tests for new functionality
- [ ] Tested with real hardware
- [ ] Tested with mock device
- [ ] Examples still work correctly

### Test Results

```
Paste test output here
```

## Code Quality

- [ ] Code follows the project's style guidelines
- [ ] Code has been formatted with `gofmt`
- [ ] Code passes `go vet`
- [ ] Code passes `golangci-lint run`
- [ ] Self-review completed
- [ ] Comments added for complex logic
- [ ] Documentation updated (if applicable)

## API Changes

### Breaking Changes

- [ ] This PR introduces breaking changes
- [ ] Migration guide included in CHANGELOG.md
- [ ] All breaking changes documented

**If breaking changes, describe:**

```go
// Before
oldAPI()

// After
newAPI()
```

### New Public API

**If new exported functions/types added:**

```go
// Document the new API here
```

- [ ] New API has godoc comments
- [ ] Examples provided for new API
- [ ] API is consistent with existing patterns

## Backwards Compatibility

- [ ] This PR is fully backwards compatible with v1.x
- [ ] This PR requires a major version bump (v2.0.0)
- [ ] Deprecation warnings added for old API (if applicable)

## Documentation

- [ ] README.md updated (if applicable)
- [ ] CHANGELOG.md updated
- [ ] GoDoc comments added/updated
- [ ] Protocol documentation updated (if applicable)
- [ ] Examples updated (if applicable)

## Performance Impact

- [ ] No performance impact
- [ ] Performance improved
- [ ] Performance slightly degraded (justified why)
- [ ] Benchmarks added/updated

**If performance changes:**

```
Benchmark results:
```

## Security Considerations

- [ ] No security implications
- [ ] Security review completed
- [ ] SECURITY.md updated (if applicable)

**If security implications:**

[Describe security considerations]

## Related Issues

- Related to #
- Depends on #
- Blocks #

## Screenshots (if applicable)

Add screenshots or terminal output if it helps visualize the changes.

## Checklist

- [ ] I have read the [CONTRIBUTING.md](../CONTRIBUTING.md) guide
- [ ] My code follows the project's code style
- [ ] I have performed a self-review of my code
- [ ] I have commented complex or non-obvious code
- [ ] I have made corresponding changes to documentation
- [ ] My changes generate no new warnings
- [ ] I have added tests that prove my fix/feature works
- [ ] All new and existing tests pass locally
- [ ] I have updated CHANGELOG.md with my changes
- [ ] I have checked my code for security vulnerabilities
- [ ] I have verified backwards compatibility (for v1.x)

## Additional Notes

Add any additional notes, concerns, or questions for reviewers.
