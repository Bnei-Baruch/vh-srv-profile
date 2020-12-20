package config

import (
	"log"
	"os"
)

//AppConfig
type AppConfig struct {
	AppPort           string
	AppMode           string
	DBNetwork         string
	DBHost            string
	DBUserName        string
	DBPassword        string
	DBName            string
	DBApplicationName string
}

//IsDev check if app in dev mode
func (config AppConfig) IsDev() bool {
	if config.AppMode == "dev" {
		return true
	}
	return false
}

//New
func New() AppConfig {
	return AppConfig{
		AppPort:           getEnvOrFatal("APP_PORT"),
		AppMode:           getEnvOrFatal("APP_MODE"),
		DBNetwork:         getEnvOrFatal("DB_NETWORK"),
		DBHost:            getEnvOrFatal("DB_HOST"),
		DBUserName:        getEnvOrFatal("DB_USERNAME"),
		DBPassword:        getEnvOrFatal("DB_PASSWORD"),
		DBName:            getEnvOrFatal("DB_NAME"),
		DBApplicationName: getEnvOrFatal("DB_APPNAME"),
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
