package config

import (
	"log/slog"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	ENV                  string         `env:"ENV" env-default:"development"`
	HTTPPort             string         `env:"HTTP_PORT" env-default:"8080"`
	CORSAllowedOrigins   []string       `env:"CORS_ALLOWED_ORIGINS" env-separator:","`
	CosmosDB             CosmosDBConfig `env-prefix:"COSMOS_"`
	ServiceBusConnection string         `env:"SERVICE_BUS_CONNECTION" env-required:"true"`
	JWTSecret            string         `env:"JWT_SECRET" env-required:"true"`
}

type CosmosDBConfig struct {
	Endpoint string `env:"ENDPOINT" env-required:"true"`
	Key      string `env:"KEY" env-required:"true"`
	Database string `env:"DATABASE" env-default:"complaintportal"`
}

func MustLoad() *Config {
	var cfg Config
	log := slog.Default()

	// Try to load from .env (for local dev) - only if file exists
	if _, err := os.Stat(".env"); err == nil {
		log.Info("attempting to load config from .env file")
		if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
			log.Warn("failed to read .env file, falling back to environment variables",
				slog.String("error", err.Error()))
			// Continue to environment variable fallback
		} else {
			log.Info("config loaded from .env file")
			return &cfg
		}
	} else {
		log.Info(".env file not found, loading from environment variables")
	}

	// Fallback to environment variables
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Error("cannot read config from environment variables",
			slog.String("error", err.Error()))
		panic(err)
	}

	log.Info("config loaded from environment variables")
	return &cfg
}
