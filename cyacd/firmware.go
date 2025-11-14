package cyacd

// Firmware represents a complete parsed .cyacd firmware file.
type Firmware struct {
	// SiliconID is the device silicon ID (4 bytes)
	SiliconID uint32

	// SiliconRev is the silicon revision (1 byte)
	SiliconRev byte

	// ChecksumType indicates the checksum algorithm:
	//   0x00 = Basic summation
	//   0x01 = CRC-16-CCITT
	ChecksumType byte

	// Rows contains all flash rows to be programmed
	Rows []*Row
}

// Row represents a single flash row from the .cyacd file.
type Row struct {
	// ArrayID is the flash array identifier
	ArrayID byte

	// RowNum is the flash row number
	RowNum uint16

	// Data is the flash row data to be programmed
	Data []byte

	// Checksum is the row checksum (for validation)
	Checksum byte
}
