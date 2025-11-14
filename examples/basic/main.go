package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	"github.com/moffa90/go-cyacd/bootloader"
	"github.com/moffa90/go-cyacd/cyacd"
)

// MockDevice is a simple mock implementation of io.ReadWriter for demonstration.
// In real applications, replace this with your actual hardware communication.
type MockDevice struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
}

func NewMockDevice() *MockDevice {
	return &MockDevice{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *MockDevice) Read(p []byte) (int, error) {
	// In a real implementation, read from your hardware
	return m.readBuf.Read(p)
}

func (m *MockDevice) Write(p []byte) (int, error) {
	// In a real implementation, write to your hardware
	return m.writeBuf.Write(p)
}

func main() {
	fmt.Println("go-cyacd Basic Example")
	fmt.Println("======================")
	fmt.Println()

	// Note: This example uses a mock device for demonstration.
	// In a real application, you would open your actual USB/UART/etc device here:
	//
	//   device, err := myusb.OpenDevice("ABC12345")
	//   if err != nil {
	//       log.Fatal(err)
	//   }
	//   defer device.Close()

	device := NewMockDevice()

	// Parse firmware file
	// For this example, we'll create a minimal valid .cyacd content
	cyacdContent := "1E9602AA00\n" + // Header: SiliconID=0x1E9602AA, Rev=0x00, Checksum=Basic
		"000000040401020304F6\n" // Row: ArrayID=0, RowNum=0, Data=[1,2,3,4]

	fw, err := cyacd.ParseReader(bytes.NewReader([]byte(cyacdContent)))
	if err != nil {
		log.Fatalf("Failed to parse firmware: %v", err)
	}

	fmt.Printf("Firmware loaded:\n")
	fmt.Printf("  Silicon ID:    0x%08X\n", fw.SiliconID)
	fmt.Printf("  Silicon Rev:   0x%02X\n", fw.SiliconRev)
	fmt.Printf("  Checksum Type: 0x%02X\n", fw.ChecksumType)
	fmt.Printf("  Total Rows:    %d\n", len(fw.Rows))
	fmt.Println()

	// Create programmer with progress tracking
	prog := bootloader.New(device,
		bootloader.WithProgressCallback(func(p bootloader.Progress) {
			fmt.Printf("[%s] %.1f%% - Row %d/%d\n",
				p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
		}),
	)

	// Program the device
	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F} // Example bootloader key

	fmt.Println("Starting programming...")
	fmt.Println()

	// Note: With a mock device, this will fail when trying to communicate
	// with the actual bootloader. This example demonstrates the API usage.
	err = prog.Program(context.Background(), fw, key)
	if err != nil {
		fmt.Printf("Programming failed (expected with mock device): %v\n", err)
		fmt.Println()
		fmt.Println("In a real application with actual hardware, this would program the device.")
		return
	}

	fmt.Println()
	fmt.Println("Programming successful!")
}
