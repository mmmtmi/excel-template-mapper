package excel

import (
	"fmt"

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
		return nil, fmt.Errorf("ヘッダー位置の指定は1行目以上にしてください。")
	}
	if opt.DataStartRow <= 0 {
		return nil, fmt.Errorf("データ開始位置の指定は1行目以上にしてください。")
	}
	if opt.DataStartRow <= opt.HeaderRow {
		return nil, fmt.Errorf("データの開始位置は2行目からにしてください。")
	}

	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		fmt.Errorf("シートの取得に失敗しました。")
	}
	fmt.Println("sheets:", sheets)

	sheet := sheets[0]

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("GetRows failed: %w", err)
	}

	rawHeaders := rows[opt.HeaderRow-1]
	headers := make([]string, len(rawHeaders))
	for i, h := range rawHeaders {
		headers[i] = h
	}

	out := &Table{
		Headers: headers,
		Rows:    make([]map[string]any, 0),
	}

	for r := opt.DataStartRow - 1; r < len(rows); r++ {
		row := rows[r]
		record := make(map[string]any)
		for c, key := range headers {
			var val any = nil
			if c < len(row) {
				cell := row[c]
				if cell != "" {
					val = cell
				}
			}
			record[key] = val
		}

		out.Rows = append(out.Rows, record)
	}
	return out, nil

}
