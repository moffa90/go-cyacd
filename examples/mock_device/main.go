package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/moffa90/go-cyacd/bootloader"
	"github.com/moffa90/go-cyacd/cyacd"
	"github.com/moffa90/go-cyacd/protocol"
)

// RealisticMockDevice simulates a real Cypress bootloader device
// It validates commands and generates proper responses
type RealisticMockDevice struct {
	siliconID    uint32
	siliconRev   byte
	flashStart   uint16
	flashEnd     uint16
	flash        map[uint16][]byte // Simulated flash memory
	inBootloader bool
	latency      time.Duration
}

func NewRealisticMockDevice() *RealisticMockDevice {
	return &RealisticMockDevice{
		siliconID:    0x1E9602AA,
		siliconRev:   0x00,
		flashStart:   0x0000,
		flashEnd:     0x01FF,
		flash:        make(map[uint16][]byte),
		inBootloader: false,
		latency:      10 * time.Millisecond,
	}
}

func (d *RealisticMockDevice) Read(p []byte) (int, error) {
	// Simulate device response time
	time.Sleep(d.latency)

	// This would be implemented by your actual hardware
	// For now, just return empty data
	return len(p), nil
}

func (d *RealisticMockDevice) Write(p []byte) (int, error) {
	// Simulate device write time
	time.Sleep(d.latency / 2)

	// Parse the command frame
	if len(p) < protocol.MinFrameSize {
		return 0, fmt.Errorf("frame too short")
	}

	cmd := p[1] // Command byte

	// Generate appropriate response based on command
	var response []byte
	var err error

	switch cmd {
	case protocol.CmdEnterBootloader:
		response = d.handleEnterBootloader()
	case protocol.CmdGetFlashSize:
		response = d.handleGetFlashSize(p)
	case protocol.CmdProgramRow:
		response = d.handleProgramRow(p)
	case protocol.CmdVerifyRow:
		response = d.handleVerifyRow(p)
	case protocol.CmdVerifyChecksum:
		response = d.handleVerifyChecksum()
	case protocol.CmdExitBootloader:
		response = d.handleExitBootloader()
	case protocol.CmdSendData:
		response = d.handleSendData(p)
	default:
		response = buildResponseFrame(protocol.ErrCmd, nil)
	}

	// In a real device, this response would be sent back via Read()
	// For this mock, we just log it
	if err == nil && response != nil {
		fmt.Printf("[DEVICE] Sent response: cmd=0x%02X, status=0x%02X, len=%d\n",
			cmd, response[1], len(response))
	}

	return len(p), nil
}

func (d *RealisticMockDevice) handleEnterBootloader() []byte {
	d.inBootloader = true

	data := make([]byte, 8)
	binary.LittleEndian.PutUint32(data[0:4], d.siliconID)
	data[4] = d.siliconRev
	data[5] = 0x01 // Bootloader version
	data[6] = 0x1E
	data[7] = 0x00

	fmt.Println("[DEVICE] Entered bootloader mode")
	return buildResponseFrame(protocol.StatusSuccess, data)
}

func (d *RealisticMockDevice) handleGetFlashSize(frame []byte) []byte {
	if !d.inBootloader {
		return buildResponseFrame(protocol.ErrActive, nil)
	}

	data := make([]byte, 4)
	binary.LittleEndian.PutUint16(data[0:2], d.flashStart)
	binary.LittleEndian.PutUint16(data[2:4], d.flashEnd)

	fmt.Printf("[DEVICE] Flash size: 0x%04X - 0x%04X\n", d.flashStart, d.flashEnd)
	return buildResponseFrame(protocol.StatusSuccess, data)
}

func (d *RealisticMockDevice) handleProgramRow(frame []byte) []byte {
	if !d.inBootloader {
		return buildResponseFrame(protocol.ErrActive, nil)
	}

	// Parse row data from frame
	// Format: [arrayID(1)][rowNum(2)][data(N)]
	if len(frame) < 11 { // Min: SOP+CMD+LEN+arrayID+rowNum+CHK+EOP
		return buildResponseFrame(protocol.ErrLength, nil)
	}

	dataStart := 4
	arrayID := frame[dataStart]
	rowNum := binary.LittleEndian.Uint16(frame[dataStart+1 : dataStart+3])
	dataLen := binary.LittleEndian.Uint16(frame[2:4])
	rowData := frame[dataStart+3 : dataStart+int(dataLen)]

	// Validate row is in range
	if rowNum < d.flashStart || rowNum > d.flashEnd {
		fmt.Printf("[DEVICE] Row out of range: %d\n", rowNum)
		return buildResponseFrame(protocol.ErrRow, nil)
	}

	// Store in simulated flash
	d.flash[rowNum] = make([]byte, len(rowData))
	copy(d.flash[rowNum], rowData)

	fmt.Printf("[DEVICE] Programmed row %d (array %d): %d bytes\n",
		rowNum, arrayID, len(rowData))
	return buildResponseFrame(protocol.StatusSuccess, nil)
}

