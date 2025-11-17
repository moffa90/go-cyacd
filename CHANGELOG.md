# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.5.1] - 2025-01-17

### Added

- **bootloader/options.go**: New `WithLenientVerifyRow()` option for compatibility with legacy/non-standard bootloader firmware
- **protocol/responses.go**: Lenient mode in `ParseVerifyRowResponse()` accepts both 0-byte and 1-byte responses

### Changed

- **protocol/responses.go**: `ParseVerifyRowResponse()` now accepts a `lenient bool` parameter
  - When `lenient=false` (default): Strictly requires 1 byte per Infineon AN60317 v1.60 specification
  - When `lenient=true`: Accepts 0-byte (returns 0x00) or 1-byte (returns checksum) for compatibility
- **bootloader/programmer.go**: Updated to pass `LenientVerifyRow` config to response parser

### Fixed

- Support for devices with non-standard firmware that return 0-byte VerifyRow (0x3A) responses instead of the required 1-byte checksum

### Notes

**Default behavior unchanged:** The library remains strictly spec-compliant by default. Use `WithLenientVerifyRow()` option only if your device firmware returns 0-byte responses for command 0x3A.

Per Infineon AN60317 v1.60, command 0x3A (Get Row Checksum) should return exactly 1 byte containing the row checksum. Devices returning 0 bytes are non-compliant but may exist in legacy or custom implementations.

**Usage Example:**
```go
prog := bootloader.New(device,
    bootloader.WithLenientVerifyRow(),  // Enable for non-standard devices
)
```

---

## [0.5.0] - 2025-01-17

### BREAKING CHANGES

**Critical Checksum Calculation Fix**: The packet checksum calculation now correctly includes the SOP (Start of Packet, 0x01) byte as specified in the official Infineon bootloader protocol documentation (AN60317).

Previously, the library incorrectly excluded the SOP byte from checksum calculations, which caused all commands and responses to have incorrect checksums. This prevented successful communication with real Cypress/Infineon bootloader devices.

### Changed

- **protocol/checksum.go**: Updated `calculatePacketChecksum()` documentation to reflect that SOP is included in checksum calculation per Infineon AN60317 specification
- **protocol/commands.go**: All 12 command builder functions now include SOP byte in checksum calculation:
  - `BuildEnterBootloaderCmd`
  - `BuildGetFlashSizeCmd`
  - `BuildProgramRowCmd`
  - `BuildSendDataCmd`
  - `BuildVerifyRowCmd`
  - `BuildVerifyChecksumCmd`
  - `BuildEraseRowCmd`
  - `BuildSyncBootloaderCmd`
  - `BuildExitBootloaderCmd`
  - `BuildGetMetadataCmd`
  - `BuildGetAppStatusCmd`
  - `BuildSetActiveAppCmd`
- **protocol/responses.go**: Response parsing now validates checksums including SOP byte
- **Test files**: Updated all test expectations to match corrected checksum calculations

### Migration Guide

This change affects the binary protocol communication with bootloader devices. The checksum for every packet will be different by exactly 1 byte (the value of SOP, which is 0x01).

**Before v0.5.0 (INCORRECT)**:
```
Checksum = 2's complement of (CMD + LEN_L + LEN_H + DATA)
Example: 0x38 + 0x06 + 0x00 + [key bytes] = checksum 0xFC32
Command: 01 38 06 00 CA FE 00 00 CA FE 32 FC 17
                                        ^^--^^  Wrong checksum
```

**After v0.5.0 (CORRECT)**:
```
Checksum = 2's complement of (SOP + CMD + LEN_L + LEN_H + DATA)
Example: 0x01 + 0x38 + 0x06 + 0x00 + [key bytes] = checksum 0xFC33
Command: 01 38 06 00 CA FE 00 00 CA FE 33 FC 17
                                        ^^--^^  Correct checksum
```

#### Action Required

If you have:
- **Custom mock devices** or test fixtures that calculate checksums manually
- **Packet capture tools** with hardcoded checksum expectations
- **Documentation** with example packets

You **MUST** update them to include the SOP byte (0x01) in checksum calculations.

#### Why This Change

