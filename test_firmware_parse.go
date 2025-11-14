package main

import (
	"fmt"
	"log"
	"os"

	"github.com/moffa90/go-cyacd/cyacd"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run test_firmware_parse.go <firmware.cyacd>")
		os.Exit(1)
	}

	firmwarePath := os.Args[1]
	fmt.Printf("Parsing firmware file: %s\n\n", firmwarePath)

	fw, err := cyacd.Parse(firmwarePath)
	if err != nil {
		log.Fatalf("Failed to parse firmware: %v", err)
	}

	fmt.Println("✅ Firmware parsed successfully!")
	fmt.Println()
	fmt.Printf("Firmware Information:\n")
	fmt.Printf("  Silicon ID:     0x%08X\n", fw.SiliconID)
	fmt.Printf("  Silicon Rev:    0x%02X\n", fw.SiliconRev)
	fmt.Printf("  Checksum Type:  0x%02X", fw.ChecksumType)
	if fw.ChecksumType == 0x00 {
		fmt.Printf(" (Basic Summation)\n")
	} else if fw.ChecksumType == 0x01 {
		fmt.Printf(" (CRC-16)\n")
	} else {
		fmt.Printf(" (Unknown)\n")
	}
	fmt.Printf("  Total Rows:     %d\n", len(fw.Rows))
	fmt.Println()

	// Display first few rows
	displayRows := 5
	if len(fw.Rows) < displayRows {
		displayRows = len(fw.Rows)
	}

	fmt.Printf("First %d rows:\n", displayRows)
	for i := 0; i < displayRows; i++ {
		row := fw.Rows[i]
		fmt.Printf("  Row %d:\n", i)
		fmt.Printf("    Array ID:  0x%02X\n", row.ArrayID)
		fmt.Printf("    Row Num:   0x%04X (%d)\n", row.RowNum, row.RowNum)
		fmt.Printf("    Data Len:  %d bytes\n", len(row.Data))
		fmt.Printf("    Checksum:  0x%02X\n", row.Checksum)

		// Display first 16 bytes of data
		dataDisplay := len(row.Data)
		if dataDisplay > 16 {
			dataDisplay = 16
		}
		fmt.Printf("    Data:      % 02X", row.Data[:dataDisplay])
		if len(row.Data) > 16 {
			fmt.Printf("... (%d more bytes)", len(row.Data)-16)
		}
		fmt.Println()
	}

	if len(fw.Rows) > displayRows {
		fmt.Printf("\n... and %d more rows\n", len(fw.Rows)-displayRows)
	}

	// Calculate total bytes
	totalBytes := 0
	for _, row := range fw.Rows {
		totalBytes += len(row.Data)
	}

	fmt.Println()
	fmt.Printf("Statistics:\n")
	fmt.Printf("  Total data: %d bytes (%.2f KB)\n", totalBytes, float64(totalBytes)/1024.0)
	fmt.Printf("  Average row size: %.1f bytes\n", float64(totalBytes)/float64(len(fw.Rows)))
	fmt.Println()
	fmt.Println("✅ Firmware file is valid and ready for programming!")
}
