# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Metadata

- **Maintainer**: Jose Moffa (moffa3e@gmail.com)
- **License**: MIT
- **Go Version**: 1.21+
- **Dependencies**: Zero external dependencies (only Go standard library)
- **Repository**: https://github.com/moffa90/go-cyacd

## Overview

go-cyacd is a **professional, production-ready Go library** for programming Cypress/Infineon microcontrollers with .cyacd firmware files. It implements the complete **Infineon bootloader protocol v1.30** with zero external dependencies (only Go standard library).

**Key Design Principles**:
1. **Hardware-agnostic**: Users provide an `io.ReadWriter` interface for device communication (USB, UART, SPI, I2C, network, etc.), and the library handles all protocol logic
2. **No magic numbers or strings**: All constants are defined in `protocol/constants.go`
3. **Professional code quality**: Follows Go best practices, comprehensive documentation, structured error handling
4. **Production-tested**: Extensively validated with real Cypress/Infineon firmware

## Development Commands

### Testing
```bash
# Run all tests with race detection
make test
# or directly:
go test -v -race ./...

# Run tests for specific package
go test -v ./protocol
go test -v ./cyacd
go test -v ./bootloader
```

### Code Quality
```bash
# Format code
make fmt
# or directly:
gofmt -s -w .

# Run go vet
make vet
# or directly:
go vet ./...

# Run linters (requires golangci-lint)
make lint
# or directly:
golangci-lint run
```

### Building Examples
```bash
# Build all examples
make examples
# or directly:
go build -o bin/basic ./examples/basic
```

## Architecture

The library is organized into three clean layers that implement the bootloader protocol:

### 1. cyacd Package - Firmware File Parser
**Path**: `cyacd/`

Parses .cyacd firmware files into structured data:
- `Parse()` - Parse from file path
- `ParseReader()` - Parse from io.Reader

**File Format**:
- Header (12 hex chars): `[SiliconID(8)][SiliconRev(2)][ChecksumType(2)]`
- Rows: `[ArrayID(2)][RowNum(4)][DataLen(4)][Data(N)][Checksum(2)]`
- Also supports Intel HEX hybrid format (rows starting with `:`)

Key types:
- `Firmware` - Contains header metadata and array of rows
- `Row` - Single flash row with ArrayID, RowNum, Data, Checksum

### 2. protocol Package - Low-Level Protocol Implementation
**Path**: `protocol/`

Implements the raw Infineon bootloader protocol frames and commands:

**Command Builders** (`commands.go`):
- `BuildEnterBootloaderCmd()` - Enter bootloader with 6-byte key
- `BuildProgramRowCmd()` - Program flash row
- `BuildVerifyRowCmd()` - Verify row checksum
- `BuildVerifyChecksumCmd()` - Verify application checksum
- `BuildGetFlashSizeCmd()` - Query flash array range
- `BuildExitBootloaderCmd()` - Exit and reset device
- Plus: SendData, EraseRow, SyncBootloader, GetMetadata, GetAppStatus, SetActiveApp

**Response Parsers** (`responses.go`):
- `ParseResponse()` - Extract status code and data from frame
- `ParseEnterBootloaderResponse()` - Device identification info
- `ParseVerifyRowResponse()` - Row checksum
- `ParseGetFlashSizeResponse()` - Flash row range
- etc.

**Frame Format**:
```
[SOP(0x01)][CMD/STATUS][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP(0x17)]
```

**Checksum Calculation** (`checksum.go`):
- `calculatePacketChecksum()` - Frame checksum (includes SOP through DATA)
- `CalculateRowChecksumWithMetadata()` - Row verification checksum with metadata

All protocol constants are defined in `constants.go` (commands, status codes, frame markers, sizes).

### 3. bootloader Package - High-Level Programmer API
**Path**: `bootloader/`

Orchestrates the complete programming sequence:

**Main Type**: `Programmer` - Created via `New(device io.ReadWriter, opts ...Option)`

