package logger

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

type Config struct {
	Level       string `yaml:"level" env:"LOGGER_LEVEL" default:"info"`
	Format      string `yaml:"format" env:"LOGGER_FORMAT" default:"json"`
	ServiceName string `yaml:"service_name" env:"SERVICE_NAME" default:"app"`
	WithSource  bool   `yaml:"with_source" env:"LOGGER_WITH_SOURCE" default:"false"`
}

func (c Config) Validate(_ context.Context) error {
	levelMap := getLevelMap()

	if _, ok := levelMap[strings.ToLower(c.Level)]; !ok {
		return fmt.Errorf("invalid log level: %s", c.Level)
	}

	switch strings.ToLower(c.Format) {
	case "json", "text":
	default:
		return errors.New("invalid logger format")
	}

	return nil
}
