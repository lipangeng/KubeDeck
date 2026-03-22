package storage

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type    string `json:"type"` // sqlite, mysql, postgres
	DSN     string `json:"dsn"`
	MaxIdle int    `json:"max_idle"`
	MaxOpen int    `json:"max_open"`
}

// Database wraps gorm.DB with repository accessors
type Database struct {
	db *gorm.DB
}

// NewDatabase creates and initializes a database connection
func NewDatabase(cfg DatabaseConfig) (*Database, error) {
	var dialector gorm.Dialector

	// Set defaults
	if cfg.Type == "" {
		cfg.Type = "sqlite"
	}
	if cfg.DSN == "" && cfg.Type == "sqlite" {
		// Default to data directory in project root
		cfg.DSN = getDefaultSQLitePath()
	}
	if cfg.MaxIdle == 0 {
		cfg.MaxIdle = 5
	}
	if cfg.MaxOpen == 0 {
		cfg.MaxOpen = 25
	}

	switch cfg.Type {
	case "mysql":
		dialector = mysql.Open(cfg.DSN)
	case "postgres":
		dialector = postgres.Open(cfg.DSN)
	case "sqlite", "":
		dialector = sqlite.Open(cfg.DSN)
	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Auto migrate schema
	err = db.AutoMigrate(&User{}, &Role{}, &RoleBinding{}, &AuditLog{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate schema: %w", err)
	}

	// Connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdle)
	sqlDB.SetMaxOpenConns(cfg.MaxOpen)

	return &Database{db: db}, nil
}

// DB returns the underlying gorm.DB
func (d *Database) DB() *gorm.DB {
	return d.db
}

// Close closes the database connection
func (d *Database) Close() error {
	sqlDB, err := d.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// getDefaultSQLitePath returns the default SQLite database path
func getDefaultSQLitePath() string {
	// Try to find project root (look for go.mod)
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			// Not found, use current directory
			dir = "."
			break
		}
		dir = parent
	}

	dataDir := filepath.Join(dir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		// Fallback to current directory
		return "kubedeck.db"
	}

	return filepath.Join(dataDir, "kubedeck.db")
}

// UserRepository returns the user repository
func (d *Database) UserRepository() UserRepository {
	return &userRepository{db: d.db}
}

// RoleRepository returns the role repository
func (d *Database) RoleRepository() RoleRepository {
	return &roleRepository{db: d.db}
}

// RoleBindingRepository returns the role binding repository
func (d *Database) RoleBindingRepository() RoleBindingRepository {
	return &roleBindingRepository{db: d.db}
}

// AuditLogRepository returns the audit log repository
func (d *Database) AuditLogRepository() AuditLogRepository {
	return &auditLogRepository{db: d.db}
}
