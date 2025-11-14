package cyacd

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io"
	"os"
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
	defer f.Close()

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

		row, err := parseRow(line)
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
	if len(line) != 12 {
		return nil, fmt.Errorf("invalid header length: got %d characters, expected 12", len(line))
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
		Rows:         make([]*Row, 0, 256), // Reasonable initial capacity
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
	// Minimum row: arrayID(2) + rowNum(4) + dataLen(4) + checksum(2) = 12 chars
	if len(line) < 12 {
		return nil, fmt.Errorf("row too short: got %d characters, minimum is 12", len(line))
	}

	data, err := hex.DecodeString(line)
	if err != nil {
		return nil, fmt.Errorf("invalid hex data: %w", err)
	}

	if len(data) < 6 {
		return nil, fmt.Errorf("row data too short: got %d bytes, minimum is 6", len(data))
	}

	arrayID := data[0]
	rowNum := uint16(data[1]) | uint16(data[2])<<8    // Little-endian
	dataLen := uint16(data[3]) | uint16(data[4])<<8   // Little-endian

	expectedLen := int(6 + dataLen) // header(5) + data + checksum(1)
	if len(data) != expectedLen {
		return nil, fmt.Errorf("data length mismatch: got %d bytes, expected %d (header=5 + data=%d + checksum=1)",
			len(data), expectedLen, dataLen)
	}

	rowData := data[5 : 5+dataLen]
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