func (d *RealisticMockDevice) handleVerifyRow(frame []byte) []byte {
	if !d.inBootloader {
		return buildResponseFrame(protocol.ErrActive, nil)
	}

	dataStart := 4
	rowNum := binary.LittleEndian.Uint16(frame[dataStart+1 : dataStart+3])

	// Get row data from flash
	rowData, exists := d.flash[rowNum]
	if !exists {
		return buildResponseFrame(protocol.ErrRow, nil)
	}

	// Calculate checksum
	checksum := protocol.CalculateRowChecksum(rowData)

	fmt.Printf("[DEVICE] Verified row %d: checksum=0x%02X\n", rowNum, checksum)
	return buildResponseFrame(protocol.StatusSuccess, []byte{checksum})
}

func (d *RealisticMockDevice) handleVerifyChecksum() []byte {
	if !d.inBootloader {
		return buildResponseFrame(protocol.ErrActive, nil)
	}

	// Simulate valid checksum
	fmt.Println("[DEVICE] Application checksum valid")
	return buildResponseFrame(protocol.StatusSuccess, []byte{0x01})
}

func (d *RealisticMockDevice) handleExitBootloader() []byte {
	d.inBootloader = false
	fmt.Println("[DEVICE] Exited bootloader mode")
	return buildResponseFrame(protocol.StatusSuccess, nil)
}

func (d *RealisticMockDevice) handleSendData(frame []byte) []byte {
	if !d.inBootloader {
		return buildResponseFrame(protocol.ErrActive, nil)
	}

	dataLen := binary.LittleEndian.Uint16(frame[2:4])
	fmt.Printf("[DEVICE] Received data chunk: %d bytes\n", dataLen)
	return buildResponseFrame(protocol.StatusSuccess, nil)
}

// Helper function to build response frames
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

func main() {
	fmt.Println("=== Cypress Bootloader - Mock Device Example ===\n")
	fmt.Println("This example demonstrates how to implement a realistic mock device")
	fmt.Println("that simulates actual bootloader behavior for testing.\n")

	// Create a simple test firmware
	firmware := &cyacd.Firmware{
		SiliconID:    0x1E9602AA,
		SiliconRev:   0x00,
		ChecksumType: 0x00,
		Rows: []*cyacd.Row{
			{
				ArrayID:  0x00,
				RowNum:   0x0000,
				Data:     []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
				Checksum: protocol.CalculateRowChecksum([]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}),
			},
			{
				ArrayID:  0x00,
				RowNum:   0x0001,
				Data:     []byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18},
				Checksum: protocol.CalculateRowChecksum([]byte{0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18}),
			},
		},
	}

	// Create mock device
	device := NewRealisticMockDevice()

	fmt.Println("Mock Device Configuration:")
	fmt.Printf("  Silicon ID:   0x%08X\n", device.siliconID)
	fmt.Printf("  Silicon Rev:  0x%02X\n", device.siliconRev)
	fmt.Printf("  Flash Range:  0x%04X - 0x%04X\n", device.flashStart, device.flashEnd)
	fmt.Printf("  Latency:      %s\n\n", device.latency)

	// Create programmer
	prog := bootloader.New(device,
		bootloader.WithProgressCallback(func(p bootloader.Progress) {
			fmt.Printf("[PROGRAMMER] %s: %.1f%% (Row %d/%d)\n",
				p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
		}),
	)

	// Program the device
	fmt.Println("Starting programming with mock device...\n")
	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}

	err := prog.Program(context.Background(), firmware, key)
	if err != nil {
		log.Fatalf("Programming failed: %v", err)
	}

	fmt.Println("\n✅ Programming completed successfully!")
	fmt.Println("\nMock Device State:")
	fmt.Printf("  In Bootloader: %t\n", device.inBootloader)
	fmt.Printf("  Flash Rows:    %d\n", len(device.flash))

	for rowNum, data := range device.flash {
		fmt.Printf("    Row 0x%04X:  %d bytes: % 02X\n", rowNum, len(data), data)
	}

	fmt.Println("\nThis mock device can be used for:")
	fmt.Println("  - Testing your bootloader implementation")
	fmt.Println("  - Developing without hardware")
	fmt.Println("  - Automated testing in CI/CD")
	fmt.Println("  - Simulating error conditions")
}
