package conf

import (
	"os"
	"strings"

	"path/filepath"

	"github.com/hiltpold/lakelandcup-auth-service/utils"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Api
type ApiConfiguration struct {
	Host                  string `mapstructure:"HOST"`
	Port                  string `mapstructure:"PORT"`
	TokenSecretKey        string `mapstructure:"JWT_TOKEN_SECRET_KEY"`
	TokenExpires          int64  `mapstructure:"JWT_TOKEN_EXPIRES_H"`
	AccessTokenSecretKey  string `mapstructure:"JWT_ACCESS_TOKEN_SECRET_KEY"`
	AccessTokenExpires    int64  `mapstructure:"JWT_ACCESS_TOKEN_EXPIRES_H"`
	RefreshTokenSecretKey string `mapstructure:"JWT_REFRESH_TOKEN_SECRET_KEY"`
	RefreshTokenExpires   int64  `mapstructure:"JWT_REFRESH_TOKEN_EXPIRES_H"`
}

// PostgresConfiguration holds all the database related configuration.
type PostgresConfiguration struct {
	Host     string `mapstructure:"POSTGRES_HOST"`
	Port     string `mapstructure:"POSTGRES_PORT"`
	User     string `mapstructure:"POSTGRES_USER"`
	Password string `mapstructure:"POSTGRES_PASSWORD"`
	Database string `mapstructure:"POSTGRES_DB"`
}

// Sendgrid
type MailConfiguration struct {
	SGSecretKey string `mapstructure:"SENDGRID_KEY"`
}

// Configuration holds the api configuration
type Configuration struct {
	API  ApiConfiguration      `mapstructure:",squash"`
	DB   PostgresConfiguration `mapstructure:",squash"`
	Mail MailConfiguration     `mapstructure:",squash"`
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

	fp, fn := filepath.Split(filename)
	fn_splitted := strings.Split(fn, ".")
	configName := strings.Join(fn_splitted[0:len(fn_splitted)-1], ".")
	configType := fn_splitted[len(fn_splitted)-1]

	viper.AddConfigPath(utils.Ternary(fp != "", fp, "."))
	viper.SetConfigName(configName)
	viper.SetConfigType(configType)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()

	if err != nil {
		return nil, err
	}

	t := new(Configuration)

	err = viper.Unmarshal(t)
	err = viper.Unmarshal(&config)

	return config, nil
}
