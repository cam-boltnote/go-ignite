package main

import (
	"log"
	"os"

	"github.com/cam-boltnote/go-ignite/internal/connectors"
	"github.com/cam-boltnote/go-ignite/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
)

// @title           Go Activity Tracking API
// @version         1.0
// @description     A RESTful API service for tracking user activities with features for user management and settings management.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

// Add EmailSender to the application context
type AppContext struct {
	DB          *gorm.DB
	EmailSender *connectors.EmailSender
}

func setupRouter(ctx *AppContext) *gin.Engine {
	// Create default gin router
	router := gin.Default()

	// Update to pass both DB and EmailSender
	appRoutes := routes.NewRoutes(ctx.DB, ctx.EmailSender)
	appRoutes.RegisterRoutes(router)

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	// Initialize database connection (optional)
	var db *gorm.DB
	var err error
	database, err := connectors.NewDatabase()
	if err != nil {
		log.Printf("Warning: Failed to connect to database: %v", err)
		db = nil // Ensure db is nil if connection failed
	} else {
		db = database.GetDB()
	}

	// Initialize email sender (optional)
	var emailSender *connectors.EmailSender
	emailSender, err = connectors.NewEmailSender()
	if err != nil {
		log.Printf("Warning: Failed to initialize email sender: %v", err)
		emailSender = nil
	}

	// Create application context
	appCtx := &AppContext{
		DB:          db,
		EmailSender: emailSender,
	}

	// Setup router with context
	router := setupRouter(appCtx)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
