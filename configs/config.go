package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Port        string
		Environment string
	}

	Database struct {
		URI  string
		Name string
	}

	JWT struct {
		Secret string
	}
}

var (
	config Config
)

func LoadConfig() (*Config, error) {
	var config Config
	err := godotenv.Load(".env")
	if err != nil {
		log.Printf("Error loading .env file, relying on system environment variables: %s", err)
	}

	// Assign environment variables to the config struct
	config.App.Port = os.Getenv("PORT")
	config.Database.URI = os.Getenv("MONGO_URI")
	config.Database.Name = os.Getenv("DB_NAME")
	config.JWT.Secret = os.Getenv("JWT_SECRET")

	if config.App.Port == "" || config.Database.URI == "" || config.Database.Name == "" || config.JWT.Secret == "" {
		return &config, fmt.Errorf("missing required environment variables")
	}

	log.Printf("Config loaded: %+v", config)
	return &config, nil
}

func GetConfig() *Config {
	return &config
}
