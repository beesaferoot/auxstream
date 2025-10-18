package db

import (
	"auxstream/config"
	"context"
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes a GORM database connection with pgx driver
func InitDB(config config.Config, ctx context.Context) *gorm.DB {
	// Configure GORM logger
	gormLogger := logger.Default.LogMode(logger.Info)
	if config.GinMode == "release" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	// Open database connection using GORM with pgx driver
	db, err := gorm.Open(postgres.Open(config.DBUrl), &gorm.Config{
		Logger: gormLogger,
	})

	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Unable to get underlying sql.DB: %v\n", err)
		os.Exit(1)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	
	return db
}

// CloseDB closes the database connection
func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Printf("Unable to get underlying sql.DB: %v\n", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Printf("Error closing database: %v\n", err)
	}
}
