package config

import (
	"log"
	"os"

	// spell-checker: disable
	"github.com/joho/godotenv"
	// spell-checker: enable
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
		Expire int
	}
}

func LoadConfig() (Config, error) {
	var config Config
	// spell-checker: disable-next-line
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file, relying on system environment variables: %s", err)
	}

	// Assign environment variables to the config struct
	config.App.Port = os.Getenv("PORT")
	config.Database.URI = os.Getenv("MONGO_URI")
	config.Database.Name = os.Getenv("DB_NAME")
	config.JWT.Secret = os.Getenv("JWT_SECRET")
	config.App.Port = os.Getenv("PORT")
	config.App.Environment = os.Getenv("ENVIRONMENT")

	// Validate if the required variables are set
	if config.App.Port == "" || config.Database.URI == "" || config.Database.Name == "" || config.JWT.Secret == "" {
		return config, err
	}

	log.Printf("Config loaded: %+v", config)
	return config, nil
}
