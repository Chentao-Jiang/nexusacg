package database

import (
	"log"

	"github.com/planforever/nexusacg/internal/model"
	"github.com/planforever/nexusacg/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Connect(cfg *config.Config) *gorm.DB {
	// Only log SQL queries in development; production uses Warn to avoid leaking data
	logLevel := logger.Warn
	if cfg.Env == "development" {
		logLevel = logger.Info
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	// Auto migrate tables
	db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.Order{},
		&model.OrderItem{},
		&model.Post{},
		&model.Comment{},
		&model.Like{},
		&model.Group{},
		&model.Event{},
		&model.ServiceProvider{},
		&model.RefreshToken{},
		&model.PaymentLog{},
		&model.ProfitShareRecord{},
		&model.EmailVerificationToken{},
	)

	log.Println("database connected and migrated")
	return db
}
