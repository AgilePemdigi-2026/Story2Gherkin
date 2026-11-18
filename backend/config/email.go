package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type EmailConfig struct {
	Host         string `mapstructure:"SMTP_HOST"`
	Port         int    `mapstructure:"SMTP_PORT"`
	SenderName   string `mapstructure:"SMTP_SENDER_NAME"`
	AuthEmail    string `mapstructure:"SMTP_AUTH_EMAIL"`
	AuthPassword string `mapstructure:"SMTP_AUTH_PASSWORD"`
}

func NewEmailConfig() (*EmailConfig, error) {
	v := viper.New()
	v.SetConfigFile(".env")
	v.SetConfigType("env")
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError

		if !errors.Is(err, os.ErrNotExist) &&
			!errors.As(err, &notFound) {
			return nil, fmt.Errorf(
				"read email configuration: %w",
				err,
			)
		}
	}

	config := &EmailConfig{
		Host: strings.TrimSpace(
			v.GetString("SMTP_HOST"),
		),
		Port: v.GetInt("SMTP_PORT"),
		SenderName: strings.TrimSpace(
			v.GetString("SMTP_SENDER_NAME"),
		),
		AuthEmail: strings.TrimSpace(
			v.GetString("SMTP_AUTH_EMAIL"),
		),
		AuthPassword: v.GetString("SMTP_AUTH_PASSWORD"),
	}

	if config.Host == "" {
		return nil, errors.New("SMTP_HOST is required")
	}

	if config.Port < 1 || config.Port > 65535 {
		return nil, errors.New(
			"SMTP_PORT must be between 1 and 65535",
		)
	}

	if config.AuthEmail == "" {
		return nil, errors.New("SMTP_AUTH_EMAIL is required")
	}

	if strings.TrimSpace(config.AuthPassword) == "" {
		return nil, errors.New(
			"SMTP_AUTH_PASSWORD is required",
		)
	}

	return config, nil
}
