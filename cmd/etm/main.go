package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xuri/excelize/v2"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: etm < excel-file>")
	}
	path := os.Args[1]

	f, err := excelize.OpenFile(path)
	if err != nil {
		log.Fatalf("open failed: %v", err)
	}

	defer func() {
		_ = f.Close()
	}()

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		log.Fatal("no sheets")
	}
	fmt.Println("sheets:", sheets)

	rows, err := f.GetRows(sheets[0])
	if err != nil {
		log.Fatalf("get rows failed: %v", err)
	}

	max := 5
	if len(rows) < max {
		max = len(rows)
	}
	for i := 0; i < max; i++ {
		fmt.Printf("row %d: %#v\n", i+1, rows[i])
	}
}
