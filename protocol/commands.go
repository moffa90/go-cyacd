package protocol

import (
	"encoding/binary"
	"fmt"
)

// BuildEnterBootloaderCmd constructs an Enter Bootloader command frame.
// The key must be exactly BootloaderKeySize bytes as specified in the Infineon protocol.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][KEY(6)][CHECKSUM_L][CHECKSUM_H][EOP]
//
// Returns the complete frame ready to send, or an error if validation fails.
func BuildEnterBootloaderCmd(key []byte) ([]byte, error) {
	if len(key) != BootloaderKeySize {
		return nil, fmt.Errorf("key must be exactly %d bytes, got %d", BootloaderKeySize, len(key))
	}

	dataLen := uint16(len(key))
	frame := make([]byte, 0, MinFrameSize+len(key))

	// Start of packet
	frame = append(frame, StartOfPacket)

	// Command
	frame = append(frame, CmdEnterBootloader)

	// Data length (little-endian)
	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	// Data (6-byte key)
	frame = append(frame, key...)

	// Calculate and append checksum (exclude SOP, include everything else up to checksum)
	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	// End of packet
	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildGetFlashSizeCmd constructs a Get Flash Size command frame.
// The arrayID specifies which flash array to query.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][ARRAY_ID][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildGetFlashSizeCmd(arrayID byte) ([]byte, error) {
	dataLen := uint16(1) // Just the array ID
	frame := make([]byte, 0, MinFrameSize+1)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdGetFlashSize)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, arrayID)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildProgramRowCmd constructs a Program Row command frame.
// Programs a single flash row with the provided data.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][ARRAY_ID][ROW_L][ROW_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
//
// The data length should not exceed the maximum row size for the device.
func BuildProgramRowCmd(arrayID byte, rowNum uint16, data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	if len(data) > MaxDataSize {
		return nil, fmt.Errorf("data length %d exceeds maximum %d bytes", len(data), MaxDataSize)
	}

	// Payload: arrayID(1) + rowNum(2) + data
	payloadSize := 1 + 2 + len(data)
	dataLen := uint16(payloadSize)

	frame := make([]byte, 0, MinFrameSize+payloadSize)
	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdProgramRow)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, arrayID)

	rowBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(rowBytes, rowNum)
	frame = append(frame, rowBytes...)

	frame = append(frame, data...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildSendDataCmd constructs a Send Data command frame.
// Sends a block of data to be buffered by the bootloader.
//
// This is used to break up large data transfers into smaller chunks.
// The data is buffered and used by subsequent commands like Program Row.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][DATA...][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildSendDataCmd(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("data cannot be empty")
	}
	if len(data) > MaxDataSize {
		return nil, fmt.Errorf("data length %d exceeds maximum %d bytes", len(data), MaxDataSize)
	}

	dataLen := uint16(len(data))
	frame := make([]byte, 0, MinFrameSize+len(data))

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdSendData)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, data...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildVerifyRowCmd constructs a Verify Row command frame.
// Gets the checksum of the specified flash row for verification.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][ARRAY_ID][ROW_L][ROW_H][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildVerifyRowCmd(arrayID byte, rowNum uint16) ([]byte, error) {
	dataLen := uint16(3) // arrayID(1) + rowNum(2)
	frame := make([]byte, 0, MinFrameSize+3)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdVerifyRow)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, arrayID)

	rowBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(rowBytes, rowNum)
	frame = append(frame, rowBytes...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildVerifyChecksumCmd constructs a Verify Checksum command frame.
// Verifies the entire application checksum.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildVerifyChecksumCmd() ([]byte, error) {
	dataLen := uint16(0) // No data payload
	frame := make([]byte, 0, MinFrameSize)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdVerifyChecksum)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildEraseRowCmd constructs an Erase Row command frame.
// Erases the contents of the specified flash row.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][ARRAY_ID][ROW_L][ROW_H][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildEraseRowCmd(arrayID byte, rowNum uint16) ([]byte, error) {
	dataLen := uint16(3) // arrayID(1) + rowNum(2)
	frame := make([]byte, 0, MinFrameSize+3)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdEraseRow)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, arrayID)

	rowBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(rowBytes, rowNum)
	frame = append(frame, rowBytes...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildSyncBootloaderCmd constructs a Sync Bootloader command frame.
// Resets the bootloader to a clean state, discarding any buffered data.
//
// This command is only needed if the host and bootloader get out of sync.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildSyncBootloaderCmd() ([]byte, error) {
	dataLen := uint16(0)
	frame := make([]byte, 0, MinFrameSize)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdSyncBootloader)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildExitBootloaderCmd constructs an Exit Bootloader command frame.
// Exits the bootloader and launches the application via software reset.
//
// Before reset, the bootloadable application is verified. If valid,
// it executes after reset. If invalid, bootloader executes again.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildExitBootloaderCmd() ([]byte, error) {
	dataLen := uint16(0)
	frame := make([]byte, 0, MinFrameSize)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdExitBootloader)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildGetMetadataCmd constructs a Get Metadata command frame.
// Reports the first 56 bytes of metadata for the specified application.
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][APP_NUM][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildGetMetadataCmd(appNum byte) ([]byte, error) {
	dataLen := uint16(1)
	frame := make([]byte, 0, MinFrameSize+1)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdGetMetadata)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, appNum)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildGetAppStatusCmd constructs a Get Application Status command frame.
// Returns the status of the specified application (multi-application bootloader only).
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][APP_NUM][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildGetAppStatusCmd(appNum byte) ([]byte, error) {
	dataLen := uint16(1)
	frame := make([]byte, 0, MinFrameSize+1)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdGetAppStatus)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, appNum)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}

// BuildSetActiveAppCmd constructs a Set Active Application command frame.
// Sets the specified bootloadable application as active (multi-application bootloader only).
//
// Frame structure:
//
//	[SOP][CMD][LEN_L][LEN_H][APP_NUM][CHECKSUM_L][CHECKSUM_H][EOP]
func BuildSetActiveAppCmd(appNum byte) ([]byte, error) {
	dataLen := uint16(1)
	frame := make([]byte, 0, MinFrameSize+1)

	frame = append(frame, StartOfPacket)
	frame = append(frame, CmdSetActiveApp)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, appNum)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame, nil
}
