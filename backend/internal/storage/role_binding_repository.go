package storage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RoleBindingRepository defines the interface for role binding data access
type RoleBindingRepository interface {
	Create(*RoleBinding) error
	GetByID(id uuid.UUID) (*RoleBinding, error)
	GetByUserID(userID uuid.UUID) ([]RoleBinding, error)
	GetByRoleID(roleID uuid.UUID) ([]RoleBinding, error)
	GetAll() ([]RoleBinding, error)
	Update(*RoleBinding) error
	Delete(id uuid.UUID) error
}

type roleBindingRepository struct {
	db *gorm.DB
}

func (r *roleBindingRepository) Create(binding *RoleBinding) error {
	return r.db.Create(binding).Error
}

func (r *roleBindingRepository) GetByID(id uuid.UUID) (*RoleBinding, error) {
	var binding RoleBinding
	err := r.db.First(&binding, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &binding, nil
}

func (r *roleBindingRepository) GetByUserID(userID uuid.UUID) ([]RoleBinding, error) {
	var bindings []RoleBinding
	err := r.db.Where("user_id = ?", userID).Find(&bindings).Error
	return bindings, err
}

func (r *roleBindingRepository) GetByRoleID(roleID uuid.UUID) ([]RoleBinding, error) {
	var bindings []RoleBinding
	err := r.db.Where("role_id = ?", roleID).Find(&bindings).Error
	return bindings, err
}

func (r *roleBindingRepository) GetAll() ([]RoleBinding, error) {
	var bindings []RoleBinding
	err := r.db.Find(&bindings).Error
	return bindings, err
}

func (r *roleBindingRepository) Update(binding *RoleBinding) error {
	return r.db.Save(binding).Error
}

func (r *roleBindingRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&RoleBinding{ID: id}).Error
}
