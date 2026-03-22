package storage

import (
	"time"

	"github.com/google/uuid"
)

// User represents a platform user
type User struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Username      string    `gorm:"uniqueIndex;size:255" json:"username"`
	Email         string    `gorm:"size:255" json:"email"`
	OAuthProvider string    `gorm:"column:oauth_provider;size:100" json:"oauth_provider"`
	OAuthSub      string    `gorm:"column:oauth_sub;size:255" json:"oauth_sub"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// Role represents a RBAC role
type Role struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:255" json:"name"`
	Description string    `gorm:"size:1024" json:"description"`
	Rules       string    `gorm:"type:json" json:"rules"` // JSON serialized PolicyRules
	IsBuiltin   bool      `gorm:"default:false" json:"is_builtin"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RoleBinding binds users to roles
type RoleBinding struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	RoleID     uuid.UUID `gorm:"type:uuid;index" json:"role_id"`
	UserID     uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	ClusterIDs string    `gorm:"type:json" json:"cluster_ids"` // JSON array of cluster IDs
	Namespaces string    `gorm:"type:json" json:"namespaces"`  // JSON array or "*"
	CreatedAt  time.Time `json:"created_at"`
}

// AuditLog records user actions
type AuditLog struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID       uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	Action       string    `gorm:"size:100;index" json:"action"`
	Resource     string    `gorm:"size:100;index" json:"resource"`
	ResourceName string    `gorm:"size:255" json:"resource_name"`
	Namespace    string    `gorm:"size:255;index" json:"namespace"`
	ClusterID    string    `gorm:"size:255;index" json:"cluster_id"`
	Status       string    `gorm:"size:20;index" json:"status"`
	Timestamp    time.Time `gorm:"index" json:"timestamp"`
	Details      string    `gorm:"type:json" json:"details"`
}
