package bootloader

import (
	"strings"
	"testing"
)

func TestDeviceMismatchError(t *testing.T) {
	err := &DeviceMismatchError{
		Expected: 0x12345678,
		Actual:   0x87654321,
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "device mismatch") {
		t.Errorf("error message should contain 'device mismatch', got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "0x12345678") {
		t.Errorf("error message should contain expected ID, got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "0x87654321") {
		t.Errorf("error message should contain actual ID, got: %s", errMsg)
	}
}

func TestRowOutOfRangeError(t *testing.T) {
	err := &RowOutOfRangeError{
		RowNum:  500,
		MinRow:  0,
		MaxRow:  255,
		ArrayID: 1,
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "row 500") {
		t.Errorf("error message should contain row number, got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "out of range") {
		t.Errorf("error message should contain 'out of range', got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "0-255") {
		t.Errorf("error message should contain range, got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "array 1") {
		t.Errorf("error message should contain array ID, got: %s", errMsg)
	}
}

func TestChecksumMismatchError(t *testing.T) {
	err := &ChecksumMismatchError{
		RowNum:   42,
		Expected: 0xAB,
		Actual:   0xCD,
	}

	errMsg := err.Error()

	if !strings.Contains(errMsg, "checksum mismatch") {
		t.Errorf("error message should contain 'checksum mismatch', got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "row 42") {
		t.Errorf("error message should contain row number, got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "0xAB") {
		t.Errorf("error message should contain expected checksum, got: %s", errMsg)
	}

	if !strings.Contains(errMsg, "0xCD") {
		t.Errorf("error message should contain actual checksum, got: %s", errMsg)
	}
}

func TestVerificationError(t *testing.T) {
	tests := []struct {
		name    string
		err     *VerificationError
		wantMsg string
	}{
		{
			name: "with reason",
			err: &VerificationError{
				Reason: "invalid application signature",
			},
			wantMsg: "invalid application signature",
		},
		{
			name: "without reason",
			err: &VerificationError{
				Reason: "",
			},
			wantMsg: "application verification failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if !strings.Contains(errMsg, tt.wantMsg) {
				t.Errorf("error message should contain %q, got: %s", tt.wantMsg, errMsg)
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	// Test that all error types implement error interface
	var _ error = &DeviceMismatchError{}
	var _ error = &RowOutOfRangeError{}
	var _ error = &ChecksumMismatchError{}
	var _ error = &VerificationError{}
}
