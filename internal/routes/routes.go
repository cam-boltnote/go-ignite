package routes

import (
	"github.com/cam-boltnote/go-ignite/internal/connectors"
	"github.com/cam-boltnote/go-ignite/internal/middleware"
	"github.com/cam-boltnote/go-ignite/internal/services"

	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Routes struct {
	db             *gorm.DB
	emailSender    *connectors.EmailSender
	userRoutes     *UserRoutes
	settingsRoutes *SettingsRoutes
	testRoutes     *TestRoutes
}

func NewRoutes(db *gorm.DB, emailSender *connectors.EmailSender) *Routes {
	// Initialize test service and routes (always available)
	testService := services.NewTestService()
	testRoutes := NewTestRoutes(testService)

	// Initialize other services and routes only if dependencies are available
	var userRoutes *UserRoutes
	var settingsRoutes *SettingsRoutes

	if db != nil {
		userService := services.NewUserService(db)
		settingsService := services.NewSettingsService(db)
		userRoutes = NewUserRoutes(userService)
		settingsRoutes = NewSettingsRoutes(settingsService)
	} else {
		log.Println("Database functionality is disabled. User and settings routes will not be available.")
	}

	return &Routes{
		db:             db,
		emailSender:    emailSender,
		userRoutes:     userRoutes,
		settingsRoutes: settingsRoutes,
		testRoutes:     testRoutes,
	}
}

// RegisterRoutes registers all route groups with the router
func (r *Routes) RegisterRoutes(router *gin.Engine) {
	// Add CORS middleware
	router.Use(middleware.CORSMiddleware())

	// API versioning group
	v1 := router.Group("/api/v1")

	// Public routes (no auth required)
	if r.userRoutes != nil {
		r.userRoutes.RegisterPublicRoutes(v1)
		r.userRoutes.RegisterRoutes(v1)
	} else {
		// Register a placeholder route that returns a service unavailable message
		v1.GET("/user", func(c *gin.Context) {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": "User service is currently unavailable. Database functionality is disabled.",
			})
		})
	}

	// Protected routes (auth required)
	protected := v1.Group("")
	protected.Use(middleware.AuthMiddleware())
	{
		// Protected user routes
		if r.userRoutes != nil {
			r.userRoutes.RegisterRoutes(protected)
		}

		// Settings routes
		if r.settingsRoutes != nil {
			r.settingsRoutes.RegisterRoutes(protected)
		} else {
			// Register a placeholder route that returns a service unavailable message
			protected.GET("/settings", func(c *gin.Context) {
				c.JSON(http.StatusServiceUnavailable, gin.H{
					"error": "Settings service is currently unavailable. Database functionality is disabled.",
				})
			})
		}

		// Health check endpoint
		protected.GET("/health", func(c *gin.Context) {
			status := gin.H{
				"status": "ok",
				"services": gin.H{
					"database": r.db != nil,
					"email":    r.emailSender != nil && r.emailSender.IsEnabled(),
				},
			}
			c.JSON(http.StatusOK, status)
		})
	}
}
