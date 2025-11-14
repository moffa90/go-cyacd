package bootloader

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/moffa90/go-cyacd/cyacd"
	"github.com/moffa90/go-cyacd/protocol"
)

// MockDevice simulates a bootloader device for testing
type MockDevice struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	readErr  error
	writeErr error
	responses [][]byte
	respIdx  int
}

func NewMockDevice() *MockDevice {
	return &MockDevice{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
		responses: make([][]byte, 0),
	}
}

func (m *MockDevice) Read(p []byte) (int, error) {
	if m.readErr != nil {
		return 0, m.readErr
	}

	if m.respIdx < len(m.responses) {
		resp := m.responses[m.respIdx]
		m.respIdx++
		copy(p, resp)
		return len(resp), nil
	}

	return m.readBuf.Read(p)
}

func (m *MockDevice) Write(p []byte) (int, error) {
	if m.writeErr != nil {
		return 0, m.writeErr
	}
	return m.writeBuf.Write(p)
}

func (m *MockDevice) AddResponse(statusCode byte, data []byte) {
	frame := buildResponseFrame(statusCode, data)
	m.responses = append(m.responses, frame)
}

func (m *MockDevice) SetReadError(err error) {
	m.readErr = err
}

func (m *MockDevice) SetWriteError(err error) {
	m.writeErr = err
}

func (m *MockDevice) Reset() {
	m.readBuf.Reset()
	m.writeBuf.Reset()
	m.responses = make([][]byte, 0)
	m.respIdx = 0
	m.readErr = nil
	m.writeErr = nil
}

// Helper function to build a response frame
func buildResponseFrame(statusCode byte, data []byte) []byte {
	dataLen := uint16(len(data))
	frame := make([]byte, 0, protocol.MinFrameSize+len(data))

	frame = append(frame, protocol.StartOfPacket)
	frame = append(frame, statusCode)

	lenBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(lenBytes, dataLen)
	frame = append(frame, lenBytes...)

	frame = append(frame, data...)

	checksum := calculateChecksum(frame[1:])
	checksumBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(checksumBytes, checksum)
	frame = append(frame, checksumBytes...)

	frame = append(frame, protocol.EndOfPacket)

	return frame
}

func calculateChecksum(data []byte) uint16 {
	var sum uint16
	for _, b := range data {
		sum += uint16(b)
	}
	return 1 + (0xFFFF ^ sum)
}

// Mock logger for testing
type MockLogger struct {
	debugMsgs []string
	infoMsgs  []string
	errorMsgs []string
}

func (l *MockLogger) Debug(msg string, kv ...interface{}) {
	l.debugMsgs = append(l.debugMsgs, msg)
}

func (l *MockLogger) Info(msg string, kv ...interface{}) {
	l.infoMsgs = append(l.infoMsgs, msg)
}

func (l *MockLogger) Error(msg string, kv ...interface{}) {
	l.errorMsgs = append(l.errorMsgs, msg)
}

func TestNew(t *testing.T) {
	device := NewMockDevice()

	tests := []struct {
		name    string
		device  io.ReadWriter
		options []Option
	}{
		{
			name:    "with no options",
			device:  device,
			options: nil,
		},
		{
			name:   "with all options",
			device: device,
			options: []Option{
				WithProgressCallback(func(p Progress) {}),
				WithLogger(&MockLogger{}),
				WithTimeout(30 * time.Second),
				WithChunkSize(64),
				WithRetries(5),
				WithVerifyAfterProgram(false),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog := New(tt.device, tt.options...)
			if prog == nil {
				t.Fatal("New() returned nil")
			}
			if prog.device != tt.device {
				t.Error("device not set correctly")
			}
		})
	}
}

func TestEnterBootloader(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		statusCode byte
		deviceInfo []byte
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "successful enter",
			key:        []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			statusCode: protocol.StatusSuccess,
			deviceInfo: []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00},
			wantErr:    false,
		},
		{
			name:       "bootloader error",
			key:        []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			statusCode: protocol.ErrKey,
			deviceInfo: nil,
			wantErr:    true,
			errMsg:     "CYRET_ERR_KEY",
		},
		{
			name:    "invalid key length",
			key:     []byte{0x0A, 0x1B},
			wantErr: true,
			errMsg:  "key must be exactly 6 bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewMockDevice()
			if tt.statusCode != 0 {
				device.AddResponse(tt.statusCode, tt.deviceInfo)
			}

			prog := New(device)
			info, err := prog.EnterBootloader(context.Background(), tt.key)

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

			if info.SiliconID != 0x1E9602AA {
				t.Errorf("SiliconID = 0x%08X, want 0x1E9602AA", info.SiliconID)
			}
		})
	}
}

