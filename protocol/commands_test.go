package protocol

import (
	"bytes"
	"testing"
)

func TestBuildEnterBootloaderCmd(t *testing.T) {
	tests := []struct {
		name    string
		key     []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid 6-byte key",
			key:     []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			wantErr: false,
		},
		{
			name:    "invalid 5-byte key",
			key:     []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E},
			wantErr: true,
			errMsg:  "key must be exactly 6 bytes",
		},
		{
			name:    "invalid 7-byte key",
			key:     []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F, 0x6A},
			wantErr: true,
			errMsg:  "key must be exactly 6 bytes",
		},
		{
			name:    "nil key",
			key:     nil,
			wantErr: true,
			errMsg:  "key must be exactly 6 bytes",
		},
		{
			name:    "empty key",
			key:     []byte{},
			wantErr: true,
			errMsg:  "key must be exactly 6 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := BuildEnterBootloaderCmd(tt.key)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errMsg)) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Validate frame structure
			if frame[0] != StartOfPacket {
				t.Errorf("SOP = 0x%02X, want 0x%02X", frame[0], StartOfPacket)
			}

			if frame[1] != CmdEnterBootloader {
				t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdEnterBootloader)
			}

			if frame[len(frame)-1] != EndOfPacket {
				t.Errorf("EOP = 0x%02X, want 0x%02X", frame[len(frame)-1], EndOfPacket)
			}

			// Verify key is in frame
			keyStart := 4 // SOP(1) + CMD(1) + LEN(2)
			if !bytes.Equal(frame[keyStart:keyStart+6], tt.key) {
				t.Errorf("key in frame = %v, want %v", frame[keyStart:keyStart+6], tt.key)
			}
		})
	}
}

func TestBuildGetFlashSizeCmd(t *testing.T) {
	tests := []struct {
		name    string
		arrayID byte
	}{
		{name: "array 0", arrayID: 0},
		{name: "array 1", arrayID: 1},
		{name: "array 255", arrayID: 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := BuildGetFlashSizeCmd(tt.arrayID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if frame[0] != StartOfPacket {
				t.Errorf("SOP = 0x%02X, want 0x%02X", frame[0], StartOfPacket)
			}

			if frame[1] != CmdGetFlashSize {
				t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdGetFlashSize)
			}

			if frame[4] != tt.arrayID {
				t.Errorf("ArrayID = 0x%02X, want 0x%02X", frame[4], tt.arrayID)
			}

			if frame[len(frame)-1] != EndOfPacket {
				t.Errorf("EOP = 0x%02X, want 0x%02X", frame[len(frame)-1], EndOfPacket)
			}
		})
	}
}

func TestBuildProgramRowCmd(t *testing.T) {
	tests := []struct {
		name    string
		arrayID byte
		rowNum  uint16
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid small row",
			arrayID: 0,
			rowNum:  0,
			data:    []byte{0x01, 0x02, 0x03, 0x04},
			wantErr: false,
		},
		{
			name:    "valid large row",
			arrayID: 1,
			rowNum:  256,
			data:    make([]byte, 128),
			wantErr: false,
		},
		{
			name:    "empty data",
			arrayID: 0,
			rowNum:  0,
			data:    []byte{},
			wantErr: true,
			errMsg:  "data cannot be empty",
		},
		{
			name:    "nil data",
			arrayID: 0,
			rowNum:  0,
			data:    nil,
			wantErr: true,
			errMsg:  "data cannot be empty",
		},
		{
			name:    "data too large",
			arrayID: 0,
			rowNum:  0,
			data:    make([]byte, 257),
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := BuildProgramRowCmd(tt.arrayID, tt.rowNum, tt.data)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errMsg)) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if frame[0] != StartOfPacket {
				t.Errorf("SOP = 0x%02X, want 0x%02X", frame[0], StartOfPacket)
			}

			if frame[1] != CmdProgramRow {
				t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdProgramRow)
			}

			if frame[4] != tt.arrayID {
				t.Errorf("ArrayID = 0x%02X, want 0x%02X", frame[4], tt.arrayID)
			}

			if frame[len(frame)-1] != EndOfPacket {
				t.Errorf("EOP = 0x%02X, want 0x%02X", frame[len(frame)-1], EndOfPacket)
			}
		})
	}
}

func TestBuildSendDataCmd(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid data",
			data:    []byte{0x01, 0x02, 0x03},
			wantErr: false,
		},
		{
			name:    "max size data",
			data:    make([]byte, 256),
			wantErr: false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "data cannot be empty",
		},
		{
			name:    "data too large",
			data:    make([]byte, 257),
			wantErr: true,
			errMsg:  "exceeds maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := BuildSendDataCmd(tt.data)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !bytes.Contains([]byte(err.Error()), []byte(tt.errMsg)) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if frame[1] != CmdSendData {
				t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdSendData)
			}
		})
	}
}

func TestBuildVerifyRowCmd(t *testing.T) {
	frame, err := BuildVerifyRowCmd(0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame[0] != StartOfPacket {
		t.Errorf("SOP = 0x%02X, want 0x%02X", frame[0], StartOfPacket)
	}

	if frame[1] != CmdVerifyRow {
		t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdVerifyRow)
	}

	if frame[len(frame)-1] != EndOfPacket {
		t.Errorf("EOP = 0x%02X, want 0x%02X", frame[len(frame)-1], EndOfPacket)
	}
}

func TestBuildVerifyChecksumCmd(t *testing.T) {
	frame, err := BuildVerifyChecksumCmd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame[0] != StartOfPacket {
		t.Errorf("SOP = 0x%02X, want 0x%02X", frame[0], StartOfPacket)
	}

	if frame[1] != CmdVerifyChecksum {
		t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdVerifyChecksum)
	}

	if frame[len(frame)-1] != EndOfPacket {
		t.Errorf("EOP = 0x%02X, want 0x%02X", frame[len(frame)-1], EndOfPacket)
	}
}

func TestBuildEraseRowCmd(t *testing.T) {
	frame, err := BuildEraseRowCmd(0, 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame[1] != CmdEraseRow {
		t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdEraseRow)
	}
}

func TestBuildSyncBootloaderCmd(t *testing.T) {
	frame, err := BuildSyncBootloaderCmd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame[1] != CmdSyncBootloader {
		t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdSyncBootloader)
	}
}

func TestBuildExitBootloaderCmd(t *testing.T) {
	frame, err := BuildExitBootloaderCmd()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if frame[1] != CmdExitBootloader {
		t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdExitBootloader)
	}
}

func TestBuildGetMetadataCmd(t *testing.T) {
	tests := []struct {
		name   string
		appNum byte
	}{
		{name: "app 0", appNum: 0},
		{name: "app 1", appNum: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame, err := BuildGetMetadataCmd(tt.appNum)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if frame[1] != CmdGetMetadata {
				t.Errorf("CMD = 0x%02X, want 0x%02X", frame[1], CmdGetMetadata)
			}

			if frame[4] != tt.appNum {
				t.Errorf("AppNum = 0x%02X, want 0x%02X", frame[4], tt.appNum)
			}
		})
	}
}

func BenchmarkBuildEnterBootloaderCmd(b *testing.B) {
	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildEnterBootloaderCmd(key)
	}
}

func BenchmarkBuildProgramRowCmd(b *testing.B) {
	data := make([]byte, 128)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildProgramRowCmd(0, 0, data)
	}
}
