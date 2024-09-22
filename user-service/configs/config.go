package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	App struct {
		Port string
	}

	Database struct {
		URI  string
		Name string
	}

	JWT struct {
		Secret string
	}
}

func LoadConfig() (Config, error) {
	var config Config

	// Load environment variables from the .env file
	err := godotenv.Load("/Users/mbanhawy/dev/work/luka-platform/user-service/.env")

	if err != nil {
		log.Printf("Error loading .env file, relying on system environment variables: %s", err)
	}

	// Assign environment variables to the config struct
	config.App.Port = os.Getenv("PORT")
	config.Database.URI = os.Getenv("MONGO_URI")
	config.Database.Name = os.Getenv("DB_NAME")
	config.JWT.Secret = os.Getenv("JWT_SECRET")

	// Validate if the required variables are set
	if config.App.Port == "" || config.Database.URI == "" || config.Database.Name == "" || config.JWT.Secret == "" {
		return config, err
	}

	log.Printf("Config loaded: %+v", config)
	return config, nil
}
