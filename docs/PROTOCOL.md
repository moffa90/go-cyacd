# Infineon Bootloader Protocol

This document describes the Infineon/Cypress bootloader protocol as implemented in go-cyacd.

## Protocol Version

This library implements **Infineon Bootloader Protocol v1.30**, as documented in the Infineon AN60317 application note.

## Overview

The Infineon bootloader protocol is a packet-based, request-response protocol used for programming PSoC microcontrollers. Each command is sent as a framed packet, and the bootloader responds with a status code and optional data.

## Frame Structure

All packets (commands and responses) use the same frame structure:

```
[SOP][STATUS/CMD][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
```

| Field | Size | Value | Description |
|-------|------|-------|-------------|
| SOP | 1 byte | 0x01 | Start of Packet marker |
| STATUS/CMD | 1 byte | varies | Command code (host→device) or Status code (device→host) |
| LEN_L | 1 byte | varies | Data length low byte (little-endian) |
| LEN_H | 1 byte | varies | Data length high byte (little-endian) |
| DATA | 0-256 bytes | varies | Command/response data payload |
| CHECKSUM_L | 1 byte | varies | Checksum low byte (little-endian) |
| CHECKSUM_H | 1 byte | varies | Checksum high byte (little-endian) |
| EOP | 1 byte | 0x17 | End of Packet marker |

### Minimum Frame Size

The minimum frame size (with no data) is **7 bytes**:
```
[SOP][CMD/STATUS][0x00][0x00][CHECKSUM_L][CHECKSUM_H][EOP]
```

### Checksum Calculation

The checksum is a 16-bit value calculated over all bytes from SOP through DATA (excluding CHECKSUM and EOP):

```
checksum = 1 + (0xFFFF ^ sum_of_all_bytes)
```

This is a 2's complement checksum stored in little-endian format.

## Command Set

### 0x38 - Enter Bootloader

Enters bootloader mode and returns device information.

