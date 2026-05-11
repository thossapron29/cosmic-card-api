package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppName     string
	AppVersion  string
	AppEnv      string
	Port        string
	DatabaseURL string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		AppName:     os.Getenv("APP_NAME"),
		AppVersion:  os.Getenv("APP_VERSION"),
		AppEnv:      os.Getenv("APP_ENV"),
		Port:        os.Getenv("PORT"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	required := map[string]string{
		"APP_NAME":     c.AppName,
		"APP_VERSION":  c.AppVersion,
		"APP_ENV":      c.AppEnv,
		"PORT":         c.Port,
		"DATABASE_URL": c.DatabaseURL,
	}

	for key, value := range required {
		if value == "" {
			return fmt.Errorf("missing required environment variable: %s", key)
		}
	}

	return nil
}
