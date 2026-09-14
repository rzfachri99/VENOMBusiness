package main

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/venombusiness/venombusiness/apps/api/internal/platform/database"
	"github.com/venombusiness/venombusiness/apps/api/internal/platform/migrations"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	driver := getenv("VENOM_DATABASE_DRIVER", "sqlite")
	dsn := getenv("VENOM_DATABASE_URL", "file:venombusiness.db?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)")
	switch os.Args[1] {
	case "backup":
		out := argOr(2, fmt.Sprintf("venombusiness-%s.backup", time.Now().Format("20060102-150405")))
		must(backup(driver, dsn, out))
		fmt.Println("backup created:", out)
	case "restore":
		if len(os.Args) < 3 {
			fatal("restore requires a backup path")
		}
		must(restore(driver, dsn, os.Args[2]))
		fmt.Println("restore completed")
	case "migrate-to-postgres":
		if driver != "sqlite" {
			fatal("migrate-to-postgres requires VENOM_DATABASE_DRIVER=sqlite")
		}
		target := os.Getenv("VENOM_TARGET_DATABASE_URL")
		if target == "" {
			fatal("set VENOM_TARGET_DATABASE_URL to the destination postgres:// URL")
		}
		must(migrateToPostgres(dsn, target))
		fmt.Println("SQLite -> PostgreSQL migration completed")
	default:
		usage()
		os.Exit(2)
	}
}

func migrateToPostgres(sourceDSN, targetDSN string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	src, err := database.Open(ctx, "sqlite", sourceDSN)
	if err != nil {
		return err
	}
	defer src.Close()
	dst, err := database.Open(ctx, "postgres", targetDSN)
	if err != nil {
		return err
	}
	defer dst.Close()
	if err := migrations.Up(ctx, dst, "postgres"); err != nil {
		return err
	}
	var count int
	if err := dst.QueryRowContext(ctx, `SELECT COUNT(*) FROM companies`).Scan(&count); err != nil {
		return err
	}
	if count != 0 {
		return fmt.Errorf("destination PostgreSQL database is not empty")
	}
	tables := []string{"users", "companies", "company_members", "sessions", "customers", "products", "document_sequences", "quotations", "quotation_items", "invoices", "invoice_items", "payments", "expense_categories", "expenses", "audit_logs"}
	tx, err := dst.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for _, table := range tables {
		if err := copyTable(ctx, src, tx, table); err != nil {
			return fmt.Errorf("copy %s: %w", table, err)
		}
	}
	return tx.Commit()
}
func copyTable(ctx context.Context, src *sql.DB, dst *sql.Tx, table string) error {
	rows, err := src.QueryContext(ctx, "SELECT * FROM "+table)
	if err != nil {
		return err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return err
	}
	marks := make([]string, len(cols))
	for i := range marks {
		marks[i] = "?"
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)", table, strings.Join(cols, ","), strings.Join(marks, ","))
	for rows.Next() {
		vals := make([]any, len(cols))
		ptr := make([]any, len(cols))
		for i := range vals {
			ptr[i] = &vals[i]
		}
		if err := rows.Scan(ptr...); err != nil {
			return err
		}
		if _, err := dst.ExecContext(ctx, q, vals...); err != nil {
			return err
		}
	}
	return rows.Err()
}
func backup(driver, dsn, out string) error {
	if driver == "postgres" {
		cmd := exec.Command("pg_dump", "--format=custom", "--file", out, dsn)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	_, err := sqlitePath(dsn)
	if err != nil {
		return err
	}
	db, err := database.Open(context.Background(), "sqlite", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	_ = os.Remove(out)
	safe := strings.ReplaceAll(out, "'", "''")
	_, err = db.Exec("VACUUM INTO '" + safe + "'")
	return err
}
func restore(driver, dsn, in string) error {
	if driver == "postgres" {
		cmd := exec.Command("pg_restore", "--clean", "--if-exists", "--no-owner", "--dbname", dsn, in)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	dst, err := sqlitePath(dsn)
	if err != nil {
		return err
	}
	return copyFile(in, dst)
}
func sqlitePath(dsn string) (string, error) {
	if !strings.HasPrefix(dsn, "file:") {
		return "", fmt.Errorf("SQLite backup requires a file: DSN")
	}
	x := strings.TrimPrefix(strings.SplitN(dsn, "?", 2)[0], "file:")
	if x == "" || x == ":memory:" {
		return "", fmt.Errorf("cannot backup in-memory SQLite")
	}
	return filepath.Clean(x), nil
}
func copyFile(src, dst string) error {
	in, e := os.Open(src)
	if e != nil {
		return e
	}
	defer in.Close()
	out, e := os.Create(dst)
	if e != nil {
		return e
	}
	defer out.Close()
	_, e = io.Copy(out, in)
	return e
}
func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
func argOr(i int, d string) string {
	if len(os.Args) > i {
		return os.Args[i]
	}
	return d
}
func must(e error) {
	if e != nil {
		fatal(e.Error())
	}
}
func fatal(s string) { fmt.Fprintln(os.Stderr, "error:", s); os.Exit(1) }
func usage() {
	fmt.Println("VENOMBusiness admin utility\n  venomctl backup [path]\n  venomctl restore <path>\n  venomctl migrate-to-postgres\n\nPostgreSQL backup/restore requires pg_dump/pg_restore in PATH. Stop the API before SQLite file backup/restore.")
}
