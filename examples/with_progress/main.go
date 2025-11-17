package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/moffa90/go-cyacd/bootloader"
	"github.com/moffa90/go-cyacd/cyacd"
)

// MockDevice simulates a bootloader device for demonstration
type MockDevice struct{}

func (d *MockDevice) Read(p []byte) (int, error) {
	time.Sleep(10 * time.Millisecond) // Simulate device latency
	return len(p), nil
}

func (d *MockDevice) Write(p []byte) (int, error) {
	time.Sleep(5 * time.Millisecond) // Simulate device latency
	return len(p), nil
}

// ProgressBar renders a visual progress bar
type ProgressBar struct {
	width int
}

func NewProgressBar(width int) *ProgressBar {
	return &ProgressBar{width: width}
}

func (pb *ProgressBar) Render(percentage float64) string {
	filled := int(float64(pb.width) * percentage / 100.0)
	if filled > pb.width {
		filled = pb.width
	}

	bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
	return fmt.Sprintf("[%s] %.1f%%", bar, percentage)
}

func main() {
	fmt.Println("=== Cypress Bootloader - Progress Tracking Example ===")

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: with_progress <firmware.cyacd>")
		fmt.Println("\nThis example demonstrates advanced progress tracking during programming.")
		fmt.Println("It shows real-time updates with:")
		fmt.Println("  - Visual progress bar")
		fmt.Println("  - Current phase (entering, programming, verifying, etc.)")
		fmt.Println("  - Row count and percentage")
		fmt.Println("  - Bytes written")
		fmt.Println("  - Elapsed time")
		fmt.Println("  - Estimated time remaining")
		os.Exit(1)
	}

	// Parse firmware file
	fmt.Printf("Loading firmware: %s\n", os.Args[1])
	fw, err := cyacd.Parse(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to parse firmware: %v", err)
	}

	fmt.Printf("Firmware loaded: SiliconID=0x%08X, Rows=%d\n\n", fw.SiliconID, len(fw.Rows))

	// Create mock device (replace with your actual hardware)
	device := &MockDevice{}

	// Setup progress tracking
	progressBar := NewProgressBar(40)
	startTime := time.Now()
	var lastPhase string

	progressCallback := func(p bootloader.Progress) {
		// Clear line and move cursor to beginning
		fmt.Print("\r\033[K")

		// Print phase header if changed
		if p.Phase != lastPhase {
			if lastPhase != "" {
				fmt.Println() // New line after previous phase
			}
			fmt.Printf("\n📦 Phase: %s\n", strings.ToUpper(p.Phase))
			lastPhase = p.Phase
		}

		// Calculate ETA
		elapsed := p.ElapsedTime
		var eta time.Duration
		if p.Percentage > 0 {
			totalEstimate := time.Duration(float64(elapsed) * 100.0 / p.Percentage)
			eta = totalEstimate - elapsed
		}

		// Render progress bar
		bar := progressBar.Render(p.Percentage)

		// Print detailed progress
		fmt.Printf("%s | Row %d/%d | %d bytes | Elapsed: %s | ETA: %s",
			bar,
			p.CurrentRow,
			p.TotalRows,
			p.BytesWritten,
			elapsed.Round(time.Second),
			eta.Round(time.Second),
		)
	}

	// Create programmer with progress callback
	prog := bootloader.New(device,
		bootloader.WithProgressCallback(progressCallback),
		bootloader.WithVerifyAfterProgram(true),
	)

	// Program the device
	fmt.Println("Starting programming sequence...")
	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}

	ctx := context.Background()
	err = prog.Program(ctx, fw, key)

	// Final status
	fmt.Println() // New line after progress
	fmt.Println()

	if err != nil {
		fmt.Printf("❌ Programming failed: %v\n", err)
		os.Exit(1)
	}

	totalTime := time.Since(startTime)
	avgSpeed := float64(len(fw.Rows)) / totalTime.Seconds()

	fmt.Println("✅ Programming completed successfully!")
	fmt.Printf("   Total time: %s\n", totalTime.Round(time.Millisecond))
	fmt.Printf("   Average speed: %.1f rows/sec\n", avgSpeed)
	fmt.Println("\nDevice is ready to use.")
}
