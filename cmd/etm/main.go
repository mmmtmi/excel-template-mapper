package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/mmmtmi/excel-template-mapper/internal/dbconn"
	"github.com/mmmtmi/excel-template-mapper/internal/excel"
	"github.com/mmmtmi/excel-template-mapper/internal/model"
	"github.com/mmmtmi/excel-template-mapper/internal/store/mysql"
)

func main() {

	templateName := flag.String("template", "", "template name")
	flag.Parse()

	excelPath := ""

	if flag.NArg() >= 1 {
		excelPath = flag.Arg(0)
	}

	if excelPath == "" && *templateName == "" {
		log.Fatal("usage: etm [--template demo_v1] [excel-file]")
	}

	var table *excel.Table
	if excelPath != "" {
		f, err := excelize.OpenFile(excelPath)
		if err != nil {
			log.Fatalf("open failed: %v", err)
		}
		defer func() { _ = f.Close() }()

		table, err = excel.ReadTable(f, excel.ReadOptions{
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

	var rules []model.Rule
	if *templateName != "" {

		// env読み込み
		cfg, err := dbconn.LoadMySQLConfigFromEnv(".env")
		if err != nil {
			log.Fatal(err)
		}

		// db 接続
		db, err := dbconn.Open(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()

		// ctx 初期化
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// mapping template の取得

		// dbヘルスチェック
		if err := dbconn.Ping(ctx, db); err != nil {
			log.Fatal(err)
		}
		log.Println("データベースに正常に接続されました！")

		//簡単な確認
		one, err := dbconn.SelectOne(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("SELECT 1 => %d", one)

		//テーブルの確認
		tables, err := dbconn.ListTables(ctx, db)
		if err != nil {
			log.Fatal(err)
		}
		for _, t := range tables {
			log.Printf("table: %s", t)
		}

		// テーブルの取り出し
		tpl, err := mysql.GetTemplateByName(ctx, db, *templateName)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("template: id=%s name=%s target=%s sheet=%v header_row=%d data_start_row=%d",
			tpl.ID, tpl.Name, tpl.Target, tpl.SheetName, tpl.HeaderRow, tpl.DataStartRow)

		templateID := tpl.ID
		rules, err = mysql.ListRulesByTemplateID(ctx, db, templateID)
		if err != nil {
			log.Fatal(err)
		}

		for _, r := range rules {
			log.Printf("rule: %s %s -> %s transform=%v required=%t priority=%d",
				r.SourceType, r.SourceKey, r.TargetLabel, r.Transform, r.Required, r.Priority)
		}
	}

	if table != nil && len(rules) > 0 {
		keyAndLabel := make(map[string]string)
		for _, r := range rules {
			if r.SourceType == "HEADER" {
				keyAndLabel[r.SourceKey] = r.TargetLabel
			}
		}
		for key, value := range keyAndLabel {
			log.Printf("key=%s,value=%s", key, value)
		}

		outRows := make([]map[string]any, 0, len(table.Rows))
		for _, row := range table.Rows {
			outRow := make(map[string]any)
			for c, targetLabel := range keyAndLabel {
				outRow[targetLabel] = row[c]
			}
			outRows = append(outRows, outRow)
		}

		b, err := json.MarshalIndent(outRows, "", "  ")
		if err != nil {
			log.Fatalf("json marshal failed: %v", err)
		}
		fmt.Println(string(b))

	}

}
