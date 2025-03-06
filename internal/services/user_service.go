package services

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"unicode"

	"github.com/cam-boltnote/go-ignite/internal/connectors"
	"github.com/cam-boltnote/go-ignite/internal/models"
	"github.com/cam-boltnote/go-ignite/internal/utils"

	"gorm.io/gorm"
)

// getPasswordLengthConfig loads password length configuration from environment variables
// with fallback default values
func getPasswordLengthConfig() (min, max int) {
	min = 8  // default minimum
	max = 72 // default maximum

	if minStr := os.Getenv("PASSWORD_MIN_LENGTH"); minStr != "" {
		if val, err := strconv.Atoi(minStr); err == nil {
			min = val
		}
	}

	if maxStr := os.Getenv("PASSWORD_MAX_LENGTH"); maxStr != "" {
		if val, err := strconv.Atoi(maxStr); err == nil {
			max = val
		}
	}

	return min, max
}

// UserService handles user-related database operations and business logic
type UserService struct {
	db              *gorm.DB
	settingsService *SettingsService
	minPassLength   int
	maxPassLength   int
	emailSender     *connectors.EmailSender
	logger          *utils.Logger
}

// NewUserService creates a new user service instance
func NewUserService(db *gorm.DB) *UserService {
	minLength, maxLength := getPasswordLengthConfig()

	// Initialize logger
	logger := utils.GetLogger().WithService("user_service")

	// Initialize email sender
	emailSender, err := connectors.NewEmailSender()
	if err != nil {
		logger.Error("Failed to initialize email sender", err, nil)
	}

	return &UserService{
		db:              db,
		settingsService: NewSettingsService(db),
		minPassLength:   minLength,
		maxPassLength:   maxLength,
		emailSender:     emailSender,
		logger:          logger,
	}
}

// CreateUserInput represents the input for creating a new user
type CreateUserInput struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Role      string `json:"role,omitempty"`
}

// validatePassword validates password strength requirements
func (s *UserService) validatePassword(password string) error {
	if len(password) < s.minPassLength {
		return fmt.Errorf("password must be at least %d characters long", s.minPassLength)
	}
	if len(password) > s.maxPassLength {
		return fmt.Errorf("password must not exceed %d characters", s.maxPassLength)
	}

	// Check for at least one number
	hasNumber := false
	for _, char := range password {
		if unicode.IsNumber(char) {
			hasNumber = true
			break
		}
	}
	if !hasNumber {
		return errors.New("password must contain at least one number")
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, char := range password {
		if unicode.IsUpper(char) {
			hasUpper = true
			break
		}
	}
	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}

	return nil
}