**Programming Flow** (`Program()` method):
1. Enter bootloader with key
2. Validate device silicon ID matches firmware
3. Get flash size and validate all rows are in range
4. Program all rows (with optional chunking via SendData)
5. Verify each row after programming (if enabled)
6. Verify application checksum
7. Exit bootloader

**Configuration Options** (`options.go`):
- `WithProgressCallback()` - Real-time progress tracking (phase, percentage, rows, bytes, time)
- `WithLogger()` - Pluggable logging interface
- `WithTimeout()` / `WithReadTimeout()` / `WithWriteTimeout()` - Timeouts
- `WithChunkSize()` - Data chunk size (default 57 bytes to match reference bootloader-usb implementation)
- `WithRetries()` - Retry logic for failed commands
- `WithVerifyAfterProgram()` - Enable/disable row verification (default: true)
- `WithCommandDelay()` - Inter-command delay (25ms for Serial, 1ms for USB)
- `WithLenientVerifyRow()` - Accept non-standard VerifyRow responses

**Progress Phases**: entering, programming, verifying, exiting, complete

**HID Device Handling**: The `sendCommandWithResponse()` method automatically detects and strips HID Report ID bytes (often 0x00) that some USB devices prepend to responses.

## Critical Implementation Details

### Checksum Calculation
**IMPORTANT**: The packet checksum includes the SOP byte through all DATA bytes (see `protocol/checksum.go`). This was fixed in v0.5.0 and is critical for proper communication.

Row verification uses a composite checksum that includes both the row data checksum AND metadata (ArrayID, RowNum, DataLen).

### Data Chunking Strategy

**CRITICAL**: The chunk size must account for command frame overhead to fit within USB HID packet limits (typically 64 bytes).

**Frame Overhead Calculations**:
- **ProgramRow command**: 10 bytes overhead
  - SOP(1) + CMD(1) + LEN(2) + ArrayID(1) + RowNum(2) + CHECKSUM(2) + EOP(1) = 10 bytes
  - Maximum safe data: 64 - 10 = **54 bytes**
- **SendData command**: 7 bytes overhead
  - SOP(1) + CMD(1) + LEN(2) + CHECKSUM(2) + EOP(1) = 7 bytes
  - Maximum safe data: 64 - 7 = **57 bytes**

**Default Configuration**:
- `bootloader/options.go`: `DefaultChunkSize = 57` (matches reference bootloader-usb)
- `protocol/constants.go`: Provides `MaxPacketSize = 64`, `SendDataOverhead = 7`, `ProgramRowOverhead = 10`

**Chunking Algorithm** (in `programRow()`):
Replicates reference bootloader-usb algorithm: `(r.Size()-offset+7) > PacketSize`
1. While (remaining data + SendData overhead) > 64 bytes: send 57-byte chunks using `SendData`
2. Send final chunk (remaining data ≤ 57 bytes) with `ProgramRow` command

**Example**: Row with 200 bytes
- SendData(57 bytes) × 3
- ProgramRow(29 bytes)

**Historical Bugs**:
- v0.5.0: `DefaultChunkSize = 64` caused ERR_LENGTH errors
- v0.5.1: Changed to 54 (too conservative)
- v0.5.2: Changed to 57 to match reference bootloader-usb implementation ✓

### Context Support
All operations support `context.Context` for cancellation and timeouts. Always check `ctx.Err()` in loops.

### Error Handling
The library provides structured error types:
- `protocol.ProtocolError` - Bootloader error responses (includes status code)
- `bootloader.DeviceMismatchError` - Silicon ID mismatch
- `bootloader.RowOutOfRangeError` - Invalid row number
- `bootloader.ChecksumMismatchError` - Verification failure
- `bootloader.VerificationError` - Application validation failure

Use type assertions or `errors.As()` to handle specific errors.

### Testing Philosophy
- Tests use `_test.go` files in the same package
- Test data and fixtures are typically in `testdata/` directories
- The library includes a mock device for testing without hardware
- Test configuration is in `.golangci.yml` with relaxed rules for test files

## Code Style and Standards