The official Infineon bootloader protocol specification (AN60317, Page 4, Pages 22-23) explicitly states:

> "The checksum is computed for the entire packet with the exception of the Checksum and End of Packet fields."

This clearly includes the SOP byte. Our previous implementation violated the specification by excluding it, which is why users experienced "checksum mismatch" errors (0x08) when communicating with real devices.

### Fixed

- Bootloader communication now works correctly with real Cypress/Infineon PSoC devices
- Checksum validation no longer fails with status code 0x08 (ErrChecksum)
- Library now conforms to official Infineon AN60317 bootloader protocol specification

---

## [0.4.1] - 2025-01-16

### Fixed

- Fixed type mismatch compilation errors in `protocol/checksum.go` by explicitly typing constants
- Fixed constant type inference issues that caused errors: `invalid operation: crc ^= uint16(b) << BitsPerByte`
- Updated typed constants: `ChecksumMask`, `CRC16Polynomial`, `CRC16InitialValue`, `CRC16HighBitMask`, `BitsPerByte`

### Changed

- Explicitly typed all checksum-related constants to avoid Go type inference problems

---

## [0.4.0] - 2025-01-15

### Added

- Type-safe `Phase` type for bootloader operation phases
- 50+ exported constants across multiple files for better code maintainability
- Comprehensive constant definitions:
  - `PhaseEntering`, `PhaseProgramming`, `PhaseVerifying`, `PhaseExiting`, `PhaseComplete`
  - `DefaultReadTimeout`, `DefaultWriteTimeout`, `DefaultChunkSize`, `DefaultRetries`, `MaxChunkSize`
  - `EnterBootloaderResponseSize`, `GetFlashSizeResponseSize`, `VerifyRowResponseSize`, etc.
  - `HeaderLength`, `MinimumRowLength`, `RowHeaderSize`, `RowChecksumSize`
  - `ChecksumMask`, `CRC16Polynomial`, `CRC16InitialValue`, `CRC16HighBitMask`

### Changed

- Replaced all magic numbers and strings with well-documented named constants
- `Progress.Phase` changed from `string` to `Phase` type for compile-time safety and IDE autocomplete
- Updated examples to use new `Phase` constants

### Improved

- Code readability and maintainability
- IDE autocomplete support for phase values
- API documentation with clear constant names

---

## [0.3.4] - 2025-01-14

### Fixed

- Fixed row checksum verification to include metadata (ArrayID, RowNum, DataSize) per Cypress protocol specification
- `CalculateRowChecksumWithMetadata()` now correctly computes device-expected checksum

### Added

- Command delay support via `WithCommandDelay()` option for devices requiring inter-command delays

---

## [0.3.2] - 2025-01-13

### Fixed

- Fixed HID packet padding to ensure correct 64-byte alignment for HID devices

---

## [0.3.1] - 2025-01-12

### Added

- Intel HEX format support for hybrid .cyacd files
- `parseIntelHexRow()` function to handle Intel HEX EOF headers followed by raw firmware data

---

## [0.3.0] - Initial Release

### Added

- Complete Cypress/Infineon bootloader protocol implementation
- Support for PSoC 3, PSoC 4, and PSoC 5LP devices
- `.cyacd` firmware file parser
- High-level `Programmer` API with progress callbacks
- Low-level protocol command builders and response parsers
- Error types with detailed diagnostics
- Comprehensive test coverage
- Examples for basic, advanced, and progress tracking use cases

[0.5.1]: https://github.com/moffa90/go-cyacd/compare/v0.5.0...v0.5.1
[0.5.0]: https://github.com/moffa90/go-cyacd/compare/v0.4.1...v0.5.0
[0.4.1]: https://github.com/moffa90/go-cyacd/compare/v0.4.0...v0.4.1
[0.4.0]: https://github.com/moffa90/go-cyacd/compare/v0.3.4...v0.4.0
[0.3.4]: https://github.com/moffa90/go-cyacd/compare/v0.3.2...v0.3.4
[0.3.2]: https://github.com/moffa90/go-cyacd/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/moffa90/go-cyacd/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/moffa90/go-cyacd/releases/tag/v0.3.0
