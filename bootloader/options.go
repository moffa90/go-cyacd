package bootloader

import "time"

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
	// Default is 57 bytes (fits in 64-byte USB packets)
	ChunkSize int

	// Retries is the number of retry attempts for failed commands
	Retries int

	// VerifyAfterProgram enables row verification after each program operation
	VerifyAfterProgram bool
}

// defaultConfig returns the default configuration.
func defaultConfig() Config {
	return Config{
		ReadTimeout:        5 * time.Second,
		WriteTimeout:       5 * time.Second,
		ChunkSize:          57, // 64-byte packet - 7 byte overhead
		Retries:            3,
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
// Default is 57 bytes (to fit in 64-byte USB packets).
//
// Example:
//
//	prog := bootloader.New(device, bootloader.WithChunkSize(64))
func WithChunkSize(size int) Option {
	return func(c *Config) {
		if size > 0 && size <= 256 {
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