### Best Practices
- **No magic numbers or strings**: ALL constants must be defined in appropriate `constants.go` files
  - Protocol constants: `protocol/constants.go`
  - Parser constants: `cyacd/parser.go`
  - Configuration defaults: `bootloader/options.go`
- **Professional library standards**: This is a production library, not a script
  - Comprehensive godoc comments on all exported types/functions
  - Structured error types with context
  - Defensive programming with validation
  - Clear separation of concerns

### Go Code Standards
- **Line length**: Max 140 characters (`.golangci.yml` lll linter)
- **Function length**: Max 100 lines / 50 statements (funlen linter, relaxed for tests)
- **Comments**: All exported types and functions must have godoc comments starting with the type/function name
- **Error handling**: Always wrap errors with context using `fmt.Errorf("operation: %w", err)`
- **Naming**: Follow standard Go conventions; use descriptive names (avoid abbreviations unless standard)
- **Imports**: Group imports (stdlib, external, internal) with blank lines between groups

### Git Commit Standards
- **Author**: All commits must be authored by Jose Moffa (moffa3e@gmail.com)
- **Messages**: Use conventional commit format
  - `feat:` for new features
  - `fix:` for bug fixes
  - `docs:` for documentation
  - `refactor:` for code refactoring
  - `test:` for test additions/changes
- **Scope**: Be specific and concise

## Package Dependencies

- **Zero external dependencies** - Only Go standard library
- **Minimum Go version**: 1.21 (specified in `go.mod`)
- No CGo, no external C libraries

## Important Files

### Core Implementation
- `protocol/constants.go` - All protocol constants, command codes, status codes, frame markers, sizes
- `bootloader/programmer.go` - Main programming orchestration logic
- `protocol/checksum.go` - Checksum algorithms (CRITICAL: includes SOP byte, fixed in v0.5.0)
- `cyacd/parser.go` - Firmware file parsing with format validation
- `bootloader/options.go` - Configuration options and defaults (includes DefaultChunkSize)

### Key Configuration Values
- **Chunk Sizes**:
  - `bootloader/options.go`: `DefaultChunkSize = 54` (for ProgramRow with overhead)
  - `protocol/constants.go`: `DefaultChunkSize = 57` (for SendData only)
- **Timeouts**:
  - `DefaultReadTimeout = 5 * time.Second`
  - `DefaultWriteTimeout = 5 * time.Second`
- **Retries**: `DefaultRetries = 3`
- **Verification**: `VerifyAfterProgram = true` (default enabled)

## Common Development Patterns

### Adding New Protocol Commands
1. Add command constant to `protocol/constants.go`
2. Add builder function to `protocol/commands.go`
3. Add response parser to `protocol/responses.go`
4. Add corresponding method to `bootloader/programmer.go` if needed
5. Add tests to `protocol/commands_test.go` and `protocol/responses_test.go`

### Adding Configuration Options
1. Add field to `Config` struct in `bootloader/options.go`
2. Add default value in `defaultConfig()`
3. Create `WithXxx()` option function with validation
4. Document usage in function godoc

### Testing with Mock Devices
See `examples/mock_device/` for how to implement a mock `io.ReadWriter` for testing without hardware.

## Hardware Integration

### Interface Design
The library is **completely hardware-agnostic**. Users implement the `io.ReadWriter` interface:

```go
type io.ReadWriter interface {
    Read(p []byte) (n int, err error)
    Write(p []byte) (n int, err error)
}
```

### Supported Transports
Any transport can be used as long as it implements `io.ReadWriter`:
- **USB HID**: Most common for Cypress/Infineon bootloaders
- **UART/Serial**: Common for embedded systems
- **SPI**: High-speed serial peripheral interface
- **I2C**: Inter-integrated circuit bus
- **Network**: TCP/UDP for remote programming
- **Custom**: Any byte stream transport

