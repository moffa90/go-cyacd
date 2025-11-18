package cyacd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
)

// Constants for CYACD file format parsing.
const (
	// HeaderLength is the expected length of the header line in hex characters
	HeaderLength = 12

	// MinimumRowLength is the minimum length for a row line in hex characters
	MinimumRowLength = 12

	// MinimumRowDataBytes is the minimum number of bytes in a row
	MinimumRowDataBytes = 6

	// RowHeaderSize is the size of row metadata (arrayID + rowNum + dataLen)
	RowHeaderSize = 5

	// RowChecksumSize is the size of the row checksum field
	RowChecksumSize = 1

	// IntelHexHeaderLength is the length of the Intel HEX header in hex characters
	IntelHexHeaderLength = 10

	// DefaultRowCapacity is the default initial capacity for the rows slice
	DefaultRowCapacity = 256
)

// Parse parses a .cyacd file from the given file path.
// Returns the complete firmware structure or an error if parsing fails.
//
// Example:
//
//	fw, err := cyacd.Parse("firmware.cyacd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Silicon ID: 0x%08X\n", fw.SiliconID)
func Parse(path string) (*Firmware, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = f.Close() }()

	return ParseReader(f)
}

// ParseReader parses a .cyacd file from any io.Reader.
// This is useful for testing and reading from non-file sources.
//
// Example:
//
//	data := strings.NewReader(cyacdContent)
//	fw, err := cyacd.ParseReader(data)
func ParseReader(r io.Reader) (*Firmware, error) {
	scanner := bufio.NewScanner(r)

	// Parse header (first line)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("failed to read header: %w", err)
		}
		return nil, fmt.Errorf("empty file")
	}

	header := scanner.Text()
	fw, err := parseHeader(header)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	// Parse rows
	lineNum := 1
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Check if this is Intel HEX format (starts with ':')
		var row *Row
		var err error
		if line != "" && line[0] == ':' {
			row, err = parseIntelHexRow(line)
		} else {
			row, err = parseRow(line)
		}

		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum, err)
		}

		fw.Rows = append(fw.Rows, row)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(fw.Rows) == 0 {
		return nil, fmt.Errorf("no rows found in file")
	}

	return fw, nil
}

// parseHeader parses the .cyacd file header.
//
// Header format (12 hex characters):
//
//	[SiliconID(4 bytes)][SiliconRev(1 byte)][ChecksumType(1 byte)]
//
// Example: "1E9602AA00" = SiliconID: 0x1E9602AA, Rev: 0x00, Checksum: 0x00
func parseHeader(line string) (*Firmware, error) {
	if len(line) != HeaderLength {
		return nil, fmt.Errorf("invalid header length: got %d characters, expected %d", len(line), HeaderLength)
	}

	data, err := hex.DecodeString(line)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	// Silicon ID is big-endian in the file
	siliconID := uint32(data[0])<<24 | uint32(data[1])<<16 |
		uint32(data[2])<<8 | uint32(data[3])

	fw := &Firmware{
		SiliconID:    siliconID,
		SiliconRev:   data[4],
		ChecksumType: data[5],
		Rows:         make([]*Row, 0, DefaultRowCapacity),
	}

	// Validate checksum type
	if fw.ChecksumType != 0x00 && fw.ChecksumType != 0x01 {
		return nil, fmt.Errorf("invalid checksum type: 0x%02X (must be 0x00 or 0x01)", fw.ChecksumType)
	}

	return fw, nil
}

