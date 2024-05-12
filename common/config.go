package common

import (
	"github.com/kelseyhightower/envconfig"
)

type envConfig struct {
	Port string `envconfig:"APP_PORT" default:"7471"`
	Mode string `envconfig:"APP_MODE" default:"debug"`
	Env  string `envconfig:"APP_ENV" default:"dev"`

	PgHost   string `envconfig:"DB_HOST" default:"localhost"`
	PgPort   string `envconfig:"DB_PORT" default:"5678"`
	PgUser   string `envconfig:"DB_USER" default:"postgres"`
	PgPass   string `envconfig:"DB_PASSWORD" default:"password"`
	PgDbName string `envconfig:"DB_NAME" default:"profiledb"`

	NatsUrl string `envconfig:"NATS_URL"`

	KeycloakServerUrl    string `envconfig:"KEYCLOAK_SERVER_URL"`
	KeycloakRealm        string `envconfig:"KEYCLOAK_REALM"`
	KeycloakClientID     string `envconfig:"KEYCLOAK_CLIENT_ID"`
	KeycloakClientSecret string `envconfig:"KEYCLOAK_CLIENT_SECRET"`

	OrdersServiceUrl string `envconfig:"ORDERS_SERVICE_URL"`
}

var Config = new(envConfig)

func LoadConfig() {
	envconfig.Process("LIST", Config)
}