### USB HID Considerations
- **Packet Size**: Typically 64 bytes (accounted for in DefaultChunkSize)
- **Report ID**: Some devices prepend a Report ID byte (0x00), which is automatically detected and stripped by `sendCommandWithResponse()`
- **Command Delay**: USB/HID typically works with 1ms delay or no delay
  ```go
  bootloader.WithCommandDelay(1 * time.Millisecond)
  ```

### Serial/UART Considerations
- **Command Delay**: Serial typically needs 25ms delay between commands
  ```go
  bootloader.WithCommandDelay(25 * time.Millisecond)
  ```
- **Baud Rate**: Configured in your device implementation
- **Flow Control**: Handle in your device implementation

## Protocol Details

### Frame Structure
All communication uses this frame format:
```
[SOP][CMD/STATUS][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
 0x01  1 byte     2 bytes LE     N bytes   2 bytes LE              0x17
```

### Status Codes (from `protocol/constants.go`)
- `0x00` StatusSuccess - Command executed successfully
- `0x03` ErrLength - Data amount outside expected range
- `0x04` ErrData - Data not of proper form
- `0x05` ErrCommand - Command not recognized
- `0x06` ErrKey - Invalid bootloader key
- `0x08` ErrChecksum - Checksum mismatch
- `0x09` ErrArray - Invalid flash array ID
- `0x0A` ErrRow - Invalid flash row number
- `0x0C` ErrApp - Application not valid
- `0x0D` ErrActive - Application currently marked as active
- `0x0F` ErrUnknown - Unknown error

