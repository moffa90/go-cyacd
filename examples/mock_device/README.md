# Mock Device Example

A comprehensive example demonstrating how to implement a realistic mock Cypress bootloader device for testing without hardware.

## What This Example Shows

- Complete mock device implementation
- Proper protocol response generation
- Simulated flash memory
- Command parsing and handling
- Frame checksum calculation
- Device state management
- Realistic timing simulation

## Running the Example

```bash
cd examples/mock_device
go run main.go
```

This will program a test firmware using the mock device and display detailed logging of all communications.

## Mock Device Features

The `RealisticMockDevice` simulates a real Cypress PSoC bootloader:

### Simulated Hardware

- **Silicon ID**: 0x1E9602AA (PSoC 5LP)
- **Flash Range**: 0x0000 - 0x01FF (512 rows)
- **Flash Memory**: In-memory map storing programmed data
- **Bootloader State**: Tracks whether in bootloader mode
- **Latency**: Simulates device response time (10ms)

### Supported Commands

The mock device handles all essential bootloader commands:

- ✅ **Enter Bootloader** (0x38) - Enter bootloader mode
- ✅ **Get Flash Size** (0x32) - Return valid flash range
- ✅ **Program Row** (0x39) - Store data in simulated flash
- ✅ **Verify Row** (0x3A) - Calculate and return row checksum
- ✅ **Verify Checksum** (0x31) - Validate application
- ✅ **Exit Bootloader** (0x3B) - Exit bootloader mode
- ✅ **Send Data** (0x37) - Buffer data chunks

### Protocol Compliance

- Proper frame structure (SOP, Status, Length, Data, Checksum, EOP)
- Correct checksum calculation
- Little-endian byte ordering
- Status code responses
- State validation

## Code Structure

### Mock Device Implementation

```go
type RealisticMockDevice struct {
    siliconID    uint32
    siliconRev   byte
    flashStart   uint16
    flashEnd     uint16
    flash        map[uint16][]byte  // Simulated flash
    inBootloader bool
    latency      time.Duration
    responseQueue []byte
}
```

### Read/Write Methods

The device implements `io.ReadWriter`:

```go
func (d *RealisticMockDevice) Read(p []byte) (int, error) {
    // Simulate device response time
    time.Sleep(d.latency)

    // Return exactly one complete frame
    // Parse frame length, copy to buffer, remove from queue
}

func (d *RealisticMockDevice) Write(p []byte) (int, error) {
    // Parse command
    // Generate appropriate response
    // Queue response for Read()
}
```

### Command Handlers

Each command has a dedicated handler:

```go
func (d *RealisticMockDevice) handleEnterBootloader() []byte
func (d *RealisticMockDevice) handleGetFlashSize(frame []byte) []byte
func (d *RealisticMockDevice) handleProgramRow(frame []byte) []byte
func (d *RealisticMockDevice) handleVerifyRow(frame []byte) []byte
func (d *RealisticMockDevice) handleVerifyChecksum() []byte
func (d *RealisticMockDevice) handleExitBootloader() []byte
func (d *RealisticMockDevice) handleSendData(frame []byte) []byte
```

### Response Generation

```go
func buildResponseFrame(statusCode byte, data []byte) []byte {
    // Build frame: SOP + Status + Len + Data + Checksum + EOP
    // Calculate checksum over SOP through Data
    // Return complete frame
}
```

## Detailed Logging

The example includes extensive logging:

```
[DEVICE] Entered bootloader mode
[PROGRAMMER] entering: 0.0% (Row 0/2)
[DEVICE] Flash size: 0x0000 - 0x01FF
[DEVICE] Programmed row 0 (array 0): 8 bytes
[PROGRAMMER] programming: 50.0% (Row 1/2)
[DEVICE] Verified row 0: checksum=0xC4
[DEVICE] Programmed row 1 (array 0): 8 bytes
[DEVICE] Verified row 1: checksum=0x54
[PROGRAMMER] verifying: 100.0% (Row 2/2)
[DEVICE] Application checksum valid
[PROGRAMMER] exiting: 100.0% (Row 2/2)
[DEVICE] Exited bootloader mode
[PROGRAMMER] complete: 100.0% (Row 2/2)

✅ Programming completed successfully!

Mock Device State:
  In Bootloader: false
  Flash Rows:    2
    Row 0x0000:  8 bytes: 01 02 03 04 05 06 07 08
    Row 0x0001:  8 bytes: 11 12 13 14 15 16 17 18
```

## Use Cases

### 1. Development Without Hardware

Test your bootloader integration before hardware is available:

```go
device := NewRealisticMockDevice()
prog := bootloader.New(device)
err := prog.Program(ctx, fw, key)
```

### 2. Automated Testing

Use in unit tests and CI/CD:

```go
func TestProgramming(t *testing.T) {
    device := NewRealisticMockDevice()
    prog := bootloader.New(device)

    err := prog.Program(context.Background(), firmware, key)
    if err != nil {
        t.Fatalf("Programming failed: %v", err)
    }

    // Verify flash contents
    if len(device.flash) != len(firmware.Rows) {
        t.Errorf("Expected %d rows, got %d", len(firmware.Rows), len(device.flash))
    }
}
```

### 3. Error Simulation

Simulate error conditions:

```go
// Simulate device with smaller flash
device := NewRealisticMockDevice()
device.flashEnd = 0x00FF  // Only 256 rows

// Try to program 512-row firmware
// Should fail with RowOutOfRangeError
```

### 4. Performance Testing

Measure programming performance:

```go
device := NewRealisticMockDevice()
device.latency = 1 * time.Millisecond  // Fast device

start := time.Now()
prog.Program(ctx, largeFirmware, key)
duration := time.Since(start)

fmt.Printf("Programmed %d rows in %s\n", len(firmware.Rows), duration)
```

## Customization

### Adjust Device Parameters

```go
device := NewRealisticMockDevice()
device.siliconID = 0x04A2B456      // Different chip
device.flashEnd = 0x03FF            // Larger flash (1024 rows)
device.latency = 50 * time.Millisecond  // Slower device
```

### Add Error Simulation

```go
type UnreliableMockDevice struct {
    *RealisticMockDevice
    errorRate float64
}

func (d *UnreliableMockDevice) Write(p []byte) (int, error) {
    if rand.Float64() < d.errorRate {
        return 0, fmt.Errorf("simulated communication error")
    }
    return d.RealisticMockDevice.Write(p)
}
```

### Track Statistics

```go
type StatsMockDevice struct {
    *RealisticMockDevice
    commandCount map[byte]int
}

func (d *StatsMockDevice) Write(p []byte) (int, error) {
    cmd := p[1]
    d.commandCount[cmd]++
    return d.RealisticMockDevice.Write(p)
}
```

## Limitations

While realistic, this mock device has some limitations:

1. **No CRC-16 Support**: Only implements basic sum checksums
2. **Simplified Verification**: Always returns valid checksums
3. **No Multi-App Support**: Single application only
4. **No Erase Simulation**: Erase command not fully implemented
5. **No Timing Constraints**: Doesn't enforce real timing requirements

For production use, you should test with actual hardware to catch timing-sensitive issues and hardware-specific behaviors.

## Next Steps

- Use this mock device in your tests
- Extend it with error simulation for robust testing
- See [basic](../basic/) for simple usage example
- See [advanced](../advanced/) for production configuration
- Read [PROTOCOL.md](../../docs/PROTOCOL.md) for protocol specification
- Read [CONTRIBUTING.md](../../CONTRIBUTING.md) to contribute improvements
