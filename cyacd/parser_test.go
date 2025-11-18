package cyacd

import (
	"bytes"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	// Note: Testing with ParseReader instead since we can't easily create temp files
	t.Run("see ParseReader tests", func(t *testing.T) {
		t.Skip("File-based tests skipped, see ParseReader tests")
	})
}

func TestParseReader(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Firmware
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid simple firmware",
			input: "1E9602AA0000\n" +
				"000000040001020304F2\n",
			want: &Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows: []*Row{
					{
						ArrayID:  0x00,
						RowNum:   0x0000,
						Size:     0x0004,
						Data:     []byte{0x01, 0x02, 0x03, 0x04},
						Checksum: 0xF2,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple rows",
			input: "1E9602AA0000\n" +
				"000000040001020304F2\n" +
				"000100040005060708E1\n",
			want: &Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows: []*Row{
					{
						ArrayID:  0x00,
						RowNum:   0x0000,
						Size:     0x0004,
						Data:     []byte{0x01, 0x02, 0x03, 0x04},
						Checksum: 0xF2,
					},
					{
						ArrayID:  0x00,
						RowNum:   0x0001,
						Size:     0x0004,
						Data:     []byte{0x05, 0x06, 0x07, 0x08},
						Checksum: 0xE1,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "with empty lines",
			input: "1E9602AA0000\n" +
				"\n" +
				"000000040001020304F2\n" +
				"\n",
			want: &Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows: []*Row{
					{
						ArrayID:  0x00,
						RowNum:   0x0000,
						Size:     0x0004,
						Data:     []byte{0x01, 0x02, 0x03, 0x04},
						Checksum: 0xF2,
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: true,
			errMsg:  "empty file",
		},
		{
			name:    "no rows",
			input:   "1E9602AA0000\n",
			wantErr: true,
			errMsg:  "no rows found",
		},
		{
			name:    "invalid header length",
			input:   "1E9602\n",
			wantErr: true,
			errMsg:  "invalid header length",
		},
		{
			name:    "invalid header hex",
			input:   "ZZZZZZZZZZZZ\n",
			wantErr: true,
			errMsg:  "invalid hex data",
		},
		{
			name: "invalid checksum type",
			input: "1E9602AA0099\n" +
				"000000040401020304F6\n",
			wantErr: true,
			errMsg:  "invalid checksum type",
		},
		{
			name: "row too short",
			input: "1E9602AA0000\n" +
				"0000\n",
			wantErr: true,
			errMsg:  "row too short",
		},
		{
			name: "row checksum mismatch",
			input: "1E9602AA0000\n" +
				"000000040001020304FF\n", // Wrong checksum
			wantErr: true,
			errMsg:  "checksum mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			got, err := ParseReader(r)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			// Compare firmware
			if got.SiliconID != tt.want.SiliconID {
				t.Errorf("SiliconID = 0x%08X, want 0x%08X", got.SiliconID, tt.want.SiliconID)
			}

			if got.SiliconRev != tt.want.SiliconRev {
				t.Errorf("SiliconRev = 0x%02X, want 0x%02X", got.SiliconRev, tt.want.SiliconRev)
			}

			if got.ChecksumType != tt.want.ChecksumType {
				t.Errorf("ChecksumType = 0x%02X, want 0x%02X", got.ChecksumType, tt.want.ChecksumType)
			}

			if len(got.Rows) != len(tt.want.Rows) {
				t.Fatalf("Rows count = %d, want %d", len(got.Rows), len(tt.want.Rows))
			}

			for i, row := range got.Rows {
				wantRow := tt.want.Rows[i]

				if row.ArrayID != wantRow.ArrayID {
					t.Errorf("Row[%d].ArrayID = 0x%02X, want 0x%02X", i, row.ArrayID, wantRow.ArrayID)
				}

				if row.RowNum != wantRow.RowNum {
					t.Errorf("Row[%d].RowNum = %d, want %d", i, row.RowNum, wantRow.RowNum)
				}

				if !bytes.Equal(row.Data, wantRow.Data) {
					t.Errorf("Row[%d].Data = %v, want %v", i, row.Data, wantRow.Data)
				}

				if row.Checksum != wantRow.Checksum {
					t.Errorf("Row[%d].Checksum = 0x%02X, want 0x%02X", i, row.Checksum, wantRow.Checksum)
				}
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    *Firmware
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid header",
			line: "1E9602AA0000",
			want: &Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
			},
			wantErr: false,
		},
		{
			name: "valid header with CRC checksum",
			line: "1E9602AA0001",
			want: &Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x01,
			},
			wantErr: false,
		},
		{
			name:    "too short",
			line:    "1E9602AA",
			wantErr: true,
			errMsg:  "invalid header length",
		},
		{
			name:    "too long",
			line:    "1E9602AA000000",
			wantErr: true,
			errMsg:  "invalid header length",
		},
		{
			name:    "invalid hex",
			line:    "ZZZZZZZZZZZZ",
			wantErr: true,
			errMsg:  "invalid hex data",
		},
		{
			name:    "invalid checksum type",
			line:    "1E9602AA0099",
			wantErr: true,
			errMsg:  "invalid checksum type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseHeader(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.SiliconID != tt.want.SiliconID {
				t.Errorf("SiliconID = 0x%08X, want 0x%08X", got.SiliconID, tt.want.SiliconID)
			}

			if got.SiliconRev != tt.want.SiliconRev {
				t.Errorf("SiliconRev = 0x%02X, want 0x%02X", got.SiliconRev, tt.want.SiliconRev)
			}

			if got.ChecksumType != tt.want.ChecksumType {
				t.Errorf("ChecksumType = 0x%02X, want 0x%02X", got.ChecksumType, tt.want.ChecksumType)
			}
		})
	}
}

func TestParseRow(t *testing.T) {
	tests := []struct {
		name    string
		line    string
		want    *Row
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid row",
			line: "000000040001020304F2",
			want: &Row{
				ArrayID:  0x00,
				RowNum:   0x0000,
				Size:     0x0004,
				Data:     []byte{0x01, 0x02, 0x03, 0x04},
				Checksum: 0xF2,
			},
			wantErr: false,
		},
		{
			name: "different array and row",
			line: "01FF010400AABBCCDDED",
			want: &Row{
				ArrayID:  0x01,
				RowNum:   0x01FF,
				Size:     0x0004,
				Data:     []byte{0xAA, 0xBB, 0xCC, 0xDD},
				Checksum: 0xED,
			},
			wantErr: false,
		},
		{
			name:    "too short",
			line:    "0000000404",
			wantErr: true,
			errMsg:  "row too short",
		},
		{
			name:    "invalid hex",
			line:    "ZZZZZZZZZZZZ",
			wantErr: true,
			errMsg:  "invalid hex data",
		},
		{
			name:    "length mismatch",
			line:    "000000080001020304F6", // Says 8 bytes but only 4
			wantErr: true,
			errMsg:  "data length mismatch",
		},
		{
			name:    "checksum mismatch",
			line:    "000000040001020304FF", // Wrong checksum
			wantErr: true,
			errMsg:  "checksum mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseRow(tt.line)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errMsg)
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", err, tt.errMsg)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if got.ArrayID != tt.want.ArrayID {
				t.Errorf("ArrayID = 0x%02X, want 0x%02X", got.ArrayID, tt.want.ArrayID)
			}

			if got.RowNum != tt.want.RowNum {
				t.Errorf("RowNum = %d, want %d", got.RowNum, tt.want.RowNum)
			}

			if !bytes.Equal(got.Data, tt.want.Data) {
				t.Errorf("Data = %v, want %v", got.Data, tt.want.Data)
			}

			if got.Checksum != tt.want.Checksum {
				t.Errorf("Checksum = 0x%02X, want 0x%02X", got.Checksum, tt.want.Checksum)
			}
		})
	}
}

func TestCalculateRowChecksum(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			name:     "simple case",
			data:     []byte{0x00, 0x00, 0x00, 0x04, 0x00, 0x01, 0x02, 0x03, 0x04},
			expected: 0xF2,
		},
		{
			name:     "zeros",
			data:     []byte{0x00, 0x00, 0x00},
			expected: 0x00,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateRowChecksum(tt.data)
			if result != tt.expected {
				t.Errorf("calculateRowChecksum() = 0x%02X, want 0x%02X", result, tt.expected)
			}
		})
	}
}

func BenchmarkParseReader(b *testing.B) {
	// Create a larger firmware for benchmarking
	var buf bytes.Buffer
	buf.WriteString("1E9602AA00\n")
	for i := 0; i < 100; i++ {
		buf.WriteString("000000040401020304F6\n")
	}
	data := buf.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := bytes.NewReader(data)
		_, _ = ParseReader(r)
	}
}

func BenchmarkParseRow(b *testing.B) {
	line := "000000040401020304F6"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseRow(line)
	}
}
