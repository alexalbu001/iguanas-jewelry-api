package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	Database     DatabaseConfig
	ImageStorage ImageStorage
	Redis        RedisConfig
	Stripe       StripeConfig
	Google       GoogleConfig
	SQS          SQSConfig
	Logging      LoggingConfig
	CORS         CORSConfig
	Sendgrid     SendgridConfig
	AppPort      int    `envconfig:"PORT" default:"8080"`
	AdminEmail   string `envconfig:"ADMIN_EMAIL" default:"alexalbu001@gmail.com"`
	Env          string `envconfig:"ENV" default:"dev"`
	Version      string `envconfig:"VERSION" default:"test-123"`
	JWTSecret    string `envconfig:"JWT_SECRET" required:"true"`
	WorkerMode   string `envconfig:"WORKER_MODE" default:"scheduler"`
}

type ImageStorage struct {
	Mode    string `envconfig:"IMAGE_STORAGE_MODE" default:"local"`
	BaseURL string `envconfig:"IMAGE_STORAGE_BASE_URL"`
	Bucket  string `envconfig:"IMAGE_STORAGE_BUCKET"`
	Region  string `envconfig:"IMAGE_STORAGE_REGION"`
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
	RedirectURL  string `envconfig:"REDIRECT_URL" default:"https://localhost:8080/auth/google/callback"`
	AdminOrigin  string `envconfig:"ADMIN_ORIGIN" default:"http://localhost:3001"`
}

type SQSConfig struct {
	QueueURL       string `envconfig:"QUEUE_URL"`
	AWSEndpointURL string `envconfig:"AWS_ENDPOINT_URL_SQS"`
}

type LoggingConfig struct {
	LogLevel  string `envconfig:"LOG_LEVEL" default:"info"`
	LogFormat string `envconfig:"LOG_FORMAT"`
}

type CORSConfig struct {
	AllowOrigins []string `envconfig:"CORS_ALLOWED_ORIGINS" default:"https://localhost:3000,http://localhost:3001"`
}

type SendgridConfig struct {
	SendgridApiKey string `envconfig:"SENDGRID_API_KEY"`
	FromEmail      string `envconfig:"FROM_EMAIL"`
	FromName       string `envconfig:"FROM_NAME"`
}

func Load() (*Config, error) {
	var cfg Config
	if err := envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Set environment-specific defaults
	if cfg.Env == "production" && cfg.AppPort == 8080 {
		cfg.AppPort = 443
	}

	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Env == "production" {
		if c.Stripe.StripeSK == "" {
			return fmt.Errorf("STRIPE_SECRET_KEY is required in production")
		}
		if c.Google.ClientID == "" {
			return fmt.Errorf("GOOGLE_CLIENT_ID is required in production")
		}
	}
	return nil
}
