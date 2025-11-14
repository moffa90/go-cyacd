package bootloader

import (
	"fmt"
)

// DeviceMismatchError indicates that the device silicon ID doesn't match the firmware.
type DeviceMismatchError struct {
	Expected uint32
	Actual   uint32
}

func (e *DeviceMismatchError) Error() string {
	return fmt.Sprintf("device mismatch: firmware expects silicon ID 0x%08X, device has 0x%08X",
		e.Expected, e.Actual)
}

// RowOutOfRangeError indicates that a firmware row is outside the device's flash range.
type RowOutOfRangeError struct {
	RowNum uint16
	MinRow uint16
	MaxRow uint16
}

func (e *RowOutOfRangeError) Error() string {
	return fmt.Sprintf("row %d is out of range: valid range is %d-%d",
		e.RowNum, e.MinRow, e.MaxRow)
}

// ChecksumMismatchError indicates that a row checksum verification failed.
type ChecksumMismatchError struct {
	RowNum   uint16
	Expected byte
	Actual   byte
}

func (e *ChecksumMismatchError) Error() string {
	return fmt.Sprintf("checksum mismatch for row %d: expected 0x%02X, got 0x%02X",
		e.RowNum, e.Expected, e.Actual)
}

// VerificationError indicates that the application checksum verification failed.
type VerificationError struct {
	Message string
}

func (e *VerificationError) Error() string {
	return fmt.Sprintf("firmware verification failed: %s", e.Message)
}
