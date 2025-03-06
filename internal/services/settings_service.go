package services

import (
	"errors"
	"fmt"

	"github.com/cam-boltnote/go-ignite/internal/models"
	"github.com/cam-boltnote/go-ignite/internal/utils"

	"gorm.io/gorm"
)

// SettingsService handles settings-related database operations and business logic
type SettingsService struct {
	db     *gorm.DB
	logger *utils.Logger
}

// NewSettingsService creates a new settings service instance
func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{
		db:     db,
		logger: utils.GetLogger().WithService("settings_service"),
	}
}

// CreateDefaultSettings creates a new settings entry with default values for a user
func (s *SettingsService) CreateDefaultSettings(userID uint) error {
	s.logger.Info("Creating default settings", map[string]interface{}{
		"user_id": userID,
	})

	settings := &models.Settings{
		UserID: userID,
		// Default values are set in the model
	}

	if err := s.db.Create(settings).Error; err != nil {
		s.logger.Error("Failed to create default settings", err, map[string]interface{}{
			"user_id": userID,
		})
		return fmt.Errorf("failed to create default settings: %v", err)
	}

	return nil
}

// GetByID retrieves settings by their ID
func (s *SettingsService) GetByID(id uint) (*models.Settings, error) {
	s.logger.Debug("Fetching settings by ID", map[string]interface{}{
		"id": id,
	})

	var settings models.Settings
	result := s.db.First(&settings, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Settings not found", map[string]interface{}{
				"id": id,
			})
			return nil, errors.New("settings not found")
		}
		s.logger.Error("Failed to fetch settings", result.Error, map[string]interface{}{
			"id": id,
		})
		return nil, result.Error
	}
	return &settings, nil
}

// GetByUserID retrieves settings by user ID
func (s *SettingsService) GetByUserID(userID uint) (*models.Settings, error) {
	s.logger.Debug("Fetching settings by user ID", map[string]interface{}{
		"user_id": userID,
	})

	var settings models.Settings
	result := s.db.Where("user_id = ?", userID).First(&settings)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			s.logger.Warn("Settings not found", map[string]interface{}{
				"user_id": userID,
			})
			return nil, gorm.ErrRecordNotFound
		}
		s.logger.Error("Failed to fetch settings", result.Error, map[string]interface{}{
			"user_id": userID,
		})
		return nil, result.Error
	}
	return &settings, nil
}

// Update updates existing settings
func (s *SettingsService) Update(settings *models.Settings) error {
	s.logger.Info("Updating settings", map[string]interface{}{
		"user_id": settings.UserID,
	})

	result := s.db.Where("user_id = ?", settings.UserID).Updates(settings)
	if result.Error != nil {
		s.logger.Error("Failed to update settings", result.Error, map[string]interface{}{
			"user_id": settings.UserID,
		})
		return result.Error
	}
	if result.RowsAffected == 0 {
		s.logger.Warn("No settings found to update", map[string]interface{}{
			"user_id": settings.UserID,
		})
		return errors.New("no settings found for this user")
	}
	return nil
}

// Delete deletes settings
func (s *SettingsService) Delete(id uint) error {
	s.logger.Info("Deleting settings", map[string]interface{}{
		"id": id,
	})

	result := s.db.Delete(&models.Settings{}, id)
	if result.Error != nil {
		s.logger.Error("Failed to delete settings", result.Error, map[string]interface{}{
			"id": id,
		})
		return result.Error
	}
	if result.RowsAffected == 0 {
		s.logger.Warn("No settings found to delete", map[string]interface{}{
			"id": id,
		})
	}
	return nil
}

// UpdateCustomSettings updates only the custom settings for a user
func (s *SettingsService) UpdateCustomSettings(userID uint, customSettings map[string]interface{}) error {
	s.logger.Info("Updating custom settings", map[string]interface{}{
		"user_id":  userID,
		"settings": customSettings,
	})

	err := s.db.Model(&models.Settings{}).
		Where("user_id = ?", userID).
		Update("custom_settings", customSettings).Error

	if err != nil {
		s.logger.Error("Failed to update custom settings", err, map[string]interface{}{
			"user_id": userID,
		})
	}
	return err
}

// GetCustomSetting retrieves a specific custom setting
func (s *SettingsService) GetCustomSetting(userID uint, key string) (interface{}, error) {
	s.logger.Debug("Fetching custom setting", map[string]interface{}{
		"user_id": userID,
		"key":     key,
	})

	var settings models.Settings
	err := s.db.Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		s.logger.Error("Failed to fetch custom setting", err, map[string]interface{}{
			"user_id": userID,
			"key":     key,
		})
		return nil, err
	}
	if settings.CustomSettings == nil {
		s.logger.Debug("No custom settings found", map[string]interface{}{
			"user_id": userID,
		})
		return nil, nil
	}
	return settings.CustomSettings[key], nil
}

// UpdateNotificationSettings updates notification preferences
func (s *SettingsService) UpdateNotificationSettings(userID uint, emailEnabled, pushEnabled bool, frequency string) error {
	s.logger.Info("Updating notification settings", map[string]interface{}{
		"user_id":       userID,
		"email_enabled": emailEnabled,
		"push_enabled":  pushEnabled,
		"frequency":     frequency,
	})

	err := s.db.Model(&models.Settings{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"email_notifications_enabled": emailEnabled,
			"push_notifications_enabled":  pushEnabled,
			"notification_frequency":      frequency,
		}).Error

	if err != nil {
		s.logger.Error("Failed to update notification settings", err, map[string]interface{}{
			"user_id": userID,
		})
	}
	return err
}

// UpdatePrivacySettings updates privacy preferences
func (s *SettingsService) UpdatePrivacySettings(userID uint, visibility string, dataSharing bool) error {
	s.logger.Info("Updating privacy settings", map[string]interface{}{
		"user_id":      userID,
		"visibility":   visibility,
		"data_sharing": dataSharing,
	})

	err := s.db.Model(&models.Settings{}).
		Where("user_id = ?", userID).
		Updates(map[string]interface{}{
			"profile_visibility": visibility,
			"data_sharing":       dataSharing,
		}).Error

	if err != nil {
		s.logger.Error("Failed to update privacy settings", err, map[string]interface{}{
			"user_id": userID,
		})
	}
	return err
}

// UpdateGeneralSettings updates general preferences
func (s *SettingsService) UpdateGeneralSettings(userID uint, timezone, language, theme string) error {
	s.logger.Info("Updating general settings", map[string]interface{}{
		"user_id":  userID,
		"timezone": timezone,
		"language": language,
		"theme":    theme,
	})

	updates := make(map[string]interface{})
	if timezone != "" {
		updates["timezone"] = timezone
	}
	if language != "" {
		updates["language"] = language
	}
	if theme != "" {
		updates["theme"] = theme
	}

	if len(updates) == 0 {
		s.logger.Debug("No general settings to update", map[string]interface{}{
			"user_id": userID,
		})
		return nil
	}

	err := s.db.Model(&models.Settings{}).
		Where("user_id = ?", userID).
		Updates(updates).Error

	if err != nil {
		s.logger.Error("Failed to update general settings", err, map[string]interface{}{
			"user_id": userID,
		})
	}
	return err
}