// CreateUser creates a new user
func (s *UserService) CreateUser(input CreateUserInput) (*models.User, error) {
	s.logger.Info("Creating new user", map[string]interface{}{
		"email": input.Email,
	})

	// Check if user already exists
	existingUser, _ := s.GetByEmail(input.Email)
	if existingUser != nil {
		s.logger.Warn("User already exists", map[string]interface{}{
			"email": input.Email,
		})
		return nil, errors.New("user with this email already exists")
	}

	// Validate password
	if err := s.validatePassword(input.Password); err != nil {
		s.logger.Warn("Invalid password", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, fmt.Errorf("invalid password: %v", err)
	}

	// Create user with default role if not specified
	if input.Role == "" {
		input.Role = "user"
	}

	user := &models.User{
		Email:     input.Email,
		Password:  input.Password, // Note: Password should be hashed before storage
		FirstName: input.FirstName,
		LastName:  input.LastName,
		Role:      input.Role,
		IsActive:  true,
	}

	if err := s.db.Create(user).Error; err != nil {
		s.logger.Error("Failed to create user", err, map[string]interface{}{
			"email": input.Email,
		})
		return nil, fmt.Errorf("error creating user: %v", err)
	}

	// Create default settings for the new user
	if err := s.settingsService.CreateDefaultSettings(user.ID); err != nil {
		s.logger.Error("Failed to create default settings", err, map[string]interface{}{
			"user_id": user.ID,
		})
		// If settings creation fails, we should probably delete the user
		s.Delete(user.ID)
		return nil, fmt.Errorf("error creating default settings: %v", err)
	}

	// Send welcome email to the user
	if s.emailSender != nil {
		s.logger.Info("Sending welcome email", map[string]interface{}{
			"email": user.Email,
		})
		welcomeSubject := "Welcome to boltnote.ai!"
		welcomeBody := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #2c3e50;">Welcome to boltnote.ai, %s!</h2>
					<p>Thank you for joining us. We're excited to help you track and organize your activities.</p>
					<p>Get started by creating your first entry!</p>
					<p style="margin: 25px 0;">
						<a href="https://app.boltnote.ai" style="background-color: #3498db; color: white; padding: 12px 25px; text-decoration: none; border-radius: 4px;">Start Using Boltnote</a>
					</p>
					<p>If you have any questions, feel free to reach out to our support team.</p>
				</div>
			</body>
			</html>
		`, input.FirstName)

		if err := s.emailSender.SendEmail(user.Email, welcomeSubject, welcomeBody); err != nil {
			// Log the error but don't fail the user creation
			s.logger.Error("Failed to send welcome email", err, map[string]interface{}{
				"email": user.Email,
			})
		}

		// Send notification to admin
		adminEmail := os.Getenv("ADMIN_NOTIFICATION_EMAIL")
		if adminEmail == "" {
			s.logger.Warn("ADMIN_NOTIFICATION_EMAIL not set in environment", map[string]interface{}{
				"fallback": "cam@boltnote.ai",
			})
			adminEmail = "cam@boltnote.ai" // fallback value
		}

		// Add debug logging
		s.logger.Info("User details for admin email", map[string]interface{}{
			"firstName": user.FirstName,
			"lastName":  user.LastName,
			"email":     user.Email,
			"id":        user.ID,
			"createdAt": user.CreatedAt.Format("2006-01-02 15:04:05"),
		})

		adminSubject := fmt.Sprintf("New User Signup: %s %s", input.FirstName, input.LastName)
		adminBody := fmt.Sprintf(`
			<html>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px; background-color: #f9f9f9; border-radius: 8px;">
					<h2 style="color: #2c3e50; margin-bottom: 20px;">New User Registration</h2>
					<div style="background-color: white; padding: 20px; border-radius: 4px; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
						<table style="width: 100%%; border-collapse: collapse;">
							<tr>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;"><strong>Name:</strong></td>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;">%[1]s %[2]s</td>
							</tr>
							<tr>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;"><strong>Email:</strong></td>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;">%[3]s</td>
							</tr>
							<tr>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;"><strong>User ID:</strong></td>
								<td style="padding: 10px 0; border-bottom: 1px solid #eee;">%[4]d</td>
							</tr>
							<tr>
								<td style="padding: 10px 0;"><strong>Signup Time:</strong></td>
								<td style="padding: 10px 0;">%[5]s</td>
							</tr>
						</table>
					</div>
					<div style="margin-top: 20px; text-align: center;">
						<a href="https://admin.boltnote.ai/users/%[4]d" style="background-color: #3498db; color: white; padding: 12px 25px; text-decoration: none; border-radius: 4px; display: inline-block;">View User Details</a>
					</div>
				</div>
			</body>
			</html>
		`, user.FirstName, user.LastName, user.Email, user.ID, user.CreatedAt.Format("2006-01-02 15:04:05"))

		if err := s.emailSender.SendEmail(adminEmail, adminSubject, adminBody); err != nil {
			// Log the error but don't fail the user creation
			s.logger.Error("Failed to send admin notification", err, map[string]interface{}{
				"email": adminEmail,
			})
		}
	}

	// Don't return the password
	user.Password = ""
	return user, nil
}

// GetByID retrieves a user by their ID
func (s *UserService) GetByID(id uint) (*models.User, error) {
	s.logger.Debug("Fetching user by ID", map[string]interface{}{
		"id": id,
	})

	var user models.User
	result := s.db.First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to fetch user", result.Error, map[string]interface{}{
			"id": id,
		})
		return nil, result.Error
	}
	user.Password = "" // Don't return the password
	return &user, nil
}

// GetByEmail retrieves a user by their email
func (s *UserService) GetByEmail(email string) (*models.User, error) {
	s.logger.Debug("Fetching user by email", map[string]interface{}{
		"email": email,
	})

	var user models.User
	result := s.db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("User not found", map[string]interface{}{
				"email": email,
			})
			return nil, errors.New("user not found")
		}
		s.logger.Error("Failed to fetch user", result.Error, map[string]interface{}{
			"email": email,
		})
		return nil, result.Error
	}
	user.Password = "" // Don't return the password
	return &user, nil
}

// Update updates an existing user
func (s *UserService) Update(user *models.User) error {
	s.logger.Info("Updating user", map[string]interface{}{
		"id":    user.ID,
		"email": user.Email,
	})

	err := s.db.Save(user).Error
	if err != nil {
		s.logger.Error("Failed to update user", err, map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
		})
	}
	return err
}

// Delete deletes a user
func (s *UserService) Delete(id uint) error {
	s.logger.Info("Deleting user", map[string]interface{}{
		"id": id,
	})

	// Start a transaction since we're updating multiple tables
	tx := s.db.Begin()
	if tx.Error != nil {
		s.logger.Error("Failed to begin transaction", tx.Error, nil)
		return fmt.Errorf("failed to begin transaction: %v", tx.Error)
	}

	// First, soft delete the settings due to foreign key constraints
	if err := tx.Where("user_id = ?", id).Delete(&models.Settings{}).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete settings", err, map[string]interface{}{
			"user_id": id,
		})
		return fmt.Errorf("failed to delete settings: %v", err)
	}

	// Then soft delete the user
	if err := tx.Delete(&models.User{}, id).Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to delete user", err, map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("failed to delete user: %v", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		s.logger.Error("Failed to commit transaction", err, map[string]interface{}{
			"id": id,
		})
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	s.logger.Info("Successfully deleted user", map[string]interface{}{
		"id": id,
	})
	return nil
}

// List retrieves a list of users with pagination
func (s *UserService) List(page, pageSize int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	offset := (page - 1) * pageSize
	err := s.db.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = s.db.Offset(offset).Limit(pageSize).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	// Clear passwords from response
	for i := range users {
		users[i].Password = ""
	}

	return users, total, nil
}

// UpdatePassword updates a user's password
func (s *UserService) UpdatePassword(id uint, hashedPassword string) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

// Deactivate deactivates a user account
func (s *UserService) Deactivate(id uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", false).Error
}

// Activate activates a user account
func (s *UserService) Activate(id uint) error {
	return s.db.Model(&models.User{}).Where("id = ?", id).Update("is_active", true).Error
}

// ValidateCredentials validates user credentials
func (s *UserService) ValidateCredentials(email, password string) (*models.User, error) {
	user, err := s.GetByEmail(email)
	if err != nil {
		return nil, err
	}

	// Note: This is a placeholder. In a real application, you would:
	// 1. Hash the provided password
	// 2. Compare it with the stored hash
	// 3. Return appropriate errors
	if user.Password != password {
		return nil, errors.New("invalid credentials")
	}

	return user, nil
}
