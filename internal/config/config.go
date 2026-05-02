package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName    string
	AppVersion string
	AppEnv     string
	Port       string
}

func Load() Config {
	err := godotenv.Load()

	if err != nil {
		log.Println("Error loading .env file")
	}

	return Config{
		AppName:    getEnv("APP_NAME", "cosmic-card-api"),
		AppVersion: getEnv("APP_VERSION", "0.1.0"),
		AppEnv:     getEnv("APP_ENV", "local"),
		Port:       getEnv("PORT", "8080"),
	}
}

func getEnv(key string, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