// parseRow parses a single row line from the .cyacd file.
//
// Row format:
//
//	[ArrayID(1 byte)][RowNum(2 bytes)][DataLen(2 bytes)][Data(N bytes)][Checksum(1 byte)]
//
// All values are hex-encoded. RowNum and DataLen are little-endian.
//
// Example: "000000040401020304 0E"
//
//	ArrayID: 0x00
//	RowNum: 0x0000 (little-endian)
//	DataLen: 0x0004 (little-endian)
//	Data: [0x01, 0x02, 0x03, 0x04]
//	Checksum: 0x0E
func parseRow(line string) (*Row, error) {
	// Minimum row: arrayID(2) + rowNum(4) + dataLen(4) + checksum(2) = MinimumRowLength chars
	if len(line) < MinimumRowLength {
		return nil, fmt.Errorf("row too short: got %d characters, minimum is %d", len(line), MinimumRowLength)
	}

	data, err := hex.DecodeString(line)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	if len(data) < MinimumRowDataBytes {
		return nil, fmt.Errorf("row data too short: got %d bytes, minimum is %d", len(data), MinimumRowDataBytes)
	}

	arrayID := data[0]
	rowNum := uint16(data[1]) | uint16(data[2])<<8  // Little-endian
	dataLen := uint16(data[3]) | uint16(data[4])<<8 // Little-endian

	expectedLen := int(RowHeaderSize + RowChecksumSize + dataLen)
	if len(data) != expectedLen {
		return nil, fmt.Errorf("data length mismatch: got %d bytes, expected %d (header=%d + data=%d + checksum=%d)",
			len(data), expectedLen, RowHeaderSize, dataLen, RowChecksumSize)
	}

	rowData := data[RowHeaderSize : RowHeaderSize+dataLen]
	checksum := data[len(data)-1]

	// Verify checksum
	calculatedChecksum := calculateRowChecksum(data[:len(data)-1])
	if checksum != calculatedChecksum {
		return nil, fmt.Errorf("checksum mismatch: got 0x%02X, expected 0x%02X",
			checksum, calculatedChecksum)
	}

	row := &Row{
		ArrayID:  arrayID,
		RowNum:   rowNum,
		Size:     dataLen,
		Data:     make([]byte, len(rowData)),
		Checksum: checksum,
	}
	copy(row.Data, rowData)

	return row, nil
}

// parseIntelHexRow parses a row in PSoC hybrid format (starting with ':').
// Despite the ':' prefix, this format is actually CYACD format with a colon prefix,
// NOT true Intel HEX format. This matches the reference bootloader-usb implementation.
//
// Format after removing ':': [ArrayID(1)][RowNum(2)][Size(2)][Data(N)][Checksum(1)]
// All multi-byte fields use BIG-ENDIAN byte order (unlike standard CYACD which uses little-endian)
//
// Example: :000045010000800020...
//
//	After removing ':': 000045010000800020...
//	- ArrayID: 00
//	- RowNum: 0045 (big-endian) = 69
//	- Size: 0100 (big-endian) = 256
//	- Data: 00800020... (256 bytes)
//	- Checksum: last byte
func parseIntelHexRow(line string) (*Row, error) {
	// Remove the leading ':'
	if len(line) < 1 || line[0] != ':' {
		return nil, fmt.Errorf("hybrid row must start with ':'")
	}

	line = line[1:] // Strip ':'

	// Minimum row length check
	if len(line) < MinimumRowLength {
		return nil, fmt.Errorf("hybrid row too short: got %d characters, minimum is %d", len(line), MinimumRowLength)
	}

	// Hex decode the entire line (CYACD format)
	data, err := hex.DecodeString(line)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	if len(data) < MinimumRowDataBytes {
		return nil, fmt.Errorf("row data too short: got %d bytes, minimum is %d", len(data), MinimumRowDataBytes)
	}

	// Parse CYACD format with BIG-ENDIAN byte order (reference implementation behavior)
	arrayID := data[0]
	rowNum := uint16(data[1])<<8 | uint16(data[2])  // BIG-ENDIAN
	dataLen := uint16(data[3])<<8 | uint16(data[4]) // BIG-ENDIAN

	expectedLen := int(RowHeaderSize + RowChecksumSize + dataLen)
	if len(data) != expectedLen {
		return nil, fmt.Errorf("data length mismatch: got %d bytes, expected %d (header=%d + data=%d + checksum=%d)",
			len(data), expectedLen, RowHeaderSize, dataLen, RowChecksumSize)
	}

	rowData := data[RowHeaderSize : RowHeaderSize+dataLen]
	checksum := data[len(data)-1]

	// Verify checksum
	calculatedChecksum := calculateRowChecksum(data[:len(data)-1])
	if checksum != calculatedChecksum {
		return nil, fmt.Errorf("checksum mismatch: got 0x%02X, expected 0x%02X",
			checksum, calculatedChecksum)
	}

	row := &Row{
		ArrayID:  arrayID,
		RowNum:   rowNum,
		Size:     dataLen,
		Data:     make([]byte, len(rowData)),
		Checksum: checksum,
	}
	copy(row.Data, rowData)

	return row, nil
}

// calculateRowChecksum computes the 8-bit checksum for a row.
// Uses basic summation with 2's complement.
func calculateRowChecksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	return ^sum + 1 // 2's complement
}
