package bootloader

import "time"

// Phase represents the current operation phase during firmware programming.
// Use the exported Phase constants for type-safe comparisons.
type Phase string

// Phase constants for firmware programming operations.
// Use these constants when checking Progress.Phase in your callbacks.
//
// Example:
//
//	callback := func(p bootloader.Progress) {
//	    switch p.Phase {
//	    case bootloader.PhaseEntering:
//	        fmt.Println("Entering bootloader...")
//	    case bootloader.PhaseProgramming:
//	        fmt.Printf("Programming: %.1f%%\n", p.Percentage)
//	    case bootloader.PhaseComplete:
//	        fmt.Println("Done!")
//	    }
//	}
const (
	// PhaseEntering indicates the bootloader is being entered
	PhaseEntering Phase = "entering"

	// PhaseProgramming indicates flash rows are being programmed
	PhaseProgramming Phase = "programming"

	// PhaseVerifying indicates firmware is being verified
	PhaseVerifying Phase = "verifying"

	// PhaseExiting indicates the bootloader is being exited
	PhaseExiting Phase = "exiting"

	// PhaseComplete indicates the operation completed successfully
	PhaseComplete Phase = "complete"
)

// Progress contains information about the programming progress.
// Passed to ProgressCallback during programming operations.
type Progress struct {
	// Phase describes the current operation phase.
	// Compare against exported Phase constants (PhaseEntering, PhaseProgramming, etc.)
	Phase Phase

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
