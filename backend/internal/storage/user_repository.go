package storage

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(*User) error
	GetByID(id uuid.UUID) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByOAuth(provider, sub string) (*User, error)
	GetAll() ([]User, error)
	Update(*User) error
	Delete(id uuid.UUID) error
}

type userRepository struct {
	db *gorm.DB
}

func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetByID(id uuid.UUID) (*User, error) {
	var user User
	err := r.db.First(&user, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*User, error) {
	var user User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByOAuth(provider, sub string) (*User, error) {
	var user User
	err := r.db.Where("oauth_provider = ? AND oauth_sub = ?", provider, sub).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll() ([]User, error) {
	var users []User
	err := r.db.Find(&users).Error
	return users, err
}

func (r *userRepository) Update(user *User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&User{ID: id}).Error
}
