# Contributing to go-cyacd

Thank you for your interest in contributing to go-cyacd! This document provides guidelines and instructions for contributing to the project.

## Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## Ways to Contribute

- **Report bugs** - File detailed bug reports with reproduction steps
- **Suggest features** - Propose new features or enhancements
- **Improve documentation** - Fix typos, clarify instructions, add examples
- **Submit code** - Fix bugs, implement features, improve performance
- **Review pull requests** - Help review and test proposed changes

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git
- A text editor or IDE with Go support

### Getting Started

1. **Fork the repository** on GitHub

2. **Clone your fork**:
   ```bash
   git clone https://github.com/YOUR_USERNAME/go-cyacd.git
   cd go-cyacd
   ```

3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/moffa90/go-cyacd.git
   ```

4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

### Running Tests

Run all tests:
```bash
go test ./...
```

Run tests with race detector:
```bash
go test -race ./...
```

Run tests with coverage:
```bash
go test -cover ./...
```

Run benchmarks:
```bash
go test -bench=. ./...
```

### Building Examples

```bash
go build ./examples/basic
go build ./examples/advanced
go build ./examples/with_progress
```

## Code Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format your code (run `go fmt ./...`)
- Use `go vet` to catch common issues (run `go vet ./...`)
- Run `golangci-lint` if available for additional checks
- Keep lines under 120 characters when practical

### Documentation

- **Add godoc comments** for all exported types, functions, and constants
- **Start comments with the name** of the thing being documented
- **Use complete sentences** with proper punctuation
- **Include examples** for non-trivial functionality
- **Document error returns** and special conditions

Example:
```go
// ParseFirmware parses a .cyacd firmware file and returns a Firmware struct.
// It validates the file format, checksums, and ensures all rows are properly formed.
//
// Returns an error if:
//   - The file cannot be read
//   - The file format is invalid
//   - Checksums do not match
//   - Required fields are missing
func ParseFirmware(path string) (*Firmware, error) {
    // ...
}
```

### Testing

- **Write tests** for all new functionality
- **Maintain test coverage** - aim for 80%+ coverage
- **Test edge cases** - null inputs, empty data, boundary conditions
- **Test error conditions** - ensure errors are properly handled
- **Use table-driven tests** for multiple test cases
- **Add benchmarks** for performance-critical code

Example table-driven test:
```go
func TestParseRow(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    *Row
        wantErr bool
    }{
        {
            name:  "valid row",
            input: "000000040001020304F2",
            want:  &Row{...},
            wantErr: false,
        },
        // More test cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := parseRow(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("parseRow() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            // Assert expectations...
        })
    }
}
```

### Error Handling

- **Return errors, don't panic** - except in truly exceptional cases
- **Wrap errors with context** using `fmt.Errorf("context: %w", err)`
- **Define custom error types** for errors that callers might handle
- **Make error messages actionable** - explain what went wrong and why

### Commit Messages

Write clear, descriptive commit messages:

```
Short summary (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain the problem this commit solves and why you chose
this solution.

- Bullet points are fine
- Use present tense ("Add feature" not "Added feature")
- Reference issues: Fixes #123
```

Examples:
- `Add support for CRC16 checksums`
- `Fix race condition in programmer progress reporting`
- `Refactor parser to reduce memory allocations`
- `Update documentation for hybrid CYACD format`

## Pull Request Process

### Before Submitting

1. **Update to latest main**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all tests**:
   ```bash
   go test ./...
   go test -race ./...
   ```

3. **Format and vet code**:
   ```bash
   go fmt ./...
   go vet ./...
   ```

4. **Update documentation** if needed:
   - Update README.md for new features
   - Update godoc comments
   - Add examples if appropriate

5. **Update CHANGELOG.md** with your changes

### Submitting the PR

1. **Push your branch** to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create a Pull Request** on GitHub from your branch to `main`

3. **Fill out the PR template** completely:
   - Describe what the PR does
   - Reference any related issues
   - List any breaking changes
   - Include test results

4. **Respond to feedback** - be open to suggestions and questions

### PR Requirements

- ✅ All tests pass
- ✅ No race conditions (verified with `-race`)
- ✅ Code is formatted (`go fmt`)
- ✅ No vet warnings (`go vet`)
- ✅ Documentation is updated
- ✅ CHANGELOG.md is updated
- ✅ Commit messages are clear
- ✅ No merge conflicts with main

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. **Go version**: Output of `go version`
2. **Library version**: Which version or commit
3. **Operating system**: OS and version
4. **Description**: What you expected vs. what happened
5. **Reproduction steps**: Minimal code to reproduce the issue
6. **Error messages**: Full error output
7. **Additional context**: Logs, firmware files (if not confidential)

### Feature Requests

When requesting features:

1. **Use case**: Describe the problem you're trying to solve
2. **Proposed solution**: How you envision the feature working
3. **Alternatives**: Other approaches you've considered
4. **Breaking changes**: Would this affect existing users?

## Project Structure

```
go-cyacd/
├── bootloader/       # Bootloader programmer implementation
├── cyacd/           # .cyacd file parser
├── protocol/        # Low-level protocol implementation
├── examples/        # Example programs
│   ├── basic/
│   ├── advanced/
│   ├── with_progress/
│   └── mock_device/
├── docs/            # Additional documentation
├── CHANGELOG.md     # Version history
├── README.md        # Project overview
└── CONTRIBUTING.md  # This file
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **MAJOR** version for incompatible API changes
- **MINOR** version for backwards-compatible functionality
- **PATCH** version for backwards-compatible bug fixes

After v1.0.0, the API is stable and breaking changes will be avoided.

## API Stability Guarantee

Starting with v1.0.0, the public API is stable:

- Exported types, functions, and methods will not change incompatibly
- New features may be added in minor versions
- Deprecations will be clearly documented and maintained for at least one major version
- See the [API stability policy](README.md#api-stability) for details

## Getting Help

- **Documentation**: See [pkg.go.dev](https://pkg.go.dev/github.com/moffa90/go-cyacd)
- **Examples**: Check the `examples/` directory
- **Issues**: Search existing [GitHub issues](https://github.com/moffa90/go-cyacd/issues)
- **Discussions**: Start a [GitHub discussion](https://github.com/moffa90/go-cyacd/discussions) for questions

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Recognition

Contributors are recognized in:
- Git commit history
- Release notes
- Project README (for significant contributions)

Thank you for contributing to go-cyacd! 🎉
