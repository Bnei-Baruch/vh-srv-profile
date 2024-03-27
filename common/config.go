package common

import (
	"os"
)

type config struct {
	Port string
	Mode string
	Env  string

	PgHost   string
	PgPort   string
	PgUser   string
	PgPass   string
	PgDbName string

	NatsUrl string

	KeycloakServerUrl    string
	KeycloakRealm        string
	KeycloakClientID     string
	KeycloakClientSecret string

	OrdersServiceUrl string
}

var Config = new(config)

func LoadConfig() {
	// defaults
	Config.Port = "7471"
	Config.Mode = "debug"
	Config.Env = "dev"
	Config.PgHost = "localhost"
	Config.PgPort = "5678"
	Config.PgUser = "postgres"
	Config.PgPass = "password"
	Config.PgDbName = "profiledb"

	// env override
	if val, ok := os.LookupEnv("APP_PORT"); ok {
		Config.Port = val
	}
	if val, ok := os.LookupEnv("APP_MODE"); ok {
		Config.Mode = val
	}
	if val, ok := os.LookupEnv("APP_ENV"); ok {
		Config.Env = val
	}
	if val, ok := os.LookupEnv("DB_HOST"); ok {
		Config.PgHost = val
	}
	if val, ok := os.LookupEnv("DB_PORT"); ok {
		Config.PgPort = val
	}
	if val, ok := os.LookupEnv("DB_USER"); ok {
		Config.PgUser = val
	}
	if val, ok := os.LookupEnv("DB_PASSWORD"); ok {
		Config.PgPass = val
	}
	if val, ok := os.LookupEnv("DB_NAME"); ok {
		Config.PgDbName = val
	}
	if val, ok := os.LookupEnv("NATS_URL"); ok {
		Config.NatsUrl = val
	}
	if val, ok := os.LookupEnv("KEYCLOAK_SERVER_URL"); ok {
		Config.KeycloakServerUrl = val
	}
	if val, ok := os.LookupEnv("KEYCLOAK_REALM"); ok {
		Config.KeycloakRealm = val
	}
	if val, ok := os.LookupEnv("KEYCLOAK_CLIENT_ID"); ok {
		Config.KeycloakClientID = val
	}
	if val, ok := os.LookupEnv("KEYCLOAK_CLIENT_SECRET"); ok {
		Config.KeycloakClientSecret = val
	}
	if val, ok := os.LookupEnv("ORDERS_SERVICE_URL"); ok {
		Config.OrdersServiceUrl = val
	}
}
