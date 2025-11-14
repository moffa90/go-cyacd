package bootloader

import "time"

// Progress contains information about the programming progress.
// Passed to ProgressCallback during programming operations.
type Progress struct {
	// Phase describes the current operation phase:
	//   "entering"    - Entering bootloader mode
	//   "programming" - Programming flash rows
	//   "verifying"   - Verifying firmware
	//   "exiting"     - Exiting bootloader
	//   "complete"    - Operation completed successfully
	Phase string

	// CurrentRow is the current row being programmed (0-based)
	CurrentRow int

	// TotalRows is the total number of rows to program
	TotalRows int

	// Percentage is the completion percentage (0.0 to 100.0)
	Percentage float64

	// BytesWritten is the total number of bytes written so far
	BytesWritten int

	// ElapsedTime is the time elapsed since programming started
	ElapsedTime time.Duration
}

// ProgressCallback is called periodically during programming to report progress.
// Implementations should return quickly to avoid blocking the programming operation.
//
// Example:
//
//	prog := bootloader.New(device,
//	    bootloader.WithProgressCallback(func(p bootloader.Progress) {
//	        fmt.Printf("[%s] %.1f%% - Row %d/%d\n",
//	            p.Phase, p.Percentage, p.CurrentRow, p.TotalRows)
//	    }),
//	)
type ProgressCallback func(Progress)

// Logger is an optional logging interface that can be provided to the programmer.
// This allows integration with any logging framework.
//
// Example with standard log package:
//
//	type StdLogger struct{}
//	func (l *StdLogger) Debug(msg string, kv ...interface{}) { log.Println(msg, kv) }
//	func (l *StdLogger) Info(msg string, kv ...interface{})  { log.Println(msg, kv) }
//	func (l *StdLogger) Error(msg string, kv ...interface{}) { log.Println(msg, kv) }
//
//	prog := bootloader.New(device, bootloader.WithLogger(&StdLogger{}))
type Logger interface {
	// Debug logs a debug message with optional key-value pairs
	Debug(msg string, keysAndValues ...interface{})

	// Info logs an info message with optional key-value pairs
	Info(msg string, keysAndValues ...interface{})

	// Error logs an error message with optional key-value pairs
	Error(msg string, keysAndValues ...interface{})
}
