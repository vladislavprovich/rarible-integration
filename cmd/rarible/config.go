package main

import (
	"context"
	"fmt"
	"log"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/vladislavprovich/rarible-integration/internal/handler"
	"github.com/vladislavprovich/rarible-integration/pkg/client/rarible"
	"github.com/vladislavprovich/rarible-integration/pkg/logger"
)

type Config struct {
	Client rarible.Config
	Server handler.Config
	Logger *logger.Config
}

func LoadConfig(ctx context.Context) (*Config, error) {
	var cfg Config

	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: .env file not found or failed to load: %v\n", err)
	}

	if err = envconfig.Process("", &cfg); err != nil {
		return nil, fmt.Errorf("failed to load root config: %w", err)
	}

	if err = cfg.ValidateWithContext(ctx); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.Server),
		validation.Field(&c.Client),
		validation.Field(&c.Logger),
	)
}
