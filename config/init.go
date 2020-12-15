package config

import (
	"log"
	"os"
)

//AppConfig
type AppConfig struct {
	AppPort string
	AppMode string
}

//New
func New() AppConfig {
	return AppConfig{
		AppPort: getEnvOrFatal("APP_PORT"),
		AppMode: getEnvOrFatal("APP_MODE"),
	}
}

func getEnvOrFatal(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Required ENV variable '%s' not found", key)
	}
	return value
}

func getEnvOrDefault(key string, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}
