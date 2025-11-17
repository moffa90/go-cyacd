package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/moffa90/go-cyacd/bootloader"
	"github.com/moffa90/go-cyacd/cyacd"
)

// CustomLogger implements bootloader.Logger interface
type CustomLogger struct {
	logger *log.Logger
	debug  bool
}

func NewCustomLogger(debug bool) *CustomLogger {
	return &CustomLogger{
		logger: log.New(os.Stdout, "", log.LstdFlags),
		debug:  debug,
	}
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

// MockDevice simulates a bootloader device
type MockDevice struct {
	failureRate float64
	callCount   int
}

func (d *MockDevice) Read(p []byte) (int, error) {
	time.Sleep(5 * time.Millisecond)
	d.callCount++
	// Simulate occasional read failures
	if d.failureRate > 0 && d.callCount%10 == 0 {
		return 0, fmt.Errorf("simulated read error")
	}
	return len(p), nil
}

func (d *MockDevice) Write(p []byte) (int, error) {
	time.Sleep(5 * time.Millisecond)
	d.callCount++
	return len(p), nil
}

// Config holds application configuration
type Config struct {
	FirmwarePath string
	KeyHex       string
	Timeout      time.Duration
	ChunkSize    int
	Retries      int
	NoVerify     bool
	Debug        bool
}

func parseFlags() *Config {
	cfg := &Config{}

	flag.StringVar(&cfg.FirmwarePath, "firmware", "", "Path to .cyacd firmware file (required)")
	flag.StringVar(&cfg.KeyHex, "key", "0A1B2C3D4E5F", "Bootloader key (12 hex characters)")
	flag.DurationVar(&cfg.Timeout, "timeout", 5*time.Minute, "Programming timeout")
	flag.IntVar(&cfg.ChunkSize, "chunk-size", 57, "Data chunk size in bytes")
	flag.IntVar(&cfg.Retries, "retries", 3, "Number of retries on failure")
	flag.BoolVar(&cfg.NoVerify, "no-verify", false, "Skip verification after programming")
	flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")

	flag.Parse()

	if cfg.FirmwarePath == "" {
		fmt.Println("Error: -firmware flag is required")
		flag.Usage()
		os.Exit(1)
	}

	return cfg
}

func parseKey(keyHex string) ([]byte, error) {
	if len(keyHex) != 12 {
		return nil, fmt.Errorf("key must be exactly 12 hex characters, got %d", len(keyHex))
	}

	key := make([]byte, 6)
	for i := 0; i < 6; i++ {
		_, err := fmt.Sscanf(keyHex[i*2:i*2+2], "%02x", &key[i])
		if err != nil {
			return nil, fmt.Errorf("invalid hex in key: %w", err)
		}
	}

	return key, nil
}

func main() {
	fmt.Println("=== Cypress Bootloader - Advanced Example ===\n")

	// Parse configuration
	cfg := parseFlags()

	// Parse bootloader key
	key, err := parseKey(cfg.KeyHex)
	if err != nil {
		log.Fatalf("Invalid key: %v", err)
	}

	// Load firmware
	fmt.Printf("Loading firmware: %s\n", cfg.FirmwarePath)
	fw, err := cyacd.Parse(cfg.FirmwarePath)
	if err != nil {
		log.Fatalf("Failed to parse firmware: %v", err)
	}

	fmt.Printf("Firmware details:\n")
	fmt.Printf("  Silicon ID:     0x%08X\n", fw.SiliconID)
	fmt.Printf("  Silicon Rev:    0x%02X\n", fw.SiliconRev)
	fmt.Printf("  Checksum Type:  0x%02X\n", fw.ChecksumType)
	fmt.Printf("  Total Rows:     %d\n", len(fw.Rows))
	fmt.Println()

	// Create device (replace with your actual hardware)
	device := &MockDevice{failureRate: 0.0}

	// Setup custom logger
	logger := NewCustomLogger(cfg.Debug)

	// Create programmer with all options
	prog := bootloader.New(device,
		bootloader.WithLogger(logger),
		bootloader.WithProgressCallback(func(p bootloader.Progress) {
			switch p.Phase {
			case bootloader.PhaseEntering:
				fmt.Printf("⏳ Entering bootloader mode...\n")
			case bootloader.PhaseProgramming:
				fmt.Printf("📝 Programming: %.1f%% - Row %d/%d\n",
					p.Percentage, p.CurrentRow, p.TotalRows)
			case bootloader.PhaseVerifying:
				fmt.Printf("✓ Verifying application...\n")
			case bootloader.PhaseExiting:
				fmt.Printf("🚀 Exiting bootloader...\n")
			case bootloader.PhaseComplete:
				fmt.Printf("✅ Programming complete!\n")
			}
		}),
		bootloader.WithTimeout(cfg.Timeout),
		bootloader.WithChunkSize(cfg.ChunkSize),
		bootloader.WithRetries(cfg.Retries),
		bootloader.WithVerifyAfterProgram(!cfg.NoVerify),
	)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	// Program the device
	fmt.Println("Starting programming sequence...")
	fmt.Printf("Configuration:\n")
	fmt.Printf("  Timeout:       %s\n", cfg.Timeout)
	fmt.Printf("  Chunk Size:    %d bytes\n", cfg.ChunkSize)
	fmt.Printf("  Retries:       %d\n", cfg.Retries)
	fmt.Printf("  Verification:  %t\n", !cfg.NoVerify)
	fmt.Println()

	startTime := time.Now()
	err = prog.Program(ctx, fw, key)
	duration := time.Since(startTime)

	fmt.Println()

	// Handle errors with detailed messages
	if err != nil {
		fmt.Printf("❌ Programming failed after %s\n\n", duration.Round(time.Millisecond))

		switch e := err.(type) {
		case *bootloader.DeviceMismatchError:
			fmt.Printf("Error: Wrong device connected\n")
			fmt.Printf("  Expected Silicon ID: 0x%08X\n", e.Expected)
			fmt.Printf("  Actual Silicon ID:   0x%08X\n", e.Actual)
			fmt.Println("\nPlease verify you have the correct device and firmware file.")

		case *bootloader.RowOutOfRangeError:
			fmt.Printf("Error: Firmware incompatible with device\n")
			fmt.Printf("  Row %d (array %d) is outside valid range %d-%d\n",
				e.RowNum, e.ArrayID, e.MinRow, e.MaxRow)
			fmt.Println("\nPlease use a firmware file compatible with your device.")

		case *bootloader.ChecksumMismatchError:
			fmt.Printf("Error: Data corruption detected\n")
			fmt.Printf("  Row %d failed checksum verification\n", e.RowNum)
			fmt.Printf("  Expected: 0x%02X, Got: 0x%02X\n", e.Expected, e.Actual)
			fmt.Println("\nThis could indicate communication errors. Try:")
			fmt.Println("  - Reducing chunk size with -chunk-size flag")
			fmt.Println("  - Checking cable connections")
			fmt.Println("  - Verifying power supply stability")

		case *bootloader.VerificationError:
			fmt.Printf("Error: Application verification failed\n")
			fmt.Printf("  Reason: %s\n", e.Reason)
			fmt.Println("\nThe firmware was written but failed final verification.")
			fmt.Println("Try programming again or use -no-verify flag (not recommended).")

		default:
			fmt.Printf("Error: %v\n", err)
		}

		os.Exit(1)
	}

	// Success!
	fmt.Printf("✅ Programming completed successfully in %s\n\n", duration.Round(time.Millisecond))

	// Calculate statistics
	totalBytes := 0
	for _, row := range fw.Rows {
		totalBytes += len(row.Data)
	}

	avgSpeed := float64(len(fw.Rows)) / duration.Seconds()
	throughput := float64(totalBytes) / duration.Seconds()

	fmt.Println("Statistics:")
	fmt.Printf("  Rows programmed:  %d\n", len(fw.Rows))
	fmt.Printf("  Bytes written:    %d\n", totalBytes)
	fmt.Printf("  Average speed:    %.1f rows/sec\n", avgSpeed)
	fmt.Printf("  Throughput:       %.1f bytes/sec\n", throughput)
	fmt.Println("\nDevice is ready to use.")
}
