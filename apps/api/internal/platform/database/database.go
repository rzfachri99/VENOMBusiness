package database

import (
	"context"
	"database/sql"
	"fmt"
)

func Open(ctx context.Context, driver, dsn string) (*sql.DB, error) {
	switch driver {
	case "sqlite":
		return OpenSQLite(ctx, dsn)
	case "postgres":
		return OpenPostgres(ctx, dsn)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", driver)
	}
}
