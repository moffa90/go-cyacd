package protocol

import "fmt"

// ProtocolError represents an error returned by the bootloader.
// Contains the status code from the bootloader response.
type ProtocolError struct {
	// Operation is the command that failed
	Operation string

	// StatusCode is the error code from the bootloader
	StatusCode byte
}

func (e *ProtocolError) Error() string {
	statusName := getStatusName(e.StatusCode)
	return fmt.Sprintf("%s failed: %s (0x%02X)", e.Operation, statusName, e.StatusCode)
}

// IsProtocolError returns true if the error is a ProtocolError.
func IsProtocolError(err error) bool {
	_, ok := err.(*ProtocolError)
	return ok
}

// getStatusName returns a human-readable name for a status code.
func getStatusName(code byte) string {
	switch code {
	case StatusSuccess:
		return "success"
	case ErrLength:
		return "invalid length"
	case ErrData:
		return "invalid data"
	case ErrCommand:
		return "unrecognized command"
	case ErrChecksum:
		return "checksum mismatch"
	case ErrArray:
		return "invalid array ID"
	case ErrRow:
		return "invalid row number"
	case ErrApp:
		return "invalid application"
	case ErrActive:
		return "application is active"
	case ErrUnknown:
		return "unknown error"
	default:
		return fmt.Sprintf("unknown status code 0x%02X", code)
	}
}
