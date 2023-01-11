package storage

import (
	"fmt"

	"github.com/hiltpold/lakelandcup-auth-service/conf"
	"github.com/hiltpold/lakelandcup-auth-service/models"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repository struct {
	DB *gorm.DB
}

func openDb(c *conf.PostgresConfiguration, gormConfig *gorm.Config, database string) *gorm.DB {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?search_path=%s", c.User, c.Password, c.Host, c.Port, database, c.AppDatabaseSchema)
	db, err := gorm.Open(postgres.Open(connectionString), gormConfig)
	if err != nil {
		logrus.Fatal(fmt.Sprintf("Unable to connect to database '%s' given url: ", database), err)
	}
	logrus.Info(fmt.Sprintf("Connected to database %s", database))
	return db
}

func Dial(c *conf.PostgresConfiguration) Repository {
	gormDefaultConfig := &gorm.Config{}
	gormAppConfig := &gorm.Config{}

	// connect to default database
	logrus.Info(fmt.Sprintf("Use default database %s for initialization", c.DefaultDatabase))
	defaultDb := openDb(c, gormDefaultConfig, c.DefaultDatabase)

	// check if database exists and create it if necessary
	var dbexists bool
	dbSQL := fmt.Sprintf("SELECT EXISTS (SELECT FROM pg_database WHERE datname = '%s') AS dbexists;", c.AppDatabase)
	defaultDb.Raw(dbSQL).Row().Scan(&dbexists)
	if !dbexists {
		logrus.Info(fmt.Sprintf("Created database %s", c.AppDatabase))
		db := defaultDb.Exec(fmt.Sprintf("CREATE DATABASE %s;", c.AppDatabase))
		if db.Error != nil {
			logrus.Fatal("Unable to create app database: ", db.Error)
		}
	}

	// connect to app databse
	appDb := openDb(c, gormAppConfig, c.AppDatabase)

	// check if schema exists and create it if necessary
	var schemaexists bool
	schemaSQL := fmt.Sprintf("SELECT EXISTS(SELECT FROM pg_namespace WHERE nspname = '%s') AS schemaexisits;", c.AppDatabaseSchema)
	appDb.Raw(schemaSQL).Row().Scan(&schemaexists)
	// create app specfic schema, if not already existing
	if !schemaexists {
		// create service specific schema
		db := appDb.Exec(fmt.Sprintf("CREATE SCHEMA %s;", c.AppDatabaseSchema))
		if db.Error != nil {
			logrus.Fatal(fmt.Sprintf("Unable to create database schema %s", c.AppDatabaseSchema), db.Error)
		}
		logrus.Info(fmt.Sprintf("Created database schema %s", c.AppDatabaseSchema))
	}

	// migrate table
	appDb.AutoMigrate(&models.User{})

	return Repository{appDb}
}