func TestExitBootloader(t *testing.T) {
	device := NewMockDevice()
	device.AddResponse(protocol.StatusSuccess, nil)

	prog := New(device)
	err := prog.ExitBootloader(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetFlashSize(t *testing.T) {
	tests := []struct {
		name       string
		arrayID    byte
		statusCode byte
		data       []byte
		wantStart  uint16
		wantEnd    uint16
		wantErr    bool
	}{
		{
			name:       "successful get flash size",
			arrayID:    0,
			statusCode: protocol.StatusSuccess,
			data:       []byte{0x00, 0x00, 0xFF, 0x01},
			wantStart:  0x0000,
			wantEnd:    0x01FF,
			wantErr:    false,
		},
		{
			name:       "bootloader error",
			arrayID:    0,
			statusCode: protocol.ErrArray,
			data:       nil,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewMockDevice()
			device.AddResponse(tt.statusCode, tt.data)

			prog := New(device)
			size, err := prog.GetFlashSize(context.Background(), tt.arrayID)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if size.StartRow != tt.wantStart {
				t.Errorf("StartRow = %d, want %d", size.StartRow, tt.wantStart)
			}

			if size.EndRow != tt.wantEnd {
				t.Errorf("EndRow = %d, want %d", size.EndRow, tt.wantEnd)
			}
		})
	}
}

func TestVerifyChecksum(t *testing.T) {
	tests := []struct {
		name       string
		statusCode byte
		data       []byte
		wantValid  bool
		wantErr    bool
	}{
		{
			name:       "valid checksum",
			statusCode: protocol.StatusSuccess,
			data:       []byte{0x01},
			wantValid:  true,
			wantErr:    false,
		},
		{
			name:       "invalid checksum",
			statusCode: protocol.StatusSuccess,
			data:       []byte{0x00},
			wantValid:  false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewMockDevice()
			device.AddResponse(tt.statusCode, tt.data)

			prog := New(device)
			valid, err := prog.VerifyChecksum(context.Background())

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
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

func TestProgram(t *testing.T) {
	tests := []struct {
		name        string
		firmware    *cyacd.Firmware
		key         []byte
		setupDevice func(*MockDevice)
		wantErr     bool
		errType     interface{}
	}{
		{
			name: "successful program",
			firmware: &cyacd.Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows: []*cyacd.Row{
					{
						ArrayID:  0x00,
						RowNum:   0x0000,
						Data:     []byte{0x01, 0x02, 0x03, 0x04},
						Checksum: 0xF6,
					},
				},
			},
			key: []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			setupDevice: func(d *MockDevice) {
				// Enter bootloader response
				d.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
				// Get flash size response
				d.AddResponse(protocol.StatusSuccess, []byte{0x00, 0x00, 0xFF, 0x01})
				// Program row response
				d.AddResponse(protocol.StatusSuccess, nil)
				// Verify row response
				d.AddResponse(protocol.StatusSuccess, []byte{0xF6})
				// Verify checksum response
				d.AddResponse(protocol.StatusSuccess, []byte{0x01})
				// Exit bootloader response
				d.AddResponse(protocol.StatusSuccess, nil)
			},
			wantErr: false,
		},
		{
			name: "device mismatch",
			firmware: &cyacd.Firmware{
				SiliconID:    0x12345678,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows:         []*cyacd.Row{},
			},
			key: []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			setupDevice: func(d *MockDevice) {
				d.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
			},
			wantErr: true,
			errType: &DeviceMismatchError{},
		},
		{
			name: "row out of range",
			firmware: &cyacd.Firmware{
				SiliconID:    0x1E9602AA,
				SiliconRev:   0x00,
				ChecksumType: 0x00,
				Rows: []*cyacd.Row{
					{
						ArrayID:  0x00,
						RowNum:   0x0FFF, // Out of range
						Data:     []byte{0x01, 0x02, 0x03, 0x04},
						Checksum: 0xF6,
					},
				},
			},
			key: []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F},
			setupDevice: func(d *MockDevice) {
				d.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
				d.AddResponse(protocol.StatusSuccess, []byte{0x00, 0x00, 0xFF, 0x01})
			},
			wantErr: true,
			errType: &RowOutOfRangeError{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewMockDevice()
			tt.setupDevice(device)

			prog := New(device)
			err := prog.Program(context.Background(), tt.firmware, tt.key)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errType != nil {
					if !errors.As(err, &tt.errType) {
						t.Errorf("error type = %T, want %T", err, tt.errType)
					}
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestProgramWithProgress(t *testing.T) {
	device := NewMockDevice()

	// Setup device responses
	device.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
	device.AddResponse(protocol.StatusSuccess, []byte{0x00, 0x00, 0xFF, 0x01})
	device.AddResponse(protocol.StatusSuccess, nil)
	device.AddResponse(protocol.StatusSuccess, []byte{0xF6})
	device.AddResponse(protocol.StatusSuccess, []byte{0x01})
	device.AddResponse(protocol.StatusSuccess, nil)

	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows: []*cyacd.Row{
			{
				ArrayID:  0x00,
				RowNum:   0x0000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04},
				Checksum: 0xF6,
			},
		},
	}

	var progressCalls []Progress
	prog := New(device, WithProgressCallback(func(p Progress) {
		progressCalls = append(progressCalls, p)
	}))

	err := prog.Program(context.Background(), firmware, []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(progressCalls) == 0 {
		t.Error("expected progress callbacks, got none")
	}

	// Check that we got all phases
	phases := make(map[string]bool)
	for _, p := range progressCalls {
		phases[p.Phase] = true
	}

	expectedPhases := []string{"entering", "programming", "verifying", "exiting", "complete"}
	for _, phase := range expectedPhases {
		if !phases[phase] {
			t.Errorf("missing phase: %s", phase)
		}
	}
}

func TestProgramWithLogging(t *testing.T) {
	device := NewMockDevice()

	// Setup device responses
	device.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
	device.AddResponse(protocol.StatusSuccess, []byte{0x00, 0x00, 0xFF, 0x01})
	device.AddResponse(protocol.StatusSuccess, nil)
	device.AddResponse(protocol.StatusSuccess, []byte{0xF6})
	device.AddResponse(protocol.StatusSuccess, []byte{0x01})
	device.AddResponse(protocol.StatusSuccess, nil)

	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows: []*cyacd.Row{
			{
				ArrayID:  0x00,
				RowNum:   0x0000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04},
				Checksum: 0xF6,
			},
		},
	}

	logger := &MockLogger{}
	prog := New(device, WithLogger(logger))

	err := prog.Program(context.Background(), firmware, []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(logger.infoMsgs) == 0 {
		t.Error("expected info log messages, got none")
	}
}

func TestProgramWithContextCancellation(t *testing.T) {
	device := NewMockDevice()
	device.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})

	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows:         []*cyacd.Row{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	prog := New(device)
	err := prog.Program(ctx, firmware, []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})

	if err == nil {
		t.Fatal("expected context cancellation error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

func TestProgramWithTimeout(t *testing.T) {
	device := NewMockDevice()

	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows:         []*cyacd.Row{},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(10 * time.Millisecond) // Ensure timeout

	prog := New(device)
	err := prog.Program(ctx, firmware, []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("error = %v, want context.DeadlineExceeded", err)
	}
}

func TestReadWriteErrors(t *testing.T) {
	tests := []struct {
		name      string
		setupErr  func(*MockDevice)
		wantError string
	}{
		{
			name: "write error",
			setupErr: func(d *MockDevice) {
				d.SetWriteError(errors.New("write failed"))
			},
			wantError: "write failed",
		},
		{
			name: "read error",
			setupErr: func(d *MockDevice) {
				d.SetReadError(errors.New("read failed"))
			},
			wantError: "read failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := NewMockDevice()
			tt.setupErr(device)

			prog := New(device)
			_, err := prog.EnterBootloader(context.Background(), []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})

			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !bytes.Contains([]byte(err.Error()), []byte(tt.wantError)) {
				t.Errorf("error = %v, want substring %q", err, tt.wantError)
			}
		})
	}
}

func BenchmarkProgram(b *testing.B) {
	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows: []*cyacd.Row{
			{
				ArrayID:  0x00,
				RowNum:   0x0000,
				Data:     make([]byte, 64),
				Checksum: 0x00,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		device := NewMockDevice()
		device.AddResponse(protocol.StatusSuccess, []byte{0xAA, 0x02, 0x96, 0x1E, 0x00, 0x01, 0x1E, 0x00})
		device.AddResponse(protocol.StatusSuccess, []byte{0x00, 0x00, 0xFF, 0x01})
		device.AddResponse(protocol.StatusSuccess, nil)
		device.AddResponse(protocol.StatusSuccess, []byte{0x00})
		device.AddResponse(protocol.StatusSuccess, []byte{0x01})
		device.AddResponse(protocol.StatusSuccess, nil)

		prog := New(device)
		_ = prog.Program(context.Background(), firmware, []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F})
	}
}
