# Progress Tracking Example

An advanced example demonstrating real-time progress tracking during firmware programming with visual feedback.

## What This Example Shows

- Visual progress bar rendering
- Detailed progress information
- Phase-based status updates
- Elapsed time tracking
- ETA (Estimated Time of Arrival) calculation
- Terminal-friendly output formatting

## Running the Example

```bash
cd examples/with_progress
go run main.go <firmware.cyacd>
```

Example:
```bash
go run main.go ../../testdata/firmware.cyacd
```

## Features

### Visual Progress Bar

The example displays a visual progress bar using Unicode block characters:

```
📦 Phase: PROGRAMMING
[████████████████████░░░░░░░░░░░░░░░░░░░░] 50.0% | Row 5/10 | 1024 bytes | Elapsed: 2s | ETA: 2s
```

### Phase Tracking

The programmer goes through several phases:

- **ENTERING** - Entering bootloader mode
- **PROGRAMMING** - Writing firmware rows
- **VERIFYING** - Verifying application checksum
- **EXITING** - Exiting bootloader and launching app
- **COMPLETE** - Programming finished

### Real-time Statistics

- Current phase and percentage
- Row count (current/total)
- Bytes written
- Elapsed time since start
- Estimated time remaining

## Code Highlights

### Progress Callback

```go
progressCallback := func(p bootloader.Progress) {
    // Update progress bar
    bar := progressBar.Render(p.Percentage)

    // Calculate ETA
    elapsed := p.ElapsedTime
    if p.Percentage > 0 {
        totalEstimate := time.Duration(float64(elapsed) * 100.0 / p.Percentage)
        eta = totalEstimate - elapsed
    }

    // Display formatted progress
    fmt.Printf("%s | Row %d/%d | %d bytes | Elapsed: %s | ETA: %s",
        bar, p.CurrentRow, p.TotalRows, p.BytesWritten, elapsed, eta)
}
```

### Progress Bar Rendering

```go
type ProgressBar struct {
    width int
}

func (pb *ProgressBar) Render(percentage float64) string {
    filled := int(float64(pb.width) * percentage / 100.0)
    bar := strings.Repeat("█", filled) + strings.Repeat("░", pb.width-filled)
    return fmt.Sprintf("[%s] %.1f%%", bar, percentage)
}
```

## Terminal Control

The example uses ANSI escape sequences for smooth terminal updates:

- `\r` - Carriage return (move to start of line)
- `\033[K` - Clear from cursor to end of line

This allows for non-flickering progress updates.

## Customization

### Adjust Progress Bar Width

```go
progressBar := NewProgressBar(60) // 60 character width
```

### Modify Output Format

Customize the printf format string to show different information:

```go
fmt.Printf("Programming: %.0f%% (%d/%d rows)",
    p.Percentage, p.CurrentRow, p.TotalRows)
```

### Add Color Support

You can add ANSI color codes for enhanced visualization:

```go
const (
    colorReset  = "\033[0m"
    colorGreen  = "\033[32m"
    colorYellow = "\033[33m"
    colorBlue   = "\033[34m"
)

fmt.Printf("%s[%s]%s %.1f%%", colorBlue, bar, colorReset, percentage)
```

## Performance Considerations

### Mock Device Latency

The example uses a mock device with simulated latency:

```go
func (d *MockDevice) Read(p []byte) (int, error) {
    time.Sleep(10 * time.Millisecond) // Simulate device latency
    return len(p), nil
}
```

With real hardware, the actual latency depends on:
- Communication method (USB, UART, etc.)
- Device response time
- Flash write speed

## Next Steps

- See [basic](../basic/) for simpler example without visual feedback
- See [advanced](../advanced/) for production configuration
- See [mock_device](../mock_device/) for realistic device simulation
- Read the [API documentation](https://pkg.go.dev/github.com/moffa90/go-cyacd) for more options

## Integration Ideas

This progress tracking approach can be integrated with:

- GUI applications (update progress bars)
- Web interfaces (send progress via WebSocket)
- CLI tools (terminal progress bars)
- Logging systems (structured log events)
- Monitoring systems (metrics/telemetry)
