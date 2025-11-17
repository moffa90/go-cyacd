package protocol

import (
	"encoding/binary"
	"fmt"
)

// ParseResponse extracts status code and data from a response frame.
// Validates frame structure, length, and checksum.
//
// Response frame structure:
//
//	[SOP][STATUS][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
//
// Returns the status code, data payload, and any validation error.
func ParseResponse(frame []byte) (statusCode byte, data []byte, err error) {
	if len(frame) < MinFrameSize {
		return 0, nil, fmt.Errorf("frame too short: got %d bytes, minimum is %d", len(frame), MinFrameSize)
	}

	if frame[0] != StartOfPacket {
		return 0, nil, fmt.Errorf("invalid start of packet: got 0x%02X, expected 0x%02X", frame[0], StartOfPacket)
	}

	if frame[len(frame)-1] != EndOfPacket {
		return 0, nil, fmt.Errorf("invalid end of packet: got 0x%02X, expected 0x%02X", frame[len(frame)-1], EndOfPacket)
	}

	statusCode = frame[1]
	dataLen := binary.LittleEndian.Uint16(frame[2:4])

	expectedLen := int(MinFrameSize + dataLen)
	if len(frame) != expectedLen {
		return 0, nil, fmt.Errorf("frame length mismatch: got %d bytes, expected %d (MinFrameSize=%d + dataLen=%d)",
			len(frame), expectedLen, MinFrameSize, dataLen)
	}

	// Verify checksum
	checksumExpected := binary.LittleEndian.Uint16(frame[len(frame)-3 : len(frame)-1])
	checksumActual := calculatePacketChecksum(frame[1 : len(frame)-3])

	if checksumExpected != checksumActual {
		return 0, nil, fmt.Errorf("checksum mismatch: got 0x%04X, expected 0x%04X",
			checksumActual, checksumExpected)
	}

	// Extract data if present
	if dataLen > 0 {
		data = frame[4 : 4+dataLen]
	}

	return statusCode, data, nil
}

// ParseEnterBootloaderResponse parses the Enter Bootloader command response.
// Returns device identification information.
//
// Data format (EnterBootloaderResponseSize bytes):
//
//	[SILICON_ID(4)][SILICON_REV(1)][BOOTLOADER_VER(3)]
func ParseEnterBootloaderResponse(data []byte) (*DeviceInfo, error) {
	if len(data) != EnterBootloaderResponseSize {
		return nil, fmt.Errorf("invalid data length for Enter Bootloader response: got %d bytes, expected %d", len(data), EnterBootloaderResponseSize)
	}

	info := &DeviceInfo{
		SiliconID:     binary.LittleEndian.Uint32(data[0:4]),
		SiliconRev:    data[4],
		BootloaderVer: [3]byte{data[5], data[6], data[7]},
	}

	return info, nil
}

// ParseGetFlashSizeResponse parses the Get Flash Size command response.
// Returns the valid flash row range for programming.
//
// Data format (4 bytes):
//
//	[START_ROW(2)][END_ROW(2)]
func ParseGetFlashSizeResponse(data []byte) (*FlashSize, error) {
	if len(data) != GetFlashSizeResponseSize {
		return nil, fmt.Errorf("invalid data length for Get Flash Size response: got %d bytes, expected %d", len(data), GetFlashSizeResponseSize)
	}

	size := &FlashSize{
		StartRow: binary.LittleEndian.Uint16(data[0:2]),
		EndRow:   binary.LittleEndian.Uint16(data[2:4]),
	}

	return size, nil
}

// ParseVerifyRowResponse parses the Verify Row command response.
// Returns the checksum byte for the verified row.
//
// Data format (1 byte):
//
//	[ROW_CHECKSUM]
func ParseVerifyRowResponse(data []byte) (byte, error) {
	if len(data) != VerifyRowResponseSize {
		return 0, fmt.Errorf("invalid data length for Verify Row response: got %d bytes, expected %d", len(data), VerifyRowResponseSize)
	}

	return data[0], nil
}

// ParseVerifyChecksumResponse parses the Verify Checksum command response.
// Returns the checksum validity indicator.
//
// Data format (1 byte):
//   - Non-zero: application checksum is valid
//   - Zero: checksums do not match (application invalid)
func ParseVerifyChecksumResponse(data []byte) (bool, error) {
	if len(data) != VerifyChecksumResponseSize {
		return false, fmt.Errorf("invalid data length for Verify Checksum response: got %d bytes, expected %d", len(data), VerifyChecksumResponseSize)
	}

	return data[0] != 0, nil
}

// ParseGetMetadataResponse parses the Get Metadata command response.
// Returns the first 56 bytes of application metadata.
//
// Data format (56 bytes): See Metadata type for field descriptions.
func ParseGetMetadataResponse(data []byte) (*Metadata, error) {
	if len(data) != GetMetadataResponseSize {
		return nil, fmt.Errorf("invalid data length for Get Metadata response: got %d bytes, expected %d", len(data), GetMetadataResponseSize)
	}

	// Parse metadata fields according to Infineon spec page 20-21
	metadata := &Metadata{
		Checksum: data[0],
	}

	// StartAddr is at different positions for PSoC 3 vs PSoC 4/5LP
	// For PSoC 4/5LP (little-endian): bytes 1-4
	metadata.StartAddr = binary.LittleEndian.Uint32(data[1:5])

	// LastRow: bytes 5-6 for PSoC 4/5LP
	metadata.LastRow = binary.LittleEndian.Uint16(data[5:7])

	// Length: bytes 9-12 for PSoC 4/5LP
	metadata.Length = binary.LittleEndian.Uint32(data[9:13])

	// Active: byte 16
	metadata.Active = data[16]

	// Verified: byte 17
	metadata.Verified = data[17]

	// Bootloader version: bytes 18-19
	metadata.BootloaderVersion = binary.LittleEndian.Uint16(data[18:20])

	// App ID: bytes 20-21
	metadata.AppID = binary.LittleEndian.Uint16(data[20:22])

	// App Version: bytes 22-23
	metadata.AppVersion = binary.LittleEndian.Uint16(data[22:24])

	// Custom ID: bytes 24-27
	metadata.CustomID = binary.LittleEndian.Uint32(data[24:28])

	return metadata, nil
}

// ParseGetAppStatusResponse parses the Get Application Status command response.
// Returns the status of the specified application (multi-app only).
//
// Data format (2 bytes):
//
//	[VALID(1)][ACTIVE(1)]
func ParseGetAppStatusResponse(data []byte) (*AppStatus, error) {
	if len(data) != GetAppStatusResponseSize {
		return nil, fmt.Errorf("invalid data length for Get App Status response: got %d bytes, expected %d", len(data), GetAppStatusResponseSize)
	}

	status := &AppStatus{
		Valid:  data[0] != 0,
		Active: data[1] != 0,
	}

	return status, nil
}
