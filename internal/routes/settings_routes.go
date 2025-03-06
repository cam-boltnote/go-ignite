package routes

import (
	"strconv"

	"github.com/cam-boltnote/go-ignite/internal/middleware"
	"github.com/cam-boltnote/go-ignite/internal/models"
	"github.com/cam-boltnote/go-ignite/internal/services"

	"github.com/gin-gonic/gin"
)

// SettingsRoutes handles all settings-related routes
type SettingsRoutes struct {
	settingsService *services.SettingsService
}

// NewSettingsRoutes creates a new settings routes instance
func NewSettingsRoutes(settingsService *services.SettingsService) *SettingsRoutes {
	return &SettingsRoutes{
		settingsService: settingsService,
	}
}

// RegisterRoutes registers protected settings-related routes
func (r *SettingsRoutes) RegisterRoutes(rg *gin.RouterGroup) {
	settings := rg.Group("/settings")
	{
		settings.OPTIONS("/:userId", middleware.CorsOptionsHandler)
		settings.GET("/:userId", r.GetSettings)
		settings.PUT("/:userId", r.UpdateSettings)

		settings.OPTIONS("/:userId/notifications", middleware.CorsOptionsHandler)
		settings.PUT("/:userId/notifications", r.UpdateNotificationSettings)

		settings.OPTIONS("/:userId/privacy", middleware.CorsOptionsHandler)
		settings.PUT("/:userId/privacy", r.UpdatePrivacySettings)

		settings.OPTIONS("/:userId/general", middleware.CorsOptionsHandler)
		settings.PUT("/:userId/general", r.UpdateGeneralSettings)

		settings.OPTIONS("/:userId/custom", middleware.CorsOptionsHandler)
		settings.PUT("/:userId/custom", r.UpdateCustomSettings)
		settings.GET("/:userId/custom/:key", r.GetCustomSetting)
	}
}

// GetSettings retrieves user settings
func (r *SettingsRoutes) GetSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	settings, err := r.settingsService.GetByUserID(uint(userID))
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, settings)
}

// UpdateSettings updates user settings
func (r *SettingsRoutes) UpdateSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var settings models.Settings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	settings.UserID = uint(userID)
	if err := r.settingsService.Update(&settings); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, settings)
}

// UpdateNotificationSettings updates notification preferences
func (r *SettingsRoutes) UpdateNotificationSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		EmailEnabled bool   `json:"email_enabled"`
		PushEnabled  bool   `json:"push_enabled"`
		Frequency    string `json:"frequency"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := r.settingsService.UpdateNotificationSettings(
		uint(userID),
		input.EmailEnabled,
		input.PushEnabled,
		input.Frequency,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Notification settings updated successfully"})
}

// UpdatePrivacySettings updates privacy preferences
func (r *SettingsRoutes) UpdatePrivacySettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		Visibility  string `json:"visibility"`
		DataSharing bool   `json:"data_sharing"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := r.settingsService.UpdatePrivacySettings(
		uint(userID),
		input.Visibility,
		input.DataSharing,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Privacy settings updated successfully"})
}

// UpdateGeneralSettings updates general preferences
func (r *SettingsRoutes) UpdateGeneralSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var input struct {
		Timezone string `json:"timezone"`
		Language string `json:"language"`
		Theme    string `json:"theme"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := r.settingsService.UpdateGeneralSettings(
		uint(userID),
		input.Timezone,
		input.Language,
		input.Theme,
	); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "General settings updated successfully"})
}

// UpdateCustomSettings updates custom settings
func (r *SettingsRoutes) UpdateCustomSettings(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	var customSettings map[string]interface{}
	if err := c.ShouldBindJSON(&customSettings); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := r.settingsService.UpdateCustomSettings(uint(userID), customSettings); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Custom settings updated successfully"})
}

// GetCustomSetting retrieves a specific custom setting
func (r *SettingsRoutes) GetCustomSetting(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("userId"), 10, 32)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid user ID"})
		return
	}

	key := c.Param("key")
	value, err := r.settingsService.GetCustomSetting(uint(userID), key)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"key": key, "value": value})
}
