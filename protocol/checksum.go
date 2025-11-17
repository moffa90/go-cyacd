package protocol

// Checksum algorithm constants.
const (
	// ChecksumMask is the 16-bit mask used in checksum calculations
	ChecksumMask = 0xFFFF

	// CRC16Polynomial is the CRC-16-CCITT polynomial (0x1021)
	CRC16Polynomial = 0x1021

	// CRC16InitialValue is the CRC-16 initial value
	CRC16InitialValue = 0xFFFF

	// CRC16HighBitMask is the high bit mask for CRC-16 calculations
	CRC16HighBitMask = 0x8000

	// BitsPerByte is the number of bits per byte
	BitsPerByte = 8
)

// calculatePacketChecksum computes the 16-bit checksum for a packet frame.
// Uses basic summation per Infineon spec: sum all bytes, then 2's complement.
//
// The checksum is calculated over all bytes from CMD through DATA,
// excluding SOP, CHECKSUM, and EOP fields.
func calculatePacketChecksum(data []byte) uint16 {
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	// Return 2's complement: invert and add 1
	return 1 + (ChecksumMask ^ sum)
}

// CalculateRowChecksum computes the 8-bit checksum for a row's data.
// This is used in .cyacd file format and for row verification.
//
// The checksum is calculated by summing all bytes and taking 2's complement.
func CalculateRowChecksum(data []byte) byte {
	var sum byte
	for _, b := range data {
		sum += b
	}
	// Return 2's complement: invert and add 1
	return ^sum + 1
}

// CalculateRowChecksumWithMetadata computes the full row checksum including metadata.
// This is what the device actually verifies during row verification commands.
//
// The device checksum includes:
//   - The data checksum from the .cyacd file
//   - ArrayID (1 byte)
//   - RowNum (2 bytes, big-endian)
//   - DataSize (2 bytes, big-endian)
//
// This matches the Cypress bootloader protocol specification and is verified
// against the working bootloader-usb implementation.
func CalculateRowChecksumWithMetadata(dataChecksum byte, arrayID byte, rowNum uint16, dataSize uint16) byte {
	sum := dataChecksum
	sum += arrayID
	sum += byte(rowNum >> 8)   // RowNum high byte
	sum += byte(rowNum)         // RowNum low byte
	sum += byte(dataSize >> 8)  // Size high byte
	sum += byte(dataSize)       // Size low byte
	return sum
}

// calculateCRC16 computes CRC-16-CCITT checksum.
// Used when packet checksum type is CRC16.
//
// CRC-16-CCITT parameters:
//   - Polynomial: CRC16Polynomial
//   - Initial value: CRC16InitialValue
//   - No final XOR
func calculateCRC16(data []byte) uint16 {
	crc := CRC16InitialValue

	for _, b := range data {
		crc ^= uint16(b) << BitsPerByte
		for i := 0; i < BitsPerByte; i++ {
			if crc&CRC16HighBitMask != 0 {
				crc = (crc << 1) ^ CRC16Polynomial
			} else {
				crc = crc << 1
			}
		}
	}

	return crc
}
