# Advanced Example

A production-ready example demonstrating comprehensive configuration, error handling, logging, and command-line interface.

## What This Example Shows

- Command-line argument parsing
- Configuration management
- Custom logging implementation
- Comprehensive error handling with actionable messages
- All programmer options (timeouts, retries, chunk size, etc.)
- Context timeout handling
- Statistics and performance metrics

## Running the Example

```bash
cd examples/advanced
go run main.go -firmware <path/to/firmware.cyacd> [options]
```

## Command-Line Options

```
  -firmware string
        Path to .cyacd firmware file (required)
  -key string
        Bootloader key (12 hex characters) (default "0A1B2C3D4E5F")
  -timeout duration
        Programming timeout (default 5m0s)
  -chunk-size int
        Data chunk size in bytes (default 57)
  -retries int
        Number of retries on failure (default 3)
  -no-verify
        Skip verification after programming
  -debug
        Enable debug logging
```

## Usage Examples

### Basic Usage

```bash
go run main.go -firmware firmware.cyacd
```

### With Custom Key

```bash
go run main.go -firmware firmware.cyacd -key AABBCCDDEEFF
```

### With Debug Logging

```bash
go run main.go -firmware firmware.cyacd -debug
```

### Skip Verification (Not Recommended)

```bash
go run main.go -firmware firmware.cyacd -no-verify
```

### Adjust Timeouts and Retries

```bash
go run main.go -firmware firmware.cyacd -timeout 10m -retries 5
```

### Small Chunk Size for Unreliable Connections

```bash
go run main.go -firmware firmware.cyacd -chunk-size 32
```

## Features

### Custom Logger

Implements the `bootloader.Logger` interface:

```go
type CustomLogger struct {
    logger *log.Logger
    debug  bool
}

func (l *CustomLogger) Debug(msg string, kv ...interface{}) {
    if l.debug {
        l.logger.Printf("[DEBUG] %s %v", msg, kv)
    }
}

func (l *CustomLogger) Info(msg string, kv ...interface{}) {
    l.logger.Printf("[INFO] %s %v", msg, kv)
}

func (l *CustomLogger) Error(msg string, kv ...interface{}) {
    l.logger.Printf("[ERROR] %s %v", msg, kv)
}
```

### Comprehensive Error Handling

The example provides detailed, actionable error messages:

#### Device Mismatch
```
Error: Wrong device connected
  Expected Silicon ID: 0x1E9602AA
  Actual Silicon ID:   0x12345678

Please verify you have the correct device and firmware file.
```

#### Row Out of Range
```
Error: Firmware incompatible with device
  Row 256 (array 0) is outside valid range 0-255

Please use a firmware file compatible with your device.
```

#### Checksum Mismatch
```
Error: Data corruption detected
  Row 42 failed checksum verification
  Expected: 0xAB, Got: 0xCD

This could indicate communication errors. Try:
  - Reducing chunk size with -chunk-size flag
  - Checking cable connections
  - Verifying power supply stability
```

#### Verification Failure
```
Error: Application verification failed
  Reason: Application checksum is invalid

The firmware was written but failed final verification.
Try programming again or use -no-verify flag (not recommended).
```

### Configuration Struct

```go
type Config struct {
    FirmwarePath string
    KeyHex       string
    Timeout      time.Duration
    ChunkSize    int
    Retries      int
    NoVerify     bool
    Debug        bool
}
```

### All Programmer Options

```go
prog := bootloader.New(device,
    bootloader.WithLogger(logger),
    bootloader.WithProgressCallback(progressFunc),
    bootloader.WithTimeout(cfg.Timeout),
    bootloader.WithChunkSize(cfg.ChunkSize),
    bootloader.WithRetries(cfg.Retries),
    bootloader.WithVerifyAfterProgram(!cfg.NoVerify),
)
```

### Statistics

After successful programming, displays:

```
✅ Programming completed successfully in 15.234s

Statistics:
  Rows programmed:  256
  Bytes written:    16384
  Average speed:    16.8 rows/sec
  Throughput:       1075.2 bytes/sec

Device is ready to use.
```

## Production Deployment

This example is suitable for production use. Consider:

### Environment Variables

Add support for environment variables:

```go
key := os.Getenv("BOOTLOADER_KEY")
if key == "" {
    key = cfg.KeyHex
}
```

### Configuration Files

Load settings from config file:

```go
// Load from config.json
cfg := loadConfig("config.json")
```

### Structured Logging

Replace simple logger with structured logging (e.g., zerolog, zap):

```go
import "github.com/rs/zerolog/log"

func (l *CustomLogger) Info(msg string, kv ...interface{}) {
    event := log.Info()
    for i := 0; i < len(kv); i += 2 {
        event = event.Interface(kv[i].(string), kv[i+1])
    }
    event.Msg(msg)
}
```

### Metrics

Add Prometheus metrics:

```go
var (
    programmingDuration = prometheus.NewHistogram(...)
    programmingErrors   = prometheus.NewCounter(...)
)
```

### Signal Handling

Handle SIGINT/SIGTERM gracefully:

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

go func() {
    <-sigCh
    cancel() // Cancel context
}()
```

## Troubleshooting

### Timeouts

If programming times out:
- Increase `-timeout` value
- Check device connection
- Verify device is responding

### Checksum Errors

If checksums fail frequently:
- Reduce `-chunk-size` (try 32 or 16)
- Check cable quality
- Ensure stable power supply
- Increase `-retries`

### Verification Failures

If verification fails:
- Try programming again
- Verify firmware file integrity
- Check device flash isn't corrupted
- As last resort: `-no-verify` (not recommended)

## Next Steps

- See [basic](../basic/) for minimal example
- See [with_progress](../with_progress/) for visual progress tracking
- See [mock_device](../mock_device/) for device simulation
- Read [PROTOCOL.md](../../docs/PROTOCOL.md) for protocol details
- Read [CONTRIBUTING.md](../../CONTRIBUTING.md) for development guidelines
