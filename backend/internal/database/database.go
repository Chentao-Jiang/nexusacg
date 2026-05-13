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
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
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
	)

	log.Println("database connected and migrated")
	return db
}
