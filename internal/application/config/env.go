package config

import (
	"os"

	"github.com/joho/godotenv"
)

type ConfigEnv struct {
	AppPort    string
	DbPort     string
	DbHost     string
	DbDatabase string
	DbUsername string
	DbPassword string
}

func LoadEnv() *ConfigEnv {
	godotenv.Load()
	return &ConfigEnv{
		AppPort:    getenv("APP_PORT", "8000"),
		DbPort:     getenv("DB_PORT", "5432"),
		DbDatabase: getenv("DB_DATABASE", "gin"),
		DbUsername: getenv("DB_USERNAME", "admin"),
		DbPassword: getenv("DB_PASSWORD", "adminpw"),
		DbHost:     getenv("DB_HOST", "localhost"),
	}
}

func getenv(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
