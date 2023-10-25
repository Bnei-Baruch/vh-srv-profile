package common

import (
	"log"
	"os"
)

type AppConfig struct {
	Port string
	Mode string
}

func LoadConfig() AppConfig {
	return AppConfig{
		Port: getEnvOrFatal("APP_PORT"),
		Mode: getEnvOrFatal("APP_MODE"),
	}
}

func GetOrdersServiceUrl() string {
	serverUrl := os.Getenv("ORDERS_SERVICE_URL")
	if serverUrl != "" {
		return serverUrl
	}

	if getEnvOrFatal("APP_MODE") == "prod" {
		serverUrl = "https://api.kli.one/pay"
	} else {
		serverUrl = "https://api.eurokab.info/pay"
	}

	return serverUrl
}

func getEnvOrFatal(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Printf("Required ENV variable %q not found\n", key)
	}
	return value
}
