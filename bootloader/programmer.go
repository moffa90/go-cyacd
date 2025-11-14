package bootloader

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/moffa90/go-cyacd/cyacd"
	"github.com/moffa90/go-cyacd/protocol"
)

// Programmer orchestrates firmware programming operations for Cypress microcontrollers.
// It handles the complete programming sequence including verification and progress tracking.
//
// Programmer is safe for concurrent use after initialization.
type Programmer struct {
	device io.ReadWriter
	config Config
}

// New creates a new Programmer with the given device and options.
// The device must implement io.ReadWriter for communication with the bootloader.
//
// Example:
//
//	device := myusb.OpenDevice("serial-number")
//	prog := bootloader.New(device,
//	    bootloader.WithProgressCallback(progressFunc),
//	    bootloader.WithTimeout(30*time.Second),
//	)
func New(device io.ReadWriter, opts ...Option) *Programmer {
	if device == nil {
		panic("device cannot be nil")
	}

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	return &Programmer{
		device: device,
		config: cfg,
	}
}

// Program performs the complete firmware programming sequence:
//  1. Enter bootloader with the provided key
//  2. Validate device silicon ID matches firmware
//  3. Get flash size and validate all rows are in range
//  4. Program all rows with progress tracking
//  5. Verify application checksum
//  6. Exit bootloader
//
// The operation can be cancelled via context.
//
// Example:
//
//	fw, _ := cyacd.Parse("firmware.cyacd")
//	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}
//	err := prog.Program(context.Background(), fw, key)
func (p *Programmer) Program(ctx context.Context, fw *cyacd.Firmware, key []byte) error {
	if fw == nil {
		return fmt.Errorf("firmware cannot be nil")
	}
	if len(key) != 6 {
		return fmt.Errorf("key must be exactly 6 bytes, got %d", len(key))
	}

	startTime := time.Now()

	// Phase 1: Enter bootloader
	p.reportProgress(Progress{
		Phase:      "entering",
		Percentage: 0,
		TotalRows:  len(fw.Rows),
	})

	deviceInfo, err := p.EnterBootloader(ctx, key)
	if err != nil {
		return fmt.Errorf("enter bootloader: %w", err)
	}

	p.logDebug("entered bootloader",
		"silicon_id", fmt.Sprintf("0x%08X", deviceInfo.SiliconID),
		"silicon_rev", fmt.Sprintf("0x%02X", deviceInfo.SiliconRev),
		"bootloader_ver", fmt.Sprintf("%d.%d.%d",
			deviceInfo.BootloaderVer[0], deviceInfo.BootloaderVer[1], deviceInfo.BootloaderVer[2]),
	)

	// Phase 2: Validate device silicon ID
	if deviceInfo.SiliconID != fw.SiliconID {
		return &DeviceMismatchError{
			Expected: fw.SiliconID,
			Actual:   deviceInfo.SiliconID,
		}
	}

	// Phase 3: Get flash size and validate rows
	p.reportProgress(Progress{
		Phase:      "programming",
		Percentage: 2,
		TotalRows:  len(fw.Rows),
	})

	// Validate all rows are in range (check first row's array)
	if len(fw.Rows) > 0 {
		flashSize, err := p.GetFlashSize(ctx, fw.Rows[0].ArrayID)
		if err != nil {
			return fmt.Errorf("get flash size: %w", err)
		}

		p.logDebug("flash size",
			"array_id", fw.Rows[0].ArrayID,
			"start_row", flashSize.StartRow,
			"end_row", flashSize.EndRow,
		)

		// Validate all rows are in range
		for _, row := range fw.Rows {
			if row.RowNum < flashSize.StartRow || row.RowNum > flashSize.EndRow {
				return &RowOutOfRangeError{
					RowNum: row.RowNum,
					MinRow: flashSize.StartRow,
					MaxRow: flashSize.EndRow,
				}
			}
		}
	}

	// Phase 4: Program rows
	bytesWritten := 0
	for i, row := range fw.Rows {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}

		if err := p.programRow(ctx, row); err != nil {
			return fmt.Errorf("program row %d (array=%d, row=%d): %w",
				i, row.ArrayID, row.RowNum, err)
		}

		// Verify if enabled
		if p.config.VerifyAfterProgram {
			if err := p.verifyRow(ctx, row); err != nil {
				return fmt.Errorf("verify row %d (array=%d, row=%d): %w",
					i, row.ArrayID, row.RowNum, err)
			}
		}

		bytesWritten += len(row.Data)

		// Report progress (2% to 90%)
		percentage := 2 + (float64(i+1)/float64(len(fw.Rows)))*88
		p.reportProgress(Progress{
			Phase:        "programming",
			CurrentRow:   i + 1,
			TotalRows:    len(fw.Rows),
			Percentage:   percentage,
			BytesWritten: bytesWritten,
			ElapsedTime:  time.Since(startTime),
		})
	}

	// Phase 5: Verify application checksum
	p.reportProgress(Progress{
		Phase:       "verifying",
		CurrentRow:  len(fw.Rows),
		TotalRows:   len(fw.Rows),
		Percentage:  92,
		ElapsedTime: time.Since(startTime),
	})

	if err := p.VerifyChecksum(ctx); err != nil {
		return fmt.Errorf("verify application: %w", err)
	}

	// Phase 6: Exit bootloader
	p.reportProgress(Progress{
		Phase:       "exiting",
		CurrentRow:  len(fw.Rows),
		TotalRows:   len(fw.Rows),
		Percentage:  95,
		ElapsedTime: time.Since(startTime),
	})

	if err := p.ExitBootloader(ctx); err != nil {
		return fmt.Errorf("exit bootloader: %w", err)
	}

	// Complete
	p.reportProgress(Progress{
		Phase:        "complete",
		CurrentRow:   len(fw.Rows),
		TotalRows:    len(fw.Rows),
		Percentage:   100,
		BytesWritten: bytesWritten,
		ElapsedTime:  time.Since(startTime),
	})

	p.logInfo("programming complete",
		"rows", len(fw.Rows),
		"bytes", bytesWritten,
		"elapsed", time.Since(startTime).String(),
	)

	return nil
}

