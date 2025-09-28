package config

import (
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Port        string `env:"PORT" envDefault:"8080"`
	JWTSecret   string `env:"JWT_SECRET" envDefault:"your-secret-key"`
	AWSRegion   string `env:"AWS_REGION" envDefault:"us-east-1"`
	TableName   string `env:"TABLE_NAME" envDefault:"notes"`
	EndpointURL string `env:"DYNAMODB_ENDPOINT" envDefault:""` // For local testing
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
