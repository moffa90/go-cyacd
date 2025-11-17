package protocol

// ProtocolVersion is the Infineon bootloader protocol version implemented by this library.
const ProtocolVersion = "1.30"

// Frame structure constants per Infineon spec.
const (
	// StartOfPacket is the frame start marker (0x01)
	StartOfPacket = 0x01

	// EndOfPacket is the frame end marker (0x17)
	EndOfPacket = 0x17

	// MinFrameSize is the minimum frame size in bytes:
	// SOP(1) + CMD/STATUS(1) + LEN(2) + CHECKSUM(2) + EOP(1)
	MinFrameSize = 7
)

// Command codes per Infineon spec section 6.2 (page 24).
const (
	// CmdEnterBootloader enters bootloader mode with a 6-byte key
	CmdEnterBootloader = 0x38

	// CmdGetFlashSize queries the valid flash row range for an array
	CmdGetFlashSize = 0x32

	// CmdProgramRow programs a single flash row
	CmdProgramRow = 0x39

	// CmdEraseRow erases a single flash row
	CmdEraseRow = 0x34

	// CmdVerifyRow gets the checksum of a programmed row
	CmdVerifyRow = 0x3A

	// CmdVerifyChecksum verifies the entire application checksum
	CmdVerifyChecksum = 0x31

	// CmdSendData sends a data chunk (for large rows)
	CmdSendData = 0x37

	// CmdSyncBootloader resets bootloader to clean state
	CmdSyncBootloader = 0x35

	// CmdExitBootloader exits bootloader and launches application
	CmdExitBootloader = 0x3B

	// CmdGetMetadata reports application metadata (56 bytes)
	CmdGetMetadata = 0x3C

	// CmdGetAppStatus returns status of specified application (multi-app only)
	CmdGetAppStatus = 0x33

	// CmdSetActiveApp sets the active bootloadable application (multi-app only)
	CmdSetActiveApp = 0x36
)

// Status/Error codes per Infineon spec page 23.
const (
	// StatusSuccess indicates command was successfully received and executed
	StatusSuccess = 0x00

	// ErrLength indicates data amount is outside expected range
	ErrLength = 0x03

	// ErrData indicates data is not of proper form
	ErrData = 0x04

	// ErrCommand indicates command is not recognized
	ErrCommand = 0x05

	// ErrKey indicates bootloader key is invalid
	ErrKey = 0x06

	// ErrChecksum indicates packet checksum doesn't match expected value
	ErrChecksum = 0x08

	// ErrArray indicates flash array ID is not valid
	ErrArray = 0x09

	// ErrRow indicates flash row number is not valid
	ErrRow = 0x0A

	// ErrApp indicates application is not valid and cannot be set as active
	ErrApp = 0x0C

	// ErrActive indicates application is currently marked as active
	ErrActive = 0x0D

	// ErrUnknown indicates an unknown error occurred
	ErrUnknown = 0x0F
)

// Checksum types per Infineon spec page 4.
const (
	// ChecksumBasicSum uses basic summation: sum all bytes, then 2's complement
	ChecksumBasicSum = 0x00

	// ChecksumCRC16 uses CRC-16-CCITT algorithm
	ChecksumCRC16 = 0x01
)

// MaxDataSize is the maximum data payload size per packet.
// This is derived from typical USB packet sizes minus protocol overhead.
const MaxDataSize = 256

// DefaultChunkSize is the recommended chunk size for SendData operations.
// Set to 57 bytes to fit within 64-byte USB packets (64 - 7 byte overhead).
const DefaultChunkSize = 57

// Response data sizes per Infineon bootloader protocol specification.
const (
	// EnterBootloaderResponseSize is the data size for Enter Bootloader response (8 bytes)
	EnterBootloaderResponseSize = 8

	// GetFlashSizeResponseSize is the data size for Get Flash Size response (4 bytes)
	GetFlashSizeResponseSize = 4

	// VerifyRowResponseSize is the data size for Verify Row response (1 byte)
	VerifyRowResponseSize = 1

	// VerifyChecksumResponseSize is the data size for Verify Checksum response (1 byte)
	VerifyChecksumResponseSize = 1

	// GetMetadataResponseSize is the data size for Get Metadata response (56 bytes)
	GetMetadataResponseSize = 56

	// GetAppStatusResponseSize is the data size for Get App Status response (2 bytes)
	GetAppStatusResponseSize = 2

	// BootloaderKeySize is the required size for bootloader key (6 bytes)
	BootloaderKeySize = 6

	// DefaultResponseBufferSize is the default buffer size for reading responses (512 bytes)
	// Large enough to handle HID packets with padding
	DefaultResponseBufferSize = 512
)