// programRow programs a single flash row, handling data chunking if necessary.
func (p *Programmer) programRow(ctx context.Context, row *cyacd.Row) error {
	chunkSize := p.config.ChunkSize
	data := row.Data

	// If data is larger than chunk size, send in multiple chunks
	for len(data) > chunkSize {
		chunk := data[:chunkSize]
		if err := p.sendData(ctx, chunk); err != nil {
			return fmt.Errorf("send data chunk: %w", err)
		}
		data = data[chunkSize:]
	}

	// Program the final chunk (or entire row if small enough)
	cmd, err := protocol.BuildProgramRowCmd(row.ArrayID, row.RowNum, data)
	if err != nil {
		return err
	}

	// Send command and wait for response
	response, err := p.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return err
	}

	// Check for success status
	statusCode, _, err := protocol.ParseResponse(response)
	if err != nil {
		return err
	}

	if statusCode != protocol.StatusSuccess {
		return &protocol.ProtocolError{StatusCode: statusCode}
	}

	return nil
}

// verifyRow verifies a programmed row's checksum.
func (p *Programmer) verifyRow(ctx context.Context, row *cyacd.Row) error {
	cmd, err := protocol.BuildVerifyRowCmd(row.ArrayID, row.RowNum)
	if err != nil {
		return err
	}

	response, err := p.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return err
	}

	statusCode, data, err := protocol.ParseResponse(response)
	if err != nil {
		return err
	}

	if statusCode != protocol.StatusSuccess {
		return &protocol.ProtocolError{
			Operation:  "verify row",
			StatusCode: statusCode,
		}
	}

	deviceChecksum, err := protocol.ParseVerifyRowResponse(data)
	if err != nil {
		return err
	}

	// Calculate expected checksum: row checksum + metadata
	expectedChecksum := protocol.CalculateRowChecksum(row.Data)
	if deviceChecksum != expectedChecksum {
		return &ChecksumMismatchError{
			RowNum:   row.RowNum,
			Expected: expectedChecksum,
			Actual:   deviceChecksum,
		}
	}

	return nil
}

// sendData sends a data chunk using the Send Data command.
func (p *Programmer) sendData(ctx context.Context, data []byte) error {
	cmd, err := protocol.BuildSendDataCmd(data)
	if err != nil {
		return err
	}

	return p.sendCommand(ctx, cmd)
}

