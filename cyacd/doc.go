// Package cyacd provides parsing for Cypress .cyacd firmware files.
//
// # CYACD File Format
//
// The .cyacd file format is used to store bootloadable firmware for Cypress microcontrollers.
// It consists of a header line followed by multiple row lines, all hex-encoded.
//
// Header Format (12 hex characters):
//
//	[SiliconID(8)][SiliconRev(2)][ChecksumType(2)]
//
// Example header:
//
//	1E9602AA00
//	  1E9602AA = Silicon ID (0x1E9602AA)
//	  00 = Silicon Revision (0x00)
//	  00 = Checksum Type (0x00 = basic summation)
//
// Row Format (variable length):
//
//	[ArrayID(2)][RowNum(4)][DataLen(4)][Data(variable)][Checksum(2)]
//
// Example row:
//
//	000000040401020304F6
//	  00 = Array ID
//	  0000 = Row Number (little-endian)
//	  0400 = Data Length (little-endian, 4 bytes)
//	  01020304 = Row Data
//	  F6 = Checksum
//
// # Usage
//
// Parse a .cyacd file from disk:
//
//	fw, err := cyacd.Parse("firmware.cyacd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	fmt.Printf("Silicon ID: 0x%08X\n", fw.SiliconID)
//	fmt.Printf("Total rows: %d\n", len(fw.Rows))
//
//	for i, row := range fw.Rows {
//	    fmt.Printf("Row %d: Array=%d, RowNum=%d, DataLen=%d\n",
//	        i, row.ArrayID, row.RowNum, len(row.Data))
//	}
//
// Parse from an io.Reader:
//
//	data := strings.NewReader(cyacdContent)
//	fw, err := cyacd.ParseReader(data)
//
// # Error Handling
//
// Parse returns detailed errors for invalid files:
//   - Invalid header format
//   - Invalid checksum type
//   - Row parsing errors with line numbers
//   - Checksum mismatches
//   - Invalid hex encoding
//
// All errors include context about what failed and where.
package cyacd
