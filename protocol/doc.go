// Package protocol implements the Cypress/Infineon bootloader communication protocol.
//
// This package provides functions to build command frames and parse response frames
// according to the Infineon Bootloader Protocol Specification v1.30.
//
// # Protocol Overview
//
// The bootloader protocol uses a packet-based communication structure:
//
//	Command:  [SOP][CMD][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
//	Response: [SOP][STATUS][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
//
// Where:
//   - SOP = Start of Packet (0x01)
//   - EOP = End of Packet (0x17)
//   - LEN = 16-bit data length (little-endian)
//   - CHECKSUM = 16-bit checksum (little-endian, 2's complement)
//
// # Command Builders
//
// Use the Build* functions to create command frames:
//
//	frame, err := protocol.BuildEnterBootloaderCmd(key)
//	frame, err := protocol.BuildProgramRowCmd(arrayID, rowNum, data)
//	// ... etc
//
// # Response Parsers
//
// Use ParseResponse to validate and extract data from response frames:
//
//	statusCode, data, err := protocol.ParseResponse(frame)
//	if statusCode != protocol.StatusSuccess {
//	    return fmt.Errorf("command failed: 0x%02X", statusCode)
//	}
//
// Then use the Parse* functions for command-specific data:
//
//	info, err := protocol.ParseEnterBootloaderResponse(data)
//	size, err := protocol.ParseGetFlashSizeResponse(data)
//	// ... etc
//
// # Error Handling
//
// Status codes other than StatusSuccess indicate errors.
// Use the ProtocolError type for structured error information:
//
//	if statusCode != protocol.StatusSuccess {
//	    err := &protocol.ProtocolError{
//	        Operation: "enter bootloader",
//	        StatusCode: statusCode,
//	    }
//	    // err.Error() returns: "enter bootloader failed: checksum mismatch (0x08)"
//	}
//
// # Reference
//
// For complete protocol details, see:
// Infineon-Component_Bootloadable_Bootloader_V1.30-Software Module Datasheets-v01_06-EN.pdf
package protocol
