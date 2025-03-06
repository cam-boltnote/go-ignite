package connectors

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/cam-boltnote/go-ignite/internal/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	db      *gorm.DB
	enabled bool
}

// NewDatabase creates a new database connection and handles auto-migration
func NewDatabase() (*Database, error) {
	// Check if required environment variables are set
	requiredVars := []string{"DB_USER", "DB_PASSWORD", "DB_HOST", "DB_PORT", "DB_NAME"}
	for _, v := range requiredVars {
		if os.Getenv(v) == "" {
			log.Printf("Database configuration missing: %s. Database functionality will be disabled.", v)
			return &Database{enabled: false}, nil
		}
	}

	// Create DSN from environment variables
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)

	// Configure GORM logger
	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// Open connection to database
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to database: %v. Database functionality will be disabled.", err)
		return &Database{enabled: false}, nil
	}

	// Get underlying SQL DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Failed to get underlying SQL DB: %v. Database functionality will be disabled.", err)
		return &Database{enabled: false}, nil
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Create database wrapper
	database := &Database{
		db:      db,
		enabled: true,
	}

	// Run auto-migrations for default models
	if err := database.AutoMigrateDefaults(); err != nil {
		log.Printf("Failed to run auto-migrations: %v. Database functionality will be disabled.", err)
		return &Database{enabled: false}, nil
	}

	return database, nil
}

// AutoMigrateDefaults runs auto-migration for default models
func (db *Database) AutoMigrateDefaults() error {
	if !db.enabled {
		log.Println("Database functionality is disabled. Skipping migrations.")
		return nil
	}
	return db.db.AutoMigrate(
		&models.User{},
		&models.Settings{},
	)
}

// AutoMigrate performs database migrations for arbitrary models
func (db *Database) AutoMigrate(models ...interface{}) error {
	if !db.enabled {
		log.Println("Database functionality is disabled. Skipping migrations.")
		return nil
	}
	for _, model := range models {
		if err := db.db.AutoMigrate(model); err != nil {
			return fmt.Errorf("failed to migrate model %T: %v", model, err)
		}
		log.Printf("Successfully migrated model %T", model)
	}
	return nil
}

// Transaction executes a function within a database transaction
func (db *Database) Transaction(fc func(tx *gorm.DB) error) error {
	if !db.enabled {
		log.Println("Database functionality is disabled. Skipping transaction.")
		return nil
	}
	return db.db.Transaction(fc)
}

// GetDB returns the underlying GORM DB instance
func (db *Database) GetDB() *gorm.DB {
	if !db.enabled {
		log.Println("Database functionality is disabled. Returning nil DB.")
		return nil
	}
	return db.db
}

// Close closes the database connection
func (db *Database) Close() error {
	if !db.enabled {
		log.Println("Database functionality is disabled. No connection to close.")
		return nil
	}
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health performs a health check on the database
func (db *Database) Health() error {
	if !db.enabled {
		log.Println("Database functionality is disabled. Health check skipped.")
		return nil
	}
	sqlDB, err := db.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Ping()
}
