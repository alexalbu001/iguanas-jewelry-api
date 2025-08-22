package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database   DatabaseConfig
	Redis      RedisConfig
	Stripe     StripeConfig
	Google     GoogleConfig
	SQS        SQSConfig
	Logging    LoggingConfig
	AppPort    string `envconfig:"PORT" default:":8080"`
	AdminEmail string `envconfig:"ADMIN_EMAIL" default:"alexalbu001@gmail.com"`
	Env        string `envconfig:"ENV" default:"dev"`
	Version    string `envconfig:"VERSION" default:"test-123"`
}

type DatabaseConfig struct {
	DatabaseURL string `envconfig:"DATABASE_URL" required:"true"`
}

type RedisConfig struct {
	RedisURL string `envconfig:"REDIS_URL" required:"true"`
}

type StripeConfig struct {
	StripeSK            string `envconfig:"STRIPE_SECRET_KEY"`
	StripeWebhookSecret string `envconfig:"STRIPE_WEBHOOK_SECRET"`
}

type GoogleConfig struct {
	ClientID     string `envconfig:"GOOGLE_CLIENT_ID"`
	ClientSecret string `envconfig:"GOOGLE_CLIENT_SECRET"`
	RedirectURL  string `envconfig:"REDIRECT_URL" default:"http://localhost:8080/auth/google/callback"`
}

type SQSConfig struct {
	QueueURL       string `envconfig:"QUEUE_URL"`
	AWSEndpointURL string `envconfig:"AWS_ENDPOINT_URL_SQS"`
}

type LoggingConfig struct {
	LogLevel  string `envconfig:"LOG_LEVEL" default:"info"`
	LogFormat string `envconfig:"LOG_FORMAT"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	return &cfg, nil
}
