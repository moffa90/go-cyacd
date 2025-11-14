package protocol

// DeviceInfo contains bootloader device identification information.
// Returned by the Enter Bootloader command.
type DeviceInfo struct {
	// SiliconID is the device silicon ID (4 bytes)
	SiliconID uint32

	// SiliconRev is the silicon revision (1 byte)
	SiliconRev byte

	// BootloaderVer is the bootloader version [major, minor, patch]
	BootloaderVer [3]byte
}

// FlashSize contains the valid flash row range for programming.
// Returned by the Get Flash Size command.
type FlashSize struct {
	// StartRow is the first programmable row number
	StartRow uint16

	// EndRow is the last programmable row number (inclusive)
	EndRow uint16
}

// Metadata contains application metadata (56 bytes total).
// Returned by the Get Metadata command.
type Metadata struct {
	// Checksum is the bootloadable application checksum
	Checksum byte

	// StartAddr is the startup routine address of the bootloadable application
	StartAddr uint32

	// LastRow is the last flash row occupied by the application
	LastRow uint16

	// Length is the size of the bootloadable application in bytes
	Length uint32

	// Active indicates the active bootloadable application
	Active byte

	// Verified is the bootloadable application verification status
	Verified byte

	// BootloaderVersion is the bootloader application version
	BootloaderVersion uint16

	// AppID is the bootloadable application ID
	AppID uint16

	// AppVersion is the bootloadable application version
	AppVersion uint16

	// CustomID is the bootloadable application custom ID (4 bytes)
	CustomID uint32
}

// AppStatus contains application status information.
// Returned by Get Application Status command (multi-app only).
type AppStatus struct {
	// Valid indicates if the application is valid
	Valid bool

	// Active indicates if the application is active
	Active bool
}
