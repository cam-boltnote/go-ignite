package models

import (
	"errors"

	"gorm.io/gorm"
)

// Settings represents user-specific application settings
type Settings struct {
	BaseModel
	UserID uint `gorm:"uniqueIndex;not null" json:"user_id"`
	User   User `gorm:"constraint:OnDelete:CASCADE;" json:"-"`

	// General Settings
	Timezone string `gorm:"default:'UTC'" json:"timezone"` // User's preferred timezone
	Language string `gorm:"default:'en'" json:"language"`  // User's preferred language
	Theme    string `gorm:"default:'light'" json:"theme"`  // UI theme preference

	// Notification Settings
	EmailNotificationsEnabled bool   `gorm:"default:true" json:"email_notifications_enabled"`
	PushNotificationsEnabled  bool   `gorm:"default:true" json:"push_notifications_enabled"`
	NotificationFrequency     string `gorm:"default:'daily'" json:"notification_frequency"` // daily, weekly, monthly

	// Privacy Settings
	ProfileVisibility string `gorm:"default:'private'" json:"profile_visibility"` // private, public, friends
	DataSharing       bool   `gorm:"default:false" json:"data_sharing"`           // Whether to share usage data

	// Custom Settings (JSON field for application-specific settings)
	CustomSettings map[string]interface{} `gorm:"type:json" json:"custom_settings"`
}

// SettingsService handles settings-related database operations
type SettingsService struct {
	db *gorm.DB
}

// NewSettingsService creates a new settings service instance
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

// Create creates new settings for a user
func (s *SettingsService) Create(settings *Settings) error {
	return s.db.Create(settings).Error
}

// GetByID retrieves settings by their ID
func (s *SettingsService) GetByID(id uint) (*Settings, error) {
	var settings Settings
	result := s.db.First(&settings, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, errors.New("settings not found")
		}
		return nil, result.Error
	}
	return &settings, nil
}

// GetByUserID retrieves settings by user ID
func (s *SettingsService) GetByUserID(userID uint) (*Settings, error) {
	var settings Settings
	result := s.db.Where("user_id = ?", userID).First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, gorm.ErrRecordNotFound
		}
		return nil, result.Error
	}
	return &settings, nil
}

// Update updates existing settings
func (s *SettingsService) Update(settings *Settings) error {
	result := s.db.Where("user_id = ?", settings.UserID).Updates(settings)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("no settings found for this user")
	}
	return nil
}

// Delete deletes settings
func (s *SettingsService) Delete(id uint) error {
	return s.db.Delete(&Settings{}, id).Error
}

// UpdateCustomSettings updates only the custom settings for a user
func (s *SettingsService) UpdateCustomSettings(userID uint, customSettings map[string]interface{}) error {
	return s.db.Model(&Settings{}).
		Where("user_id = ?", userID).
		Update("custom_settings", customSettings).Error
}

// GetCustomSetting retrieves a specific custom setting
func (s *SettingsService) GetCustomSetting(userID uint, key string) (interface{}, error) {
	var settings Settings
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		return nil, err
	}
	if settings.CustomSettings == nil {
		return nil, nil
	}
	return settings.CustomSettings[key], nil
}

// UpdateNotificationSettings updates notification preferences
func (s *SettingsService) UpdateNotificationSettings(userID uint, emailEnabled, pushEnabled bool, frequency string) error {
	return s.db.Model(&Settings{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_notifications_enabled": emailEnabled,
			"push_notifications_enabled":  pushEnabled,
			"notification_frequency":      frequency,
		}).Error
}

// UpdatePrivacySettings updates privacy preferences
func (s *SettingsService) UpdatePrivacySettings(userID uint, visibility string, dataSharing bool) error {
	return s.db.Model(&Settings{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"profile_visibility": visibility,
			"data_sharing":       dataSharing,
		}).Error
}
