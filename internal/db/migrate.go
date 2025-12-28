package db

import (
	"github.com/omniflare/auth_go_service/internal/models"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {

	// TODO: add model here ; to automigrate
	return db.AutoMigrate(&models.User{})
}
