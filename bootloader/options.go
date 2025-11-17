package bootloader

import "time"

// Default configuration values.
const (
	// DefaultReadTimeout is the default timeout for read operations
	DefaultReadTimeout = 5 * time.Second

	// DefaultWriteTimeout is the default timeout for write operations
	DefaultWriteTimeout = 5 * time.Second

	// DefaultChunkSize is the default maximum data size per Send Data command
	// Set to 64 bytes to fit within typical USB HID report sizes
	DefaultChunkSize = 64

	// DefaultRetries is the default number of retry attempts for failed commands
	DefaultRetries = 3

	// MaxChunkSize is the maximum allowed chunk size per packet
	MaxChunkSize = 256
)

// Config holds the programmer configuration.
type Config struct {
	// ProgressCallback is called during programming to report progress (optional)
	ProgressCallback ProgressCallback

	// Logger is used for logging operations (optional)
	Logger Logger

	// ReadTimeout is the timeout for read operations
	ReadTimeout time.Duration

	// WriteTimeout is the timeout for write operations
	WriteTimeout time.Duration

	// ChunkSize is the maximum data size per Send Data command
	// Default is 64 bytes
	ChunkSize int

	// Retries is the number of retry attempts for failed commands
	Retries int

	// VerifyAfterProgram enables row verification after each program operation
	VerifyAfterProgram bool

	// CommandDelay is the delay between consecutive commands
	// USB/HID typically use 1ms, Serial typically uses 25ms
	// Default is 0 (no delay)
	CommandDelay time.Duration

	// LenientVerifyRow allows accepting 0-byte or 1-byte VerifyRow responses
	// Default is false (strict: require exactly 1 byte per Infineon spec)
	// Enable this for legacy or non-standard bootloader firmware that returns 0 bytes
	LenientVerifyRow bool
}

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	return Config{
		ReadTimeout:        DefaultReadTimeout,
		WriteTimeout:       DefaultWriteTimeout,
		ChunkSize:          DefaultChunkSize,
		Retries:            DefaultRetries,
		VerifyAfterProgram: true,
	}
}

// Option is a functional option for configuring the Programmer.
type Option func(*Config)

// WithProgressCallback sets a callback function to track programming progress.
//
// Example:
//
//	prog := bootloader.New(device,
//	    bootloader.WithProgressCallback(func(p bootloader.Progress) {
//	        fmt.Printf("%.1f%% complete\n", p.Percentage)
//	    }),
//	)
func WithProgressCallback(callback ProgressCallback) Option {
	return func(c *Config) {
		c.ProgressCallback = callback
	}
}

// WithLogger sets a logger for the programmer operations.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithLogger(myLogger))
func WithLogger(logger Logger) Option {
	return func(c *Config) {
		c.Logger = logger
	}
}

// WithTimeout sets both read and write timeouts.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithTimeout(10*time.Second))
func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
		c.WriteTimeout = timeout
	}
}

// WithReadTimeout sets the read timeout.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithReadTimeout(5*time.Second))
func WithReadTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.ReadTimeout = timeout
	}
}

// WithWriteTimeout sets the write timeout.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithWriteTimeout(5*time.Second))
func WithWriteTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.WriteTimeout = timeout
	}
}

// WithChunkSize sets the maximum data size per Send Data command.
// Default is DefaultChunkSize (64 bytes).
// Maximum allowed is MaxChunkSize (256 bytes).
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithChunkSize(128))
func WithChunkSize(size int) Option {
	return func(c *Config) {
		if size > 0 && size <= MaxChunkSize {
			c.ChunkSize = size
		}
	}
}

// WithRetries sets the number of retry attempts for failed commands.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithRetries(5))
func WithRetries(retries int) Option {
	return func(c *Config) {
		if retries >= 0 {
			c.Retries = retries
		}
	}
}

// WithVerifyAfterProgram enables or disables row verification after programming.
// Default is true.
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithVerifyAfterProgram(false))
func WithVerifyAfterProgram(verify bool) Option {
	return func(c *Config) {
		c.VerifyAfterProgram = verify
	}
}

// WithCommandDelay sets the delay between consecutive commands.
// This is useful for slower transports like Serial which may need 25ms delays,
// while USB/HID typically work fine with 1ms or no delay.
//
// Example:
//
//	// For Serial devices
//	prog := bootloader.New(device, bootloader.WithCommandDelay(25*time.Millisecond))
//
//	// For USB/HID devices
//	prog := bootloader.New(device, bootloader.WithCommandDelay(1*time.Millisecond))
func WithCommandDelay(delay time.Duration) Option {
	return func(c *Config) {
		if delay >= 0 {
			c.CommandDelay = delay
		}
	}
}

// WithLenientVerifyRow enables lenient validation for VerifyRow command responses.
// When enabled, accepts both 0-byte (returns 0x00) and 1-byte (returns checksum) responses.
// Default is false (strict mode: require exactly 1 byte per Infineon AN60317 specification).
//
// Use this option for legacy or non-standard bootloader firmware that returns 0-byte
// responses for command 0x3A (Get Row Checksum) instead of the required 1-byte checksum.
//
// Example:
//
//	// For devices with non-standard firmware
//	prog := bootloader.New(device, bootloader.WithLenientVerifyRow())
func WithLenientVerifyRow() Option {
	return func(c *Config) {
		c.LenientVerifyRow = true
	}
}
