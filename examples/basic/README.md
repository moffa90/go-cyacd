# Basic Example

A minimal example demonstrating the core go-cyacd API for programming Cypress/Infineon PSoC microcontrollers.

## What This Example Shows

- Basic firmware parsing
- Creating a programmer instance
- Simple progress tracking
- Programming flow with error handling
- Mock device usage (for demonstration)

## Running the Example

```bash
cd examples/basic
go run main.go
```

## Code Overview

```go
// 1. Parse firmware
fw, err := cyacd.ParseReader(reader)

// 2. Create programmer with progress callback
prog := bootloader.New(device,
    bootloader.WithProgressCallback(func(p bootloader.Progress) {
        fmt.Printf("[%s] %.1f%% - Row %d/%d\n",
            p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
    }),
)

// 3. Program the device
key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}
err = prog.Program(context.Background(), fw, key)
```

## Important Notes

### Mock Device

This example uses a mock device for demonstration. The mock device doesn't actually communicate with hardware, so the programming will fail when it tries to read responses.

In a real application, you would replace the mock device with your actual hardware implementation:

```go
// Example: USB HID device
device, err := usb.OpenDevice("ABC12345")
if err != nil {
    log.Fatal(err)
}
defer device.Close()
```

### Firmware File

The example creates a minimal valid `.cyacd` firmware inline:

```
1E9602AA0000              # Header
000000040001020304F2      # Row 0: ArrayID=0, RowNum=0, Size=4, Data=[1,2,3,4]
```

In a real application, you would load from a file:

```go
fw, err := cyacd.Parse("firmware.cyacd")
```

## Next Steps

- See [with_progress](../with_progress/) for advanced progress tracking
- See [advanced](../advanced/) for production-ready configuration
- See [mock_device](../mock_device/) for a realistic mock device implementation
- Read the [main README](../../README.md) for complete documentation

## Hardware Implementation

To use this with real hardware, implement the `io.ReadWriter` interface:

```go
type YourDevice struct {
    // Your hardware-specific fields
}

func (d *YourDevice) Read(p []byte) (int, error) {
    // Implement reading from your hardware
    // (USB, UART, SPI, I2C, etc.)
}

func (d *YourDevice) Write(p []byte) (int, error) {
    // Implement writing to your hardware
}
```

See the [Protocol Documentation](../../docs/PROTOCOL.md) for details on the communication protocol.
