package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"

	"github.com/mmmtmi/excel-template-mapper/internal/excel"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: etm <excel-file>")
	}
	path := os.Args[1]

	f, err := excelize.OpenFile(path)
	if err != nil {
		log.Fatalf("open failed: %v", err)
	}
	defer func() { _ = f.Close() }()

	table, err := excel.ReadTable(f, excel.ReadOptions{
		HeaderRow:    1,
		DataStartRow: 2,
		TrimHeader:   true,
		SkipEmptyKey: true,
	})
	if err != nil {
		log.Fatalf("read table failed: %v", err)
	}

	// JSON pretty print
	b, err := json.MarshalIndent(table.Rows, "", "  ")
	if err != nil {
		log.Fatalf("json marshal failed: %v", err)
	}
	fmt.Println(string(b))
}
