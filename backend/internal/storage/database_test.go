package storage

import (
	"path/filepath"
	"testing"

	"github.com/google/uuid"
)

func setupTestDB(t *testing.T) *Database {
	t.Helper()

	tmpFile := filepath.Join(t.TempDir(), "test.db")
	cfg := DatabaseConfig{
		Type:    "sqlite",
		DSN:     tmpFile,
		MaxIdle: 5,
		MaxOpen: 25,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	return db
}

func TestNewDatabase_SQLite(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.db")
	cfg := DatabaseConfig{
		Type:    "sqlite",
		DSN:     tmpFile,
		MaxIdle: 5,
		MaxOpen: 25,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify tables exist
	tables := []string{"users", "roles", "role_bindings", "audit_logs"}
	for _, table := range tables {
		if !db.db.Migrator().HasTable(&User{}) {
			t.Errorf("Table %s does not exist", table)
		}
	}
}

func TestNewDatabase_DefaultsToSQLite(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test_default.db")
	cfg := DatabaseConfig{
		DSN: tmpFile,
	}

	db, err := NewDatabase(cfg)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Error("Expected non-nil database")
	}
}

func TestUserRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := db.UserRepository()

	user := &User{
		ID:            uuid.New(),
		Username:      "testuser",
		Email:         "test@example.com",
		OAuthProvider: "github",
		OAuthSub:      "123456",
	}

	// Create
	err := repo.Create(user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	// GetByID
	found, err := repo.GetByID(user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}
	if found.Username != user.Username {
		t.Errorf("Expected username %s, got %s", user.Username, found.Username)
	}

	// GetByUsername
	found, err = repo.GetByUsername(user.Username)
	if err != nil {
		t.Fatalf("Failed to get user by username: %v", err)
	}
	if found.Email != user.Email {
		t.Errorf("Expected email %s, got %s", user.Email, found.Email)
	}

	// GetByOAuth
	found, err = repo.GetByOAuth(user.OAuthProvider, user.OAuthSub)
	if err != nil {
		t.Fatalf("Failed to get user by OAuth: %v", err)
	}
	if found.ID != user.ID {
		t.Errorf("Expected ID %s, got %s", user.ID, found.ID)
	}

	// GetAll
	users, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all users: %v", err)
	}
	if len(users) != 1 {
		t.Errorf("Expected 1 user, got %d", len(users))
	}

	// Update
	user.Email = "updated@example.com"
	err = repo.Update(user)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}
	found, _ = repo.GetByID(user.ID)
	if found.Email != "updated@example.com" {
		t.Errorf("Expected updated email, got %s", found.Email)
	}

	// Delete
	err = repo.Delete(user.ID)
	if err != nil {
		t.Fatalf("Failed to delete user: %v", err)
	}
	_, err = repo.GetByID(user.ID)
	if err == nil {
		t.Error("Expected error when getting deleted user")
	}
}

func TestRoleRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := db.RoleRepository()

	role := &Role{
		ID:          uuid.New(),
		Name:        "test-role",
		Description: "Test role",
		Rules:       `[]`,
		IsBuiltin:   false,
	}

	// Create
	err := repo.Create(role)
	if err != nil {
		t.Fatalf("Failed to create role: %v", err)
	}

	// GetByID
	found, err := repo.GetByID(role.ID)
	if err != nil {
		t.Fatalf("Failed to get role by ID: %v", err)
	}
	if found.Name != role.Name {
		t.Errorf("Expected name %s, got %s", role.Name, found.Name)
	}

	// GetByName
	found, err = repo.GetByName(role.Name)
	if err != nil {
		t.Fatalf("Failed to get role by name: %v", err)
	}
	if found.Description != role.Description {
		t.Errorf("Expected description %s, got %s", role.Description, found.Description)
	}

	// GetAll
	roles, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all roles: %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("Expected 1 role, got %d", len(roles))
	}
}

func TestRoleBindingRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := db.RoleBindingRepository()

	userID := uuid.New()
	roleID := uuid.New()

	binding := &RoleBinding{
		ID:         uuid.New(),
		UserID:     userID,
		RoleID:     roleID,
		ClusterIDs: `["cluster-a", "cluster-b"]`,
		Namespaces: `["default", "production"]`,
	}

	// Create
	err := repo.Create(binding)
	if err != nil {
		t.Fatalf("Failed to create role binding: %v", err)
	}

	// GetByID
	found, err := repo.GetByID(binding.ID)
	if err != nil {
		t.Fatalf("Failed to get binding by ID: %v", err)
	}
	if found.RoleID != roleID {
		t.Errorf("Expected role ID %s, got %s", roleID, found.RoleID)
	}

	// GetByUserID
	bindings, err := repo.GetByUserID(userID)
	if err != nil {
		t.Fatalf("Failed to get bindings by user ID: %v", err)
	}
	if len(bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(bindings))
	}

	// GetByRoleID
	bindings, err = repo.GetByRoleID(roleID)
	if err != nil {
		t.Fatalf("Failed to get bindings by role ID: %v", err)
	}
	if len(bindings) != 1 {
		t.Errorf("Expected 1 binding, got %d", len(bindings))
	}
}

func TestAuditLogRepository(t *testing.T) {
	db := setupTestDB(t)
	repo := db.AuditLogRepository()

	userID := uuid.New()
	log := &AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       "create",
		Resource:     "pods",
		ResourceName: "api-server",
		Namespace:    "default",
		ClusterID:    "cluster-a",
		Status:       "success",
		Details:      `{"key": "value"}`,
	}

	// Create
	err := repo.Create(log)
	if err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// GetByUserID
	logs, err := repo.GetByUserID(userID, 10)
	if err != nil {
		t.Fatalf("Failed to get logs by user ID: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	// GetByResource
	logs, err = repo.GetByResource("pods", "default", "cluster-a", 10)
	if err != nil {
		t.Fatalf("Failed to get logs by resource: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}

	// GetAll
	logs, err = repo.GetAll(10)
	if err != nil {
		t.Fatalf("Failed to get all logs: %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("Expected 1 log, got %d", len(logs))
	}
}
