// Package bootloader provides a high-level API for programming Cypress/Infineon microcontrollers.
//
// # Overview
//
// This package orchestrates the complete firmware programming sequence:
//   - Entering bootloader mode with a security key
//   - Validating device compatibility
//   - Programming flash rows with automatic chunking
//   - Verifying programmed data
//   - Exiting bootloader mode
//
// # Basic Usage
//
// The simplest way to program a device:
//
//	// User provides hardware communication (io.ReadWriter)
//	device := myusb.OpenDevice("ABC12345")
//
//	// Parse firmware file
//	fw, err := cyacd.Parse("firmware.cyacd")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create programmer
//	prog := bootloader.New(device)
//
//	// Program with 6-byte key
//	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}
//	err = prog.Program(context.Background(), fw, key)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Progress Tracking
//
// Track programming progress with a callback:
//
//	prog := bootloader.New(device,
//	    bootloader.WithProgressCallback(func(p bootloader.Progress) {
//	        fmt.Printf("[%s] %.1f%% - Row %d/%d\n",
//	            p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
//	    }),
//	)
//
// # Configuration Options
//
// Customize behavior with functional options:
//
//	prog := bootloader.New(device,
//	    bootloader.WithProgressCallback(progressFunc),
//	    bootloader.WithLogger(myLogger),
//	    bootloader.WithTimeout(30*time.Second),
//	    bootloader.WithChunkSize(64),
//	    bootloader.WithRetries(5),
//	    bootloader.WithVerifyAfterProgram(true),
//	)
//
// # Logging
//
// Integrate with any logging framework:
//
//	type MyLogger struct {
//	    logger *log.Logger
//	}
//
//	func (l *MyLogger) Debug(msg string, kv ...interface{}) {
//	    l.logger.Println("DEBUG:", msg, kv)
//	}
//
//	func (l *MyLogger) Info(msg string, kv ...interface{}) {
//	    l.logger.Println("INFO:", msg, kv)
//	}
//
//	func (l *MyLogger) Error(msg string, kv ...interface{}) {
//	    l.logger.Println("ERROR:", msg, kv)
//	}
//
//	prog := bootloader.New(device, bootloader.WithLogger(&MyLogger{...}))
//
// # Context Support
//
// All operations support context for cancellation and timeouts:
//
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
//	defer cancel()
//
//	err := prog.Program(ctx, fw, key)
//
// # Error Handling
//
// The package provides structured error types:
//   - DeviceMismatchError: Silicon ID doesn't match firmware
//   - RowOutOfRangeError: Row number exceeds flash size
//   - ChecksumMismatchError: Row verification failed
//   - VerificationError: Application checksum failed
//   - protocol.ProtocolError: Bootloader returned an error status
//
// # Hardware Independence
//
// This package does NOT implement hardware communication.
// Users must provide an io.ReadWriter implementation for their specific hardware:
//
//	type MyUSBDevice struct {
//	    // ... your USB implementation
//	}
//
//	func (d *MyUSBDevice) Read(p []byte) (int, error) {
//	    // ... your Read implementation
//	}
//
//	func (d *MyUSBDevice) Write(p []byte) (int, error) {
//	    // ... your Write implementation
//	}
//
// This design allows the library to work with any communication method:
// USB, HID, UART, SPI, I2C, network, or even mock devices for testing.
package bootloader
