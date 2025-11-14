package protocol

import "testing"

func TestCalculateRowChecksum(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected byte
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: 0x01, // 2's complement of 0
		},
		{
			name:     "single byte",
			data:     []byte{0x01},
			expected: 0xFF, // 2's complement of 0x01
		},
		{
			name:     "multiple bytes",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			expected: 0xF6, // 2's complement of 0x0A
		},
		{
			name:     "all zeros",
			data:     []byte{0x00, 0x00, 0x00, 0x00},
			expected: 0x00,
		},
		{
			name:     "all ones",
			data:     []byte{0xFF, 0xFF, 0xFF, 0xFF},
			expected: 0x04, // overflow and 2's complement
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateRowChecksum(tt.data)
			if result != tt.expected {
				t.Errorf("CalculateRowChecksum() = 0x%02X, want 0x%02X", result, tt.expected)
			}
		})
	}
}

func TestCalculatePacketChecksum(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: 0x0001, // 2's complement of 0
		},
		{
			name:     "single byte",
			data:     []byte{0x38}, // Enter Bootloader command
			expected: 0xFFC8,       // 2's complement
		},
		{
			name: "command with length",
			data: []byte{0x38, 0x06, 0x00}, // Enter Bootloader, len=6
			expected: 0xFFC2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculatePacketChecksum(tt.data)
			if result != tt.expected {
				t.Errorf("calculatePacketChecksum() = 0x%04X, want 0x%04X", result, tt.expected)
			}
		})
	}
}

func TestCalculateCRC16(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected uint16
	}{
		{
			name:     "empty data",
			data:     []byte{},
			expected: 0xFFFF,
		},
		{
			name:     "single byte zero",
			data:     []byte{0x00},
			expected: 0x1021,
		},
		{
			name:     "test data",
			data:     []byte{0x01, 0x02, 0x03, 0x04},
			expected: 0x89C3, // CRC-16-CCITT
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCRC16(tt.data)
			if result != tt.expected {
				t.Errorf("calculateCRC16() = 0x%04X, want 0x%04X", result, tt.expected)
			}
		})
	}
}

func BenchmarkCalculateRowChecksum(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CalculateRowChecksum(data)
	}
}

func BenchmarkCalculatePacketChecksum(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculatePacketChecksum(data)
	}
}

func BenchmarkCalculateCRC16(b *testing.B) {
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		calculateCRC16(data)
	}
}
