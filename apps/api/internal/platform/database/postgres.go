package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/lib/pq"
)

var registerPostgresOnce sync.Once

// OpenPostgres opens PostgreSQL while preserving the application's portable ? placeholders.
// The lightweight driver adapter converts ? to $1, $2, ... before queries reach lib/pq.
func OpenPostgres(ctx context.Context, dsn string) (*sql.DB, error) {
	registerPostgresOnce.Do(func() { sql.Register("venom-postgres", rebindDriver{inner: &pq.Driver{}}) })
	db, err := sql.Open("venom-postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return db, nil
}

type rebindDriver struct{ inner driver.Driver }

func (d rebindDriver) Open(name string) (driver.Conn, error) {
	c, err := d.inner.Open(name)
	if err != nil {
		return nil, err
	}
	return &rebindConn{Conn: c}, nil
}

type rebindConn struct{ driver.Conn }

func (c *rebindConn) Prepare(q string) (driver.Stmt, error) { return c.Conn.Prepare(rebind(q)) }
func (c *rebindConn) PrepareContext(ctx context.Context, q string) (driver.Stmt, error) {
	if x, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return x.PrepareContext(ctx, rebind(q))
	}
	return c.Prepare(q)
}
func (c *rebindConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	if x, ok := c.Conn.(driver.ExecerContext); ok {
		return x.ExecContext(ctx, rebind(q), args)
	}
	return nil, driver.ErrSkip
}
func (c *rebindConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	if x, ok := c.Conn.(driver.QueryerContext); ok {
		return x.QueryContext(ctx, rebind(q), args)
	}
	return nil, driver.ErrSkip
}
func (c *rebindConn) Ping(ctx context.Context) error {
	if x, ok := c.Conn.(driver.Pinger); ok {
		return x.Ping(ctx)
	}
	return nil
}
func (c *rebindConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if x, ok := c.Conn.(driver.ConnBeginTx); ok {
		return x.BeginTx(ctx, opts)
	}
	return c.Conn.Begin()
}
func (c *rebindConn) CheckNamedValue(v *driver.NamedValue) error {
	if x, ok := c.Conn.(driver.NamedValueChecker); ok {
		return x.CheckNamedValue(v)
	}
	return driver.ErrSkip
}

func rebind(q string) string {
	var b strings.Builder
	b.Grow(len(q) + 8)
	n := 1
	inSingle := false
	for i := 0; i < len(q); i++ {
		ch := q[i]
		if ch == '\'' {
			inSingle = !inSingle
			b.WriteByte(ch)
			continue
		}
		if ch == '?' && !inSingle {
			fmt.Fprintf(&b, "$%d", n)
			n++
			continue
		}
		b.WriteByte(ch)
	}
	return b.String()
}
