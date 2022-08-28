package storage

import (
	"fmt"

	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/models"
	logger "github.com/hiltpold/lakelandcup-auth-service/utils"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Connection struct {
	DB *gorm.DB
}

func getUri(db *conf.PostgresConfiguration) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s", db.User, db.Password, db.Host, db.Port, db.Database)
}

func Dial(c *conf.PostgresConfiguration) Connection {
	url := getUri(c)
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})

	if err != nil {
		logger.Fatal("Failed to connect to Database", zap.Error(err))
	}

	db.AutoMigrate(&models.User{})

	return Connection{db}
}
