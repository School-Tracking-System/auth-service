package env

import (
	"log"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
	"go.uber.org/fx"
)

// Config holds all environment variables required by the auth service.
type Config struct {
	ServiceName string `env:"SERVICE_NAME" envDefault:"auth"`
	HTTPPort    string `env:"HTTP_PORT" envDefault:"8080"`
	GRPCPort    string `env:"GRPC_PORT" envDefault:"9090"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://postgres:postgres@localhost:5432/school_tracking?sslmode=disable"`
	RedisURL    string `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
	NatsURL     string `env:"NATS_URL" envDefault:"nats://localhost:4222"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"dev-secret-change-in-prod"`
	Environment string `env:"ENVIRONMENT" envDefault:"development"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"debug"`
}

// NewConfig loads the .env file and parses environment variables into a Config struct.
func NewConfig() *Config {
	// Load .env file if it exists (no error if missing).
	// System env vars take precedence over .env values.
	_ = godotenv.Load()

	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("failed to load environment variables: %v", err)
	}
	return &cfg
}

var Module = fx.Module("env", fx.Provide(NewConfig))
