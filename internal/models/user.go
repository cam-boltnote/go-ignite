package models

import (
	"errors"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	BaseModel
	Email     string `gorm:"unique;not null" json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Password  string `gorm:"not null" json:"-"`
	IsActive  bool   `gorm:"default:true" json:"is_active"`
	Role      string `gorm:"default:'user'" json:"role"` // Common roles: 'user', 'admin', 'moderator'
}

// UserService handles user-related database operations
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new user service instance
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Create creates a new user
func (s *UserService) Create(user *User) error {
	return s.db.Create(user).Error
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(id uint) (*User, error) {
	var user User
	result := s.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetByEmail retrieves a user by their email
func (s *UserService) GetByEmail(email string) (*User, error) {
	var user User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update updates an existing user
func (s *UserService) Update(user *User) error {
	return s.db.Save(user).Error
}

// Delete deletes a user
func (s *UserService) Delete(id uint) error {
	return s.db.Delete(&User{}, id).Error
}

// List retrieves a list of users with pagination
func (s *UserService) List(page, pageSize int) ([]User, int64, error) {
	var users []User
	var total int64

	offset := (page - 1) * pageSize
	err := s.db.Model(&User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(id uint, hashedPassword string) error {
	return s.db.Model(&User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// Deactivate deactivates a user account
func (s *UserService) Deactivate(id uint) error {
	return s.db.Model(&User{}).Where("id = ?", id).Update("is_active", false).Error
}

// Activate activates a user account
func (s *UserService) Activate(id uint) error {
	return s.db.Model(&User{}).Where("id = ?", id).Update("is_active", true).Error
}
