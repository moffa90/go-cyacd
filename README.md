# go-cyacd

[![Release](https://img.shields.io/github/v/release/moffa90/go-cyacd)](https://github.com/moffa90/go-cyacd/releases)
[![Go Reference](https://pkg.go.dev/badge/github.com/moffa90/go-cyacd.svg)](https://pkg.go.dev/github.com/moffa90/go-cyacd)
[![Go Report Card](https://goreportcard.com/badge/github.com/moffa90/go-cyacd)](https://goreportcard.com/report/github.com/moffa90/go-cyacd)
[![CI](https://github.com/moffa90/go-cyacd/workflows/CI/badge.svg)](https://github.com/moffa90/go-cyacd/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/moffa90/go-cyacd)](https://github.com/moffa90/go-cyacd)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Professional Go library for programming Cypress/Infineon microcontrollers with .cyacd firmware files.

## Features

- 🎯 **Pure Go** - No CGo dependencies
- 📦 **Zero external dependencies** - Only Go standard library
- 🔌 **Hardware independent** - Clean `io.ReadWriter` interface
- ⚡ **Full protocol support** - Complete Infineon bootloader protocol v1.60
- 🔄 **Progress tracking** - Real-time programming progress callbacks
- 📝 **Comprehensive logging** - Pluggable logging interface
- ⏱️ **Context support** - Cancellation and timeout support
- ✅ **Production tested** - Extensively validated with real firmware
- 🧪 **Mock device included** - Test without hardware
- 📚 **Documented** - Comprehensive godoc and examples

## Installation

```bash
go get github.com/moffa90/go-cyacd
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/moffa90/go-cyacd/bootloader"
    "github.com/moffa90/go-cyacd/cyacd"
)

func main() {
    // 1. User provides hardware communication (io.ReadWriter)
    device := openYourDevice() // Your USB/UART/etc implementation

    // 2. Parse firmware file
    fw, err := cyacd.Parse("firmware.cyacd")
    if err != nil {
        log.Fatal(err)
    }

    // 3. Create programmer with options
    prog := bootloader.New(device,
        bootloader.WithProgressCallback(func(p bootloader.Progress) {
            fmt.Printf("[%s] %.1f%% - Row %d/%d\n",
                p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
        }),
    )

    // 4. Program the device
    key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F} // Your bootloader key
    err = prog.Program(context.Background(), fw, key)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Programming successful!")
}
```

## Hardware Implementation

This library does **NOT** implement hardware communication. You provide an `io.ReadWriter`:

```go
type YourDevice struct {
    // Your hardware-specific fields
}

func (d *YourDevice) Read(p []byte) (int, error) {
    // Implement reading from your device
    // (USB, UART, SPI, I2C, network, etc.)
}

func (d *YourDevice) Write(p []byte) (int, error) {
    // Implement writing to your device
}
```

This design allows the library to work with **any** communication method.

## Advanced Usage

### Progress Tracking

```go
prog := bootloader.New(device,
    bootloader.WithProgressCallback(func(p bootloader.Progress) {
        fmt.Printf("Phase: %s\n", p.Phase) // entering, programming, verifying, exiting, complete
        fmt.Printf("Progress: %.1f%%\n", p.Percentage)
        fmt.Printf("Rows: %d/%d\n", p.CurrentRow, p.TotalRows)
        fmt.Printf("Bytes: %d\n", p.BytesWritten)
        fmt.Printf("Elapsed: %s\n", p.ElapsedTime)
    }),
)
```

### Custom Logging

```go
type MyLogger struct {
    logger *log.Logger
}

func (l *MyLogger) Debug(msg string, kv ...interface{}) {
    l.logger.Println("DEBUG:", msg, kv)
}

func (l *MyLogger) Info(msg string, kv ...interface{}) {
    l.logger.Println("INFO:", msg, kv)
}

func (l *MyLogger) Error(msg string, kv ...interface{}) {
    l.logger.Println("ERROR:", msg, kv)
}

prog := bootloader.New(device, bootloader.WithLogger(&MyLogger{...}))
```

### Configuration Options

```go
prog := bootloader.New(device,
    // Progress tracking
    bootloader.WithProgressCallback(progressFunc),

    // Logging
    bootloader.WithLogger(myLogger),

    // Timeouts
    bootloader.WithTimeout(30*time.Second),
    bootloader.WithReadTimeout(10*time.Second),
    bootloader.WithWriteTimeout(10*time.Second),

    // Data chunking
    bootloader.WithChunkSize(64), // Default: 57 bytes

    // Retry logic
    bootloader.WithRetries(5), // Default: 3

    // Verification
    bootloader.WithVerifyAfterProgram(true), // Default: true
)
```

### Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

err := prog.Program(ctx, fw, key)
if err == context.DeadlineExceeded {
    fmt.Println("Programming timed out")
}
```

## Package Structure

```
go-cyacd/
├── cyacd/          # CYACD file parser
│   ├── Parse()           # Parse .cyacd file
│   └── ParseReader()     # Parse from io.Reader
├── protocol/       # Bootloader protocol implementation
│   ├── Build*Cmd()       # Command frame builders
│   └── Parse*Response()  # Response parsers
└── bootloader/     # High-level programmer API
    ├── New()             # Create programmer
    └── Program()         # Program firmware
```

## .CYACD File Format

The library parses Cypress .cyacd firmware files:

**Header** (12 hex characters):
```
[SiliconID(8)][SiliconRev(2)][ChecksumType(2)]
```

**Row** (variable length):
```
[ArrayID(2)][RowNum(4)][DataLen(4)][Data(N)][Checksum(2)]
```

Example:
```
1E9602AA00                    # Header
000000040401020304F6          # Row 0
0001000404050607080A          # Row 1
...
```

## Error Handling

The library provides structured error types:

```go
err := prog.Program(ctx, fw, key)

switch e := err.(type) {
case *bootloader.DeviceMismatchError:
    fmt.Printf("Wrong device: expected 0x%08X, got 0x%08X\n", e.Expected, e.Actual)

case *bootloader.RowOutOfRangeError:
    fmt.Printf("Row %d out of range (%d-%d)\n", e.RowNum, e.MinRow, e.MaxRow)

case *bootloader.ChecksumMismatchError:
    fmt.Printf("Row %d checksum failed\n", e.RowNum)

case *bootloader.VerificationError:
    fmt.Println("Application verification failed")

case *protocol.ProtocolError:
    fmt.Printf("Bootloader error: %s (0x%02X)\n", e.Error(), e.StatusCode)
}
```

## Supported Commands

The library implements all Infineon bootloader commands:

- ✅ Enter Bootloader
- ✅ Exit Bootloader
- ✅ Program Row
- ✅ Erase Row
- ✅ Verify Row
- ✅ Verify Checksum
- ✅ Get Flash Size
- ✅ Send Data
- ✅ Sync Bootloader
- ✅ Get Metadata
- ✅ Get Application Status (multi-app)
- ✅ Set Active Application (multi-app)

## Examples

See the [examples](examples/) directory for complete working examples:

- [basic](examples/basic/) - Simple programming example
- [with_progress](examples/with_progress/) - Progress tracking
- [advanced](examples/advanced/) - Full-featured example
- [mock_device](examples/mock_device/) - Mock device for testing

## Documentation

- [GoDoc](https://pkg.go.dev/github.com/moffa90/go-cyacd)
- [Protocol Specification](docs/PROTOCOL.md)
- [Examples](examples/)
- [Contributing Guide](CONTRIBUTING.md)

## Requirements

- Go 1.21 or later
- No external dependencies

## API Stability

Starting with **v1.0.0**, this library follows strict semantic versioning and provides API stability guarantees:

### Stable API (v1.x.x)

All exported types, functions, constants, and methods in the public API are **stable** and will not change in backwards-incompatible ways within the v1 major version:

- **Packages**: `bootloader`, `cyacd`, `protocol`
- **Types**: All exported structs, interfaces, and type aliases
- **Functions**: All exported functions and methods
- **Constants**: All exported constants

### Compatibility Promise

- ✅ **Patch releases** (v1.0.x) - Bug fixes only, no API changes
- ✅ **Minor releases** (v1.x.0) - New features, backwards-compatible
- ❌ **Major releases** (v2.0.0) - Breaking changes allowed (with migration guide)

### What This Means

Your code written for v1.0.0 will continue to compile and work correctly with any v1.x.x release:

```go
// This code will work with v1.0.0 through v1.∞.∞
import "github.com/moffa90/go-cyacd/bootloader"

prog := bootloader.New(device)
err := prog.Program(ctx, fw, key)
```

### Deprecation Policy

If a feature needs to be removed:

1. It will be marked as **deprecated** with a clear alternative
2. It will remain functional for at least **one major version**
3. Deprecation will be documented in godoc and CHANGELOG
4. Migration guide will be provided before removal

### Pre-v1.0.0 Disclaimer

Versions before v1.0.0 (v0.x.x) are considered **development versions** and may include breaking changes in minor releases. Once v1.0.0 is released, the stability guarantees above apply.

### Version Selection

```bash
# Latest stable version (recommended for production)
go get github.com/moffa90/go-cyacd@latest

# Specific version
go get github.com/moffa90/go-cyacd@v1.2.3

# Latest v1.x (safe, backwards-compatible updates)
go get github.com/moffa90/go-cyacd@v1
```

## License

MIT License - see [LICENSE](LICENSE) file for details

## References

- Infineon Bootloader Protocol Specification v1.30
- PSoC Creator Component Datasheet: Bootloader and Bootloadable

## Support This Project

If you find this library useful, please consider supporting its development:

[![PayPal](https://img.shields.io/badge/Donate-PayPal-blue.svg)](https://paypal.me/moffax)

Your support helps maintain and improve this project. Thank you!

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
