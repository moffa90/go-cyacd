package main

import (
	"fmt"
	"github.com/moffa90/go-cyacd/cyacd"
)

func main() {
	fw, err := cyacd.Parse("/Users/joseluismoffa/Downloads/hec-2-splt-dl-v0.6.0-7c817399.cyacd")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	
	fmt.Printf("Total rows: %d\n", len(fw.Rows))
	if len(fw.Rows) > 0 {
		row0 := fw.Rows[0]
		fmt.Printf("Row 0: ArrayID=%d, RowNum=%d, DataLen=%d\n", 
			row0.ArrayID, row0.RowNum, len(row0.Data))
		fmt.Printf("First 20 bytes: %02X\n", row0.Data[:20])
	}
}
