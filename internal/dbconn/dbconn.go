package dbconn

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
)

// LoadMySQLConfigFromEnv loads .env (optional) and builds a mysql.Config from env vars.
//
// Required env:
//   - SQL_USER
//   - SQL_PASSWORD
//   - SQL_ADDR (e.g. 127.0.0.1:3306)
//   - SQL_DBNAME
//
// Optional env:
//   - SQL_NET (defaults to tcp). For backward compatibility, SQL_NST is also accepted.
func LoadMySQLConfigFromEnv(dotenvPath string) (mysql.Config, error) {
	if dotenvPath != "" {
		// Make .env behavior explicit: if you pass a path, it should exist.
		if err := godotenv.Load(dotenvPath); err != nil {
			return mysql.Config{}, fmt.Errorf("load env file %q: %w", dotenvPath, err)
		}
	}

	user := strings.TrimSpace(os.Getenv("SQL_USER"))
	pass := os.Getenv("SQL_PASSWORD")
	net := strings.TrimSpace(os.Getenv("SQL_NET"))
	if net == "" {
		// Older typo seen in this repo.
		net = strings.TrimSpace(os.Getenv("SQL_NST"))
	}
	if net == "" {
		net = "tcp"
	}
	addr := strings.TrimSpace(os.Getenv("SQL_ADDR"))
	dbname := strings.TrimSpace(os.Getenv("SQL_DBNAME"))

	var missing []string
	if user == "" {
		missing = append(missing, "SQL_USER")
	}
	if pass == "" {
		missing = append(missing, "SQL_PASSWORD")
	}
	if addr == "" {
		missing = append(missing, "SQL_ADDR")
	}
	if dbname == "" {
		missing = append(missing, "SQL_DBNAME")
	}
	if len(missing) > 0 {
		return mysql.Config{}, fmt.Errorf("missing env: %s", strings.Join(missing, ", "))
	}

	return mysql.Config{
		User:                 user,
		Passwd:               pass,
		Net:                  net,
		Addr:                 addr,
		DBName:               dbname,
		AllowNativePasswords: true,
		ParseTime:            true,
		Loc:                  time.Local,
	}, nil
}

// Open opens a DB handle using the mysql driver.
// Note: sql.Open does not verify connectivity; call Ping/PingContext yourself.
func Open(cfg mysql.Config) (*sql.DB, error) {
	return sql.Open("mysql", cfg.FormatDSN())
}

func Ping(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}

// 簡単な確認
func SelectOne(ctx context.Context, db *sql.DB) (int, error) {
	var one int
	if err := db.QueryRowContext(ctx, "SELECT 1").Scan(&one); err != nil {
		return 0, err
	}
	return one, nil
}

// テーブルの確認
func ListTables(ctx context.Context, db *sql.DB) ([]string, error) {
	rows, err := db.QueryContext(ctx, "SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
