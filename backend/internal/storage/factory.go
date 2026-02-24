package storage

import (
	"fmt"
	"strings"
)

func NewStore(driver string, dsn string) (Store, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "", "sqlite":
		return newSQLiteStore(dsn), nil
	case "mysql":
		return newMySQLStore(dsn), nil
	case "postgres":
		return newPostgresStore(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported storage driver %q", driver)
	}
}
