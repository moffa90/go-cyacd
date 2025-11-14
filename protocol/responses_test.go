package protocol

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// Helper function to build a valid response frame for testing
func buildTestResponse(statusCode byte, data []byte) []byte {
	dataLen := uint16(len(data))
	frame := make([]byte, 0, MinFrameSize+len(data))

	frame = append(frame, StartOfPacket)
	frame = append(frame, statusCode)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, data...)

	checksum := calculatePacketChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, EndOfPacket)

	return frame
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name           string
		frame          []byte
		wantStatusCode byte
		wantDataLen    int
		wantErr        bool
		errMsg         string
	}{
		{
			name:           "valid response with no data",
			frame:          buildTestResponse(StatusSuccess, nil),
			wantStatusCode: StatusSuccess,
			wantDataLen:    0,
			wantErr:        false,
		},
		{
			name:           "valid response with data",
			frame:          buildTestResponse(StatusSuccess, []byte{0x01, 0x02, 0x03}),
			wantStatusCode: StatusSuccess,
			wantDataLen:    3,
			wantErr:        false,
		},
		{
			name:           "error status code",
			frame:          buildTestResponse(ErrChecksum, nil),
			wantStatusCode: ErrChecksum,
			wantDataLen:    0,
			wantErr:        false,
		},
		{
			name:    "frame too short",
			frame:   []byte{0x01, 0x00},
			wantErr: true,
			errMsg:  "frame too short",
		},
		{
			name:    "invalid start of packet",
			frame:   []byte{0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x17},
			wantErr: true,
			errMsg:  "invalid start of packet",
		},
		{
			name:    "invalid end of packet",
			frame:   []byte{0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF},
			wantErr: true,
			errMsg:  "invalid end of packet",
		},
		{
			name: "checksum mismatch",
			frame: []byte{
				StartOfPacket,
				StatusSuccess,
				0x00, 0x00, // length
				0xFF, 0xFF, // wrong checksum
				EndOfPacket,
			},
			wantErr: true,
			errMsg:  "checksum mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			statusCode, data, err := ParseResponse(tt.frame)

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

			if statusCode != tt.wantStatusCode {
				t.Errorf("statusCode = 0x%02X, want 0x%02X", statusCode, tt.wantStatusCode)
			}

			if len(data) != tt.wantDataLen {
				t.Errorf("data length = %d, want %d", len(data), tt.wantDataLen)
			}
		})
	}
}

func TestParseEnterBootloaderResponse(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantInfo  *DeviceInfo
		wantErr   bool
		errMsg    string
	}{
		{
			name: "valid response",
			data: []byte{
				0xAA, 0x02, 0x96, 0x1E, // Silicon ID (little-endian)
				0x00,                   // Silicon Rev
				0x01, 0x1E, 0x00,       // Bootloader Ver
			},
			wantInfo: &DeviceInfo{
				SiliconID:     0x1E9602AA,
				SiliconRev:    0x00,
				BootloaderVer: [3]byte{0x01, 0x1E, 0x00},
			},
			wantErr: false,
		},
		{
			name:    "data too short",
			data:    []byte{0x01, 0x02},
			wantErr: true,
			errMsg:  "invalid data length",
		},
		{
			name:    "data too long",
			data:    make([]byte, 10),
			wantErr: true,
			errMsg:  "invalid data length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseEnterBootloaderResponse(tt.data)

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

			if info.SiliconID != tt.wantInfo.SiliconID {
				t.Errorf("SiliconID = 0x%08X, want 0x%08X", info.SiliconID, tt.wantInfo.SiliconID)
			}

			if info.SiliconRev != tt.wantInfo.SiliconRev {
				t.Errorf("SiliconRev = 0x%02X, want 0x%02X", info.SiliconRev, tt.wantInfo.SiliconRev)
			}

			if info.BootloaderVer != tt.wantInfo.BootloaderVer {
				t.Errorf("BootloaderVer = %v, want %v", info.BootloaderVer, tt.wantInfo.BootloaderVer)
			}
		})
	}
}

func TestParseGetFlashSizeResponse(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		wantSize *FlashSize
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid response",
			data: []byte{
				0x00, 0x00, // StartRow (little-endian)
				0xFF, 0x01, // EndRow (little-endian)
			},
			wantSize: &FlashSize{
				StartRow: 0x0000,
				EndRow:   0x01FF,
			},
			wantErr: false,
		},
		{
			name:    "data too short",
			data:    []byte{0x01, 0x02},
			wantErr: true,
			errMsg:  "invalid data length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			size, err := ParseGetFlashSizeResponse(tt.data)

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

			if size.StartRow != tt.wantSize.StartRow {
				t.Errorf("StartRow = %d, want %d", size.StartRow, tt.wantSize.StartRow)
			}

			if size.EndRow != tt.wantSize.EndRow {
				t.Errorf("EndRow = %d, want %d", size.EndRow, tt.wantSize.EndRow)
			}
		})
	}
}

func TestParseVerifyRowResponse(t *testing.T) {
	tests := []struct {
		name         string
		data         []byte
		wantChecksum byte
		wantErr      bool
		errMsg       string
	}{
		{
			name:         "valid checksum",
			data:         []byte{0xAB},
			wantChecksum: 0xAB,
			wantErr:      false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "invalid data length",
		},
		{
			name:    "data too long",
			data:    []byte{0x01, 0x02},
			wantErr: true,
			errMsg:  "invalid data length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checksum, err := ParseVerifyRowResponse(tt.data)

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

			if checksum != tt.wantChecksum {
				t.Errorf("checksum = 0x%02X, want 0x%02X", checksum, tt.wantChecksum)
			}
		})
	}
}

func TestParseVerifyChecksumResponse(t *testing.T) {
	tests := []struct {
		name      string
		data      []byte
		wantValid bool
		wantErr   bool
		errMsg    string
	}{
		{
			name:      "valid checksum",
			data:      []byte{0x01},
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "invalid checksum",
			data:      []byte{0x00},
			wantValid: false,
			wantErr:   false,
		},
		{
			name:    "empty data",
			data:    []byte{},
			wantErr: true,
			errMsg:  "invalid data length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := ParseVerifyChecksumResponse(tt.data)

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

			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestParseGetAppStatusResponse(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		wantStatus *AppStatus
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid and active",
			data: []byte{0x01, 0x01},
			wantStatus: &AppStatus{
				Valid:  true,
				Active: true,
			},
			wantErr: false,
		},
		{
			name: "valid but not active",
			data: []byte{0x01, 0x00},
			wantStatus: &AppStatus{
				Valid:  true,
				Active: false,
			},
			wantErr: false,
		},
		{
			name:    "data too short",
			data:    []byte{0x01},
			wantErr: true,
			errMsg:  "invalid data length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := ParseGetAppStatusResponse(tt.data)

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

			if status.Valid != tt.wantStatus.Valid {
				t.Errorf("Valid = %v, want %v", status.Valid, tt.wantStatus.Valid)
			}

			if status.Active != tt.wantStatus.Active {
				t.Errorf("Active = %v, want %v", status.Active, tt.wantStatus.Active)
			}
		})
	}
}

func BenchmarkParseResponse(b *testing.B) {
	frame := buildTestResponse(StatusSuccess, []byte{0x01, 0x02, 0x03, 0x04})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = ParseResponse(frame)
	}
}
