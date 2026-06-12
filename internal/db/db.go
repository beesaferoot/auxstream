package db

import (
	"auxstream/config"
	"log"

	"github.com/go-playground/validator/v10"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var validate = validator.New()

// InitDB initializes a GORM database connection with PostgreSQL.
func InitDB(conf config.Config) *gorm.DB {
	gormLogger := logger.Default.LogMode(logger.Info)
	if conf.GinMode == "release" {
		gormLogger = logger.Default.LogMode(logger.Error)
	}

	db, err := gorm.Open(postgres.Open(conf.DBUrl), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Unable to get underlying sql.DB: %v", err)
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db
}

// CloseDB closes the underlying connection pool, logging any error rather than returning it.
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
