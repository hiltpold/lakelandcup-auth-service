package storage

import (
	"log"

	"github.com/hiltpold/lakelandcup-auth-service/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connection struct {
	DB *gorm.DB
}

func Dial(url string) Connection {
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		log.Fatalln(err)
	}

	db.AutoMigrate(&models.User{})

	return Connection{db}
}
