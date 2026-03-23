package storage

import (
	"gorm.io/gorm"
)

// Migrate runs database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&User{},
		&Role{},
		&RoleBinding{},
		&AuditLog{},
		&ClusterConfig{},
	)
}
