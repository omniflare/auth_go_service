package db

import (
	"log"

	"github.com/omniflare/auth_go_service/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Init(config *config.Config) {
	uri := config.Database
	var err error
	DB, err = gorm.Open(postgres.Open(uri), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err.Error())
	}

	log.Printf("Connected to the database")

	if err := AutoMigrate(DB); err != nil {
		log.Fatalf("Unable to migrate the database: %v", err)
	}
}

// TODO : Optional -> we can also do a repository with all the db operations that way our db will not be a global variable
// not doing rn because that will be over engineering for this usecase.
