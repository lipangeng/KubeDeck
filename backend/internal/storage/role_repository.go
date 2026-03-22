package storage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleRepository defines the interface for role data access
type RoleRepository interface {
	Create(*Role) error
	GetByID(id uuid.UUID) (*Role, error)
	GetByName(name string) (*Role, error)
	GetAll() ([]Role, error)
	Update(*Role) error
	Delete(id uuid.UUID) error
}

type roleRepository struct {
	db *gorm.DB
}

func (r *roleRepository) Create(role *Role) error {
	return r.db.Create(role).Error
}

func (r *roleRepository) GetByID(id uuid.UUID) (*Role, error) {
	var role Role
	err := r.db.First(&role, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetByName(name string) (*Role, error) {
	var role Role
	err := r.db.Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) GetAll() ([]Role, error) {
	var roles []Role
	err := r.db.Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Update(role *Role) error {
	return r.db.Save(role).Error
}

func (r *roleRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&Role{ID: id}).Error
}
