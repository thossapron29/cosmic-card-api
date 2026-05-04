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
	_ = godotenv.Load()

	cfg := Config{
		AppName:    mustGetEnv("APP_NAME"),
		AppVersion: mustGetEnv("APP_VERSION"),
		AppEnv:     mustGetEnv("APP_ENV"),
		Port:       mustGetEnv("PORT"),
	}

	return cfg
}

func mustGetEnv(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Missing required environment variable: %s", key)
	}
	return value
}
