package excel

import (
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

type Table struct {
	Headers []string
	Rows    []map[string]any // key: header label (Japanese OK)
}

type ReadOptions struct {
	SheetName    string
	HeaderRow    int // 1-based
	DataStartRow int // 1-based
	TrimHeader   bool
	SkipEmptyKey bool // skip columns with empty header
}

// ReadTable reads a sheet as "header + rows" table.
func ReadTable(f *excelize.File, opt ReadOptions) (*Table, error) {
	if opt.HeaderRow <= 0 {
		return nil, fmt.Errorf("HeaderRow must be >= 1")
	}
	if opt.DataStartRow <= 0 {
		return nil, fmt.Errorf("DataStartRow must be >= 1")
	}
	if opt.DataStartRow <= opt.HeaderRow {
		return nil, fmt.Errorf("DataStartRow must be > HeaderRow")
	}

	sheet := opt.SheetName
	if sheet == "" {
		sheets := f.GetSheetList()
		if len(sheets) == 0 {
			return nil, fmt.Errorf("no sheets")
		}
		sheet = sheets[0]
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("GetRows failed: %w", err)
	}
	if len(rows) < opt.HeaderRow {
		return nil, fmt.Errorf("sheet has no header row: want %d, got %d", opt.HeaderRow, len(rows))
	}

	rawHeaders := rows[opt.HeaderRow-1]
	headers := make([]string, len(rawHeaders))
	for i, h := range rawHeaders {
		if opt.TrimHeader {
			h = strings.TrimSpace(h)
		}
		headers[i] = h
	}

	out := &Table{
		Headers: headers,
		Rows:    make([]map[string]any, 0),
	}

	for r := opt.DataStartRow - 1; r < len(rows); r++ {
		row := rows[r]
		record := make(map[string]any)

		// Map each cell to header key
		for c, key := range headers {
			if key == "" && opt.SkipEmptyKey {
				continue
			}
			var val any = nil
			if c < len(row) {
				cell := row[c]
				if cell != "" {
					val = cell
				}
			}
			record[key] = val
		}

		// Optionally skip fully empty rows (all nil)
		allNil := true
		for _, v := range record {
			if v != nil {
				allNil = false
				break
			}
		}
		if allNil {
			continue
		}

		out.Rows = append(out.Rows, record)
	}

	return out, nil
}