### Checksum Types
- `0x00` ChecksumBasicSum - Basic summation (2's complement)
- `0x01` ChecksumCRC16 - CRC-16-CCITT

## References

### Official Documentation
- **Infineon Bootloader Protocol Specification v1.30**
- **PSoC Creator Component Datasheet: Bootloader and Bootloadable v1.60**
- **Application Note AN60317**: Describes the bootloader protocol

### Related Projects
- **Original C Implementation**: https://github.com/Cellgain/bootloader-usb
  - Reference implementation this library was ported from
  - Contains auto-calculation of remaining chunks

### Package Documentation
- **GoDoc**: https://pkg.go.dev/github.com/moffa90/go-cyacd
- **Examples**: See `examples/` directory for working code samples

## Troubleshooting

### Common Errors

#### ERR_LENGTH (0x03)
**Symptom**: Error when programming certain rows
**Cause**: Frame size exceeds USB HID packet limit (64 bytes)
**Solution**: Ensure `DefaultChunkSize = 54` in `bootloader/options.go`
**Details**: ProgramRow has 10 bytes overhead, so max data is 54 bytes

#### ERR_CHECKSUM (0x08)
**Symptom**: Checksum validation fails
**Cause**: Incorrect checksum calculation
**Solution**: Ensure checksum includes SOP byte through DATA bytes (fixed in v0.5.0)

#### ERR_KEY (0x06)
**Symptom**: Cannot enter bootloader
**Cause**: Incorrect 6-byte bootloader key
**Solution**: Verify key matches device configuration

#### Device Mismatch
**Symptom**: Silicon ID mismatch error
**Cause**: Firmware file doesn't match target device
**Solution**: Use correct .cyacd file for your device

### Debug Tips
1. Enable logging with `WithLogger()` to see protocol-level details
2. Use progress callback with `WithProgressCallback()` to track operation
3. Check device communication with simple read/write tests first
4. Verify USB HID packet sizes match expectations (typically 64 bytes)
5. Test with mock device (`examples/mock_device/`) to isolate hardware issues

## Version History

### v0.5.1 (Current)
- **Fix**: Changed DefaultChunkSize from 64 to 54 bytes to fix ERR_LENGTH errors
- **Feature**: Added optional lenient mode for VerifyRow responses

### v0.5.0
- **Fix**: Critical checksum calculation bug - now includes SOP byte per spec
- **Feature**: Added command delay support for Serial/UART devices

### Earlier Versions
- See git history and CHANGELOG for complete version history

## Quick Reference

### Library Features
- 🎯 **Pure Go** - No CGo dependencies
- 📦 **Zero external dependencies** - Only Go standard library
- 🔌 **Hardware independent** - Clean `io.ReadWriter` interface
- ⚡ **Full protocol support** - Complete Infineon bootloader protocol v1.30
- 🔄 **Progress tracking** - Real-time programming progress callbacks
- 📝 **Comprehensive logging** - Pluggable logging interface
- ⏱️ **Context support** - Cancellation and timeout support
- ✅ **Production tested** - Extensively validated with real firmware
- 🧪 **Mock device included** - Test without hardware
- 📚 **Documented** - Comprehensive godoc and examples

### Supported Commands (All Infineon Protocol v1.30)
- ✅ Enter Bootloader (0x38)
- ✅ Exit Bootloader (0x3B)
- ✅ Program Row (0x39)
- ✅ Erase Row (0x34)
- ✅ Verify Row (0x3A)
- ✅ Verify Checksum (0x31)
- ✅ Get Flash Size (0x32)
- ✅ Send Data (0x37)
- ✅ Sync Bootloader (0x35)
- ✅ Get Metadata (0x3C)
- ✅ Get Application Status (0x33) - Multi-app only
- ✅ Set Active Application (0x36) - Multi-app only

### File Format Support
- ✅ Standard .cyacd format (Cypress Application Code and Data)
- ✅ Intel HEX hybrid format (rows starting with ':')
- ✅ Checksum validation (Basic Sum and CRC-16-CCITT)
- ✅ Multi-array flash support

### Programming Sequence
1. **Enter Bootloader** - Authenticate with 6-byte key
2. **Validate Device** - Check silicon ID matches firmware
3. **Get Flash Size** - Query valid row range
4. **Program Rows** - Write firmware data with automatic chunking
5. **Verify Rows** - Optional per-row checksum verification
6. **Verify Application** - Final application checksum check
7. **Exit Bootloader** - Reset device and launch application

### Configuration Quick Reference
```go
prog := bootloader.New(device,
    // Progress tracking
    bootloader.WithProgressCallback(func(p bootloader.Progress) {
        // p.Phase: entering, programming, verifying, exiting, complete
        // p.Percentage, p.CurrentRow, p.TotalRows, p.BytesWritten, p.ElapsedTime
    }),

    // Logging
    bootloader.WithLogger(myLogger), // Implements Logger interface

    // Timeouts
    bootloader.WithTimeout(30*time.Second),          // Both read and write
    bootloader.WithReadTimeout(10*time.Second),      // Read only
    bootloader.WithWriteTimeout(10*time.Second),     // Write only

    // Data transfer
    bootloader.WithChunkSize(54),                    // Default: 54 bytes

    // Reliability
    bootloader.WithRetries(5),                       // Default: 3
    bootloader.WithVerifyAfterProgram(true),         // Default: true

    // Transport-specific
    bootloader.WithCommandDelay(25*time.Millisecond), // Serial: 25ms, USB: 1ms or 0

    // Compatibility
    bootloader.WithLenientVerifyRow(),               // For non-standard bootloaders
)
```

## Summary for AI Assistants

When working with this codebase:

1. **Always use constants** - No magic numbers or strings
2. **Maintain hardware independence** - Never assume specific transport
3. **Professional quality** - This is production code, not a prototype
4. **Respect frame overhead** - ProgramRow = 10 bytes, SendData = 7 bytes
5. **Follow conventions** - Go best practices, clear error handling, comprehensive docs
6. **Test thoroughly** - Use mock devices, validate with real hardware
7. **Document everything** - Godoc comments, inline comments for complex logic
8. **Commit properly** - Author: Jose Moffa (moffa3e@gmail.com), conventional format

**Critical Knowledge**:
- DefaultChunkSize = 57 bytes to match reference bootloader-usb implementation
- Chunking algorithm: `(remaining + SendDataOverhead) > MaxPacketSize` ensures correct packet boundaries
- Checksum includes SOP byte through DATA bytes
- HID Report ID byte may need to be stripped from responses
- Row verification checksum includes metadata (ArrayID, RowNum, DataLen)
