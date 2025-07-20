package handler

import (
	"context"
	"time"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Port              string        `envconfig:"HTTP_SERVER_PORT"`
	Timeout           time.Duration `envconfig:"HTTP_SERVER_TIMEOUT_ROUTER"`
	WriteTimeout      time.Duration `envconfig:"HTTP_SERVER_WRITE_TIMEOUT"`
	IdleTimeout       time.Duration `envconfig:"HTTP_SERVER_IDLE_TIMEOUT"`
	MaxAge            int64         `envconfig:"HTTP_SERVER_MAX_AGE"`
	AllowCredentials  bool          `envconfig:"HTTP_SERVER_ALLOW_CREDENTIALS"`
	HTTPClientTimeout time.Duration `envconfig:"HTTP_SERVER_CLIENT_TIMEOUT"`
	APIVersion        string        `envconfig:"HTTP_SERVER_API_VERSION"`
}

func (c Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.Timeout, validation.Required, validation.Min(1)),
		validation.Field(&c.WriteTimeout, validation.Required, validation.Min(1)),
		validation.Field(&c.IdleTimeout, validation.Required, validation.Min(1)),
		validation.Field(&c.MaxAge, validation.Required),
		validation.Field(&c.AllowCredentials, validation.Required),
		validation.Field(&c.HTTPClientTimeout, validation.Required),
	)
}