**Request:**
```
[SOP][0x38][0x06][0x00][KEY(6 bytes)][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (8 bytes):**
```
[SiliconID(4)][SiliconRev(1)][BootloaderVer(3)]
```

- **SiliconID**: 4-byte device identifier (little-endian)
- **SiliconRev**: 1-byte silicon revision
- **BootloaderVer**: 3-byte bootloader version [Major.Minor.Patch]

**Key:** A 6-byte security key that must match the key programmed into the device. Default is often `0x00, 0x00, 0x00, 0x00, 0x00, 0x00`.

### 0x32 - Get Flash Size

Queries the valid flash row range for a specified array.

**Request:**
```
[SOP][0x32][0x01][0x00][ArrayID(1)][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (4 bytes):**
```
[StartRow_L][StartRow_H][EndRow_L][EndRow_H]
```

- **StartRow**: First valid row number (little-endian)
- **EndRow**: Last valid row number (little-endian)

### 0x39 - Program Row

Programs a single flash row with data.

**Request:**
```
[SOP][0x39][LEN_L][LEN_H][ArrayID(1)][RowNum_L][RowNum_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
```

- **ArrayID**: Flash array identifier (typically 0x00)
- **RowNum**: Row number to program (little-endian)
- **DATA**: Row data to program

**Response:** Status code only (no data)

**Note:** For large rows that exceed packet size limits, use Send Data command first to buffer data, then Program Row with remaining data.

### 0x37 - Send Data

Sends a data chunk to be buffered by the bootloader for subsequent Program Row command.

**Request:**
```
[SOP][0x37][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response:** Status code only (no data)

**Usage:** When row data exceeds the maximum packet size (typically 64 bytes for USB), the data is split:
1. Send chunks using Send Data (up to 57 bytes per chunk)
2. Send final chunk with Program Row

The library handles this automatically based on the configured chunk size.

### 0x3A - Verify Row

Gets the checksum of a programmed row for verification.

**Request:**
```
[SOP][0x3A][0x03][0x00][ArrayID(1)][RowNum_L][RowNum_H][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (1 byte):**
```
[Checksum(1)]
```

The checksum is calculated by the device over the programmed row data and metadata:
```
checksum = DataChecksum + ArrayID + RowNum_H + RowNum_L + Size_H + Size_L
```

Where:
- **DataChecksum**: Checksum from .cyacd file
- **ArrayID, RowNum, Size**: Row metadata (big-endian for RowNum and Size)

### 0x31 - Verify Checksum

Verifies the entire application checksum.

**Request:**
```
[SOP][0x31][0x00][0x00][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (1 byte):**
```
[Valid(1)]
```

- **Valid**: 0x01 = valid checksum, 0x00 = invalid checksum

This verifies the overall application integrity before launching.

### 0x34 - Erase Row

Erases a single flash row.

**Request:**
```
[SOP][0x34][0x03][0x00][ArrayID(1)][RowNum_L][RowNum_H][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response:** Status code only (no data)

### 0x35 - Sync Bootloader

Resets the bootloader to a clean state, discarding any buffered data.

**Request:**
```
[SOP][0x35][0x00][0x00][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response:** Status code only (no data)

**Usage:** Use this if the host and bootloader get out of sync.

### 0x3B - Exit Bootloader

Exits bootloader mode and launches the application.

**Request:**
```
[SOP][0x3B][0x00][0x00][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response:** Status code only (no data)

The bootloader verifies the application before reset. If valid, the application executes; if invalid, the bootloader executes again.

### 0x3C - Get Metadata

Reports the first 56 bytes of metadata for a specified application.

**Request:**
```
[SOP][0x3C][0x01][0x00][AppNum(1)][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (56 bytes):**
```
[Metadata(56)]
```

**Usage:** Multi-application bootloaders only.

### 0x33 - Get App Status

Returns the status of a specified application.

**Request:**
```
[SOP][0x33][0x01][0x00][AppNum(1)][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response data (2 bytes):**
```
[Valid(1)][Active(1)]
```

- **Valid**: 0x01 = application is valid, 0x00 = invalid
- **Active**: 0x01 = application is active, 0x00 = inactive

**Usage:** Multi-application bootloaders only.

### 0x36 - Set Active App

Sets a bootloadable application as the active application.

**Request:**
```
[SOP][0x36][0x01][0x00][AppNum(1)][CHECKSUM_L][CHECKSUM_H][EOP]
```

**Response:** Status code only (no data)

**Usage:** Multi-application bootloaders only.

## Status Codes

All responses include a status code in the STATUS/CMD field:

| Code | Name | Description |
|------|------|-------------|
| 0x00 | Success | Command successful |
| 0x03 | ErrLength | Data length outside expected range |
| 0x04 | ErrData | Data is not of proper form |
| 0x05 | ErrCommand | Command not recognized |
| 0x06 | ErrKey | Bootloader key is invalid |
| 0x08 | ErrChecksum | Packet checksum mismatch |
| 0x09 | ErrArray | Flash array ID not valid |
| 0x0A | ErrRow | Flash row number not valid |
| 0x0C | ErrApp | Application not valid |
| 0x0D | ErrActive | Application is currently active |
| 0x0F | ErrUnknown | Unknown error occurred |

## Programming Sequence

A typical programming sequence:

1. **Enter Bootloader** (0x38)
   - Send bootloader key
   - Receive device info
   - Verify Silicon ID matches firmware

2. **Get Flash Size** (0x32)
   - Query valid row range
   - Validate all rows are in range

3. **Program Rows** (0x39 + 0x37)
   - For each row in firmware:
     - If row size > packet limit: Send Data chunks (0x37)
     - Program Row (0x39) with final data
     - Optionally verify row (0x3A)

4. **Verify Application** (0x31)
   - Verify overall application checksum

5. **Exit Bootloader** (0x3B)
   - Launch the application

## Packet Size Considerations

### USB HID Devices

USB HID devices typically use **64-byte packets**. The protocol overhead is:

- **Send Data**: 7 bytes overhead → 57 bytes available for data
- **Program Row**: 10 bytes overhead → 54 bytes available for row data

### Chunking Strategy

For rows larger than the packet limit:

1. Calculate chunks: `(RowSize + Overhead) > PacketSize`
2. Send full chunks using Send Data (57 bytes each)
3. Send remainder using Program Row

Example for a 256-byte row with 64-byte packets:
```
SendData(57 bytes)  # Chunk 1
SendData(57 bytes)  # Chunk 2
SendData(57 bytes)  # Chunk 3
SendData(57 bytes)  # Chunk 4
ProgramRow(28 bytes + metadata)  # Final chunk with row info
```

The library handles this automatically.

## Checksum Types

The protocol supports two checksum types (specified in .cyacd file header):

### Basic Sum (0x00)

```
checksum = 2's complement of sum of all data bytes
```

### CRC-16 (0x01)

CRC-16-CCITT algorithm:
- Polynomial: 0x1021
- Initial value: 0xFFFF
- No final XOR

The library automatically uses the correct checksum type based on the firmware file.

## Row Checksum Calculation

For .cyacd files, row checksums include metadata:

```
checksum = 2's complement of (ArrayID + RowNum + Size + Data)
```

Where:
- **ArrayID**: 1 byte
- **RowNum**: 2 bytes (big-endian in file)
- **Size**: 2 bytes (big-endian in file)
- **Data**: N bytes
- **Checksum**: 1 byte (stored in file)

When verifying, the device calculates:
```
device_checksum = file_checksum + ArrayID + RowNum_H + RowNum_L + Size_H + Size_L
```

## Error Handling

### Retry Strategy

For transient errors (e.g., ErrChecksum), the library can retry:
- Configure with `WithRetries(n)` option
- Default: no retries
- Recommended: 3 retries for production

### Fatal Errors

Some errors are fatal and should not be retried:
- ErrKey (wrong bootloader key)
- ErrCommand (unsupported command)
- ErrArray (invalid array ID)
- ErrRow (row out of range)

### Context Cancellation

All operations respect context cancellation:
```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

err := prog.Program(ctx, fw, key)
```

## Timing Considerations

### Command Delays

Some commands may require delays:
- **Program Row**: May need 10-50ms for flash write
- **Erase Row**: May need 10-50ms for flash erase

Configure with `WithCommandDelay()` option if needed.

### Timeouts

Recommended timeouts:
- **Per-command**: 2-5 seconds
- **Overall programming**: 30-120 seconds (depends on firmware size)

## Transport Considerations

The protocol is transport-agnostic. Common transports:

### USB HID
- Packet size: 64 bytes
- May include padding/report ID
- Use `WithChunkSize(57)` (default)

### UART
- Packet size: Limited by buffer
- No padding concerns
- Can use larger chunk sizes

### I2C/SPI
- Packet size: Limited by device
- Check device datasheet
- May need smaller chunk sizes

## Implementation Notes

### Endianness

- **Packet length**: Little-endian (protocol frames)
- **Checksum**: Little-endian (protocol frames)
- **Silicon ID**: Little-endian (device response)
- **Row numbers**: Little-endian in protocol, **big-endian** in .cyacd files
- **Size field**: Little-endian in protocol, **big-endian** in .cyacd files

### Buffer Management

The bootloader maintains an internal data buffer:
- Send Data adds to buffer
- Program Row uses buffer + provided data
- Sync Bootloader clears buffer

## Security Considerations

### Bootloader Key

- 6-byte key prevents unauthorized programming
- Default key (all zeros) provides no security
- Production devices should use non-default keys
- Key is sent in plaintext (no encryption)

### Application Verification

- Checksum verification prevents corrupted applications
- Bootloader validates before launching
- Invalid applications won't execute

## References

- **Infineon AN60317**: PSoC 3, PSoC 4, and PSoC 5LP Bootloader and Bootloadable
- **Protocol Version**: 1.30
- **Datasheet**: See `docs/infineon-component-bootloadablebootloader-v1.60-software-module-datasheets-en.pdf`

## See Also

- [README.md](../README.md) - Library overview and usage
- [CONTRIBUTING.md](../CONTRIBUTING.md) - Development guidelines
- [pkg.go.dev documentation](https://pkg.go.dev/github.com/moffa90/go-cyacd) - API reference