// EnterBootloader sends the Enter Bootloader command with the specified key.
// Returns device identification information.
//
// Example:
//
//	key := []byte{0x0A, 0x1B, 0x2C, 0x3D, 0x4E, 0x5F}
//	info, err := prog.EnterBootloader(ctx, key)
func (p *Programmer) EnterBootloader(ctx context.Context, key []byte) (*protocol.DeviceInfo, error) {
	cmd, err := protocol.BuildEnterBootloaderCmd(key)
	if err != nil {
		return nil, err
	}

	response, err := p.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return nil, err
	}

	statusCode, data, err := protocol.ParseResponse(response)
	if err != nil {
		return nil, err
	}

	if statusCode != protocol.StatusSuccess {
		return nil, &protocol.ProtocolError{
			Operation:  "enter bootloader",
			StatusCode: statusCode,
		}
	}

	return protocol.ParseEnterBootloaderResponse(data)
}

// ExitBootloader sends the Exit Bootloader command.
// The bootloader will verify the application and reset the device.
func (p *Programmer) ExitBootloader(ctx context.Context) error {
	cmd, err := protocol.BuildExitBootloaderCmd()
	if err != nil {
		return err
	}

	// Exit bootloader may not send a response (device resets)
	_ = p.sendCommand(ctx, cmd)
	return nil
}

// GetFlashSize queries the valid flash row range for the specified array.
func (p *Programmer) GetFlashSize(ctx context.Context, arrayID byte) (*protocol.FlashSize, error) {
	cmd, err := protocol.BuildGetFlashSizeCmd(arrayID)
	if err != nil {
		return nil, err
	}

	response, err := p.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return nil, err
	}

	statusCode, data, err := protocol.ParseResponse(response)
	if err != nil {
		return nil, err
	}

	if statusCode != protocol.StatusSuccess {
		return nil, &protocol.ProtocolError{
			Operation:  "get flash size",
			StatusCode: statusCode,
		}
	}

	return protocol.ParseGetFlashSizeResponse(data)
}

// VerifyChecksum verifies the entire application checksum.
func (p *Programmer) VerifyChecksum(ctx context.Context) error {
	cmd, err := protocol.BuildVerifyChecksumCmd()
	if err != nil {
		return err
	}

	response, err := p.sendCommandWithResponse(ctx, cmd)
	if err != nil {
		return err
	}

	statusCode, data, err := protocol.ParseResponse(response)
	if err != nil {
		return err
	}

	if statusCode != protocol.StatusSuccess {
		return &protocol.ProtocolError{
			Operation:  "verify checksum",
			StatusCode: statusCode,
		}
	}

	valid, err := protocol.ParseVerifyChecksumResponse(data)
	if err != nil {
		return err
	}

	if !valid {
		return &VerificationError{
			Message: "application checksum is invalid",
		}
	}

	return nil
}

// sendCommand sends a command and expects no response (fire-and-forget).
func (p *Programmer) sendCommand(ctx context.Context, cmd []byte) error {
	_, err := p.device.Write(cmd)
	return err
}

// sendCommandWithResponse sends a command and waits for a response.
func (p *Programmer) sendCommandWithResponse(ctx context.Context, cmd []byte) ([]byte, error) {
	// Write command
	if _, err := p.device.Write(cmd); err != nil {
		return nil, fmt.Errorf("write command: %w", err)
	}

	// Read response (assume max response size of 512 bytes)
	response := make([]byte, 512)
	n, err := p.device.Read(response)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return response[:n], nil
}

// reportProgress calls the progress callback if configured.
func (p *Programmer) reportProgress(progress Progress) {
	if p.config.ProgressCallback != nil {
		p.config.ProgressCallback(progress)
	}
}

// logDebug logs a debug message if a logger is configured.
func (p *Programmer) logDebug(msg string, keysAndValues ...interface{}) {
	if p.config.Logger != nil {
		p.config.Logger.Debug(msg, keysAndValues...)
	}
}

// logInfo logs an info message if a logger is configured.
func (p *Programmer) logInfo(msg string, keysAndValues ...interface{}) {
	if p.config.Logger != nil {
		p.config.Logger.Info(msg, keysAndValues...)
	}
}

// logError logs an error message if a logger is configured.
func (p *Programmer) logError(msg string, keysAndValues ...interface{}) {
	if p.config.Logger != nil {
		p.config.Logger.Error(msg, keysAndValues...)
	}
}
