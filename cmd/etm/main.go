package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"

	"github.com/mmmtmi/excel-template-mapper/internal/dbconn"
	"github.com/mmmtmi/excel-template-mapper/internal/excel"
	"github.com/mmmtmi/excel-template-mapper/internal/model"
	"github.com/mmmtmi/excel-template-mapper/internal/store/mysql"
)

func main() {

	templateName := flag.String("template", "", "template name")
	debugFlag := flag.Bool("debug", false, "enable debug logs")
	flag.Parse()

	excelPath := ""

	if flag.NArg() >= 1 {
		excelPath = flag.Arg(0)
	}

	if excelPath == "" && *templateName == "" {
		log.Fatal("usage: etm [--template demo_v1] [excel-file]")
	}

	var table *excel.Table
	var readOptions = &excel.ReadOptions{
		HeaderRow:    1,
		DataStartRow: 2,
		TrimHeader:   true,
		SkipEmptyKey: true,
	}
	if excelPath != "" {
		f, err := excelize.OpenFile(excelPath)
		if err != nil {
			log.Fatalf("open failed: %v", err)
		}
		defer func() { _ = f.Close() }()

		table, err = excel.ReadTable(f, *readOptions)
		if err != nil {
			log.Fatalf("read table failed: %v", err)
		}

		if *debugFlag {

			// JSON pretty print
			b, err := json.MarshalIndent(table.Rows, "", "  ")
			if err != nil {
				log.Fatalf("json marshal failed: %v", err)
			}
			fmt.Println(string(b))
		}
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

		// テーブルの取り出し
		tpl, err := mysql.GetTemplateByName(ctx, db, *templateName)
		if err != nil {
			log.Fatal(err)
		}

		templateID := tpl.ID
		rules, err = mysql.ListRulesByTemplateID(ctx, db, templateID)
		if err != nil {
			log.Fatal(err)
		}

		if *debugFlag {

			// tpl.SheetNameの整形
			var ts any = nil
			if tpl.SheetName != nil {
				ts = *tpl.SheetName
			}
			log.Printf("template: id=%s name=%s target=%s sheet=%v header_row=%d data_start_row=%d",
				tpl.ID, tpl.Name, tpl.Target, ts, tpl.HeaderRow, tpl.DataStartRow)

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

			for _, r := range rules {
				// r.Transformの整形。
				var tr any = nil
				if r.Transform != nil {
					tr = *r.Transform
				}
				log.Printf("rule: %s %s -> %s transform=%v required=%t priority=%d",
					r.SourceType, r.SourceKey, r.TargetLabel, tr, r.Required, r.Priority)
			}
		}
	}

	if table != nil && len(rules) > 0 {
		// requiredがTrueのとき、対象のヘッダーが無いとエラーになる。
		set := make(map[string]bool)
		for _, th := range table.Headers {
			set[th] = true
		}
		var requiredRules []model.Rule
		for _, r := range rules {
			if r.Required {
				requiredRules = append(requiredRules, r)

				if !set[r.SourceKey] {
					log.Fatalf("required header missing: %s", r.SourceKey)
				}
			}
		}

		// 一旦表示
		// for key, value := range keyAndLabel {
		// 	log.Printf("key=%s,value=%s", key, value)
		// }

		outRows := make([]map[string]any, 0, len(table.Rows))
		for i, row := range table.Rows {
			// requiredがTrueのとき、値がないとエラーになる
			for _, r := range requiredRules {
				if r.SourceType == "HEADER" {
					if row[r.SourceKey] == nil {
						log.Fatalf("required value missing: row=%d header=%s label=%s",
							readOptions.DataStartRow+i, r.SourceKey, r.TargetLabel)
					}
				}
			}

			outRow := make(map[string]any)
			for _, r := range rules {
				if r.SourceType != "HEADER" {
					continue
				}

				val := row[r.SourceKey]
				if r.Transform != nil && *r.Transform == "trim" {
					if s, ok := val.(string); ok {
						val = strings.TrimSpace(s)
					}
				}

				outRow[r.TargetLabel] = val
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
