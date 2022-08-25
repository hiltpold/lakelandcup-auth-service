package conf

import (
	"fmt"
	"os"

	logger "github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Api
type ApiConfiguration struct {
}

// PostgresConfiguration holds all the database related configuration.
type PostgresConfiguration struct {
}

// Configuration holds the api configuration
type Configuration struct {
	Host         string `mapstructure:"HOST"`
	Port         string `mapstructure:"PORT"`
	JWTSecretKey string `mapstructure:"JWT_SECRET"`
	URI          string `mapstructure:"POSTGRES_URI"`
}

// Load the environment set with the environment file
func loadEnvironment(filename string) error {
	var err error
	if filename != "" {
		err = godotenv.Load(filename)
	} else {
		err = godotenv.Load()
		// handle if .env file does not exist, this is OK
		if os.IsNotExist(err) {
			return nil
		}
	}
	return err
}

// LoadGlobal loads configuration from file and environment variables.
func LoadConfig(filename string) (config *Configuration, err error) {
	if err := loadEnvironment(filename); err != nil {
		return nil, err
	}
	viper.AddConfigPath(".")
	viper.SetConfigName(".dev")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	t := new(Configuration)

	err = viper.Unmarshal(t)
	err = viper.Unmarshal(&config)

	logger.Info(os.Getenv("PORT"))
	fmt.Print(config)
	logger.Info(config.Host)
	logger.Info(config.JWTSecretKey)
	logger.Info(config.Port)
	return config, nil
}
