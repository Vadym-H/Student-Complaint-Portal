package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ENV                  string         `env:"ENV" env-default:"development"`
	HTTPPort             string         `env:"HTTP_PORT" env-default:"8080"`
	CosmosDB             CosmosDBConfig `env-prefix:"COSMOS"`
	ServiceBusConnection string         `env:"SERVICE_BUS_CONNECTION" env-required:"true"`
	JWTSecret            string         `env:"JWT_SECRET" env-required:"true"`
}

type CosmosDBConfig struct {
	Endpoint string `env:"_ENDPOINT" env-required:"true"`
	Key      string `env:"_KEY" env-required:"true"`
	Database string `env:"_DATABASE" env-default:"complaintportal"`
}

func MustLoad() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		log.Fatalf("cannot read config: %v", err)
	}

	return &cfg
}
