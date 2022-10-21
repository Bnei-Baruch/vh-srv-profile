package main

import (
	"log"
	"os"
)

type appConfig struct {
	appPort string
	appMode string
}

func loadConfig() appConfig {
	return appConfig{
		appPort: getEnvOrFatal("APP_PORT"),
		appMode: getEnvOrFatal("APP_MODE"),
	}
}

func getServerUrl() string {
	var serverUrl string

	if getEnvOrFatal("APP_MODE") == "prod" {
		serverUrl = "https://api.kli.one"
	} else {
		serverUrl = "https://api.eurokab.info"
	}

	return serverUrl
}

func getEnvOrFatal(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		log.Println("Required ENV variable %q not found", key)
	}
	return value
}
