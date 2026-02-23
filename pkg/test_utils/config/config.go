package config

import (
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

type Config struct {
	Slog           obs_types.SlogLoggerConfig
	JwtAuthService auth_types.JwtAuthServiceConfig
}

func NewTestConfig() Config {
	return Config{
		Slog: obs_types.SlogLoggerConfig{
			Level: "DEBUG",
		},
		JwtAuthService: auth_types.JwtAuthServiceConfig{
			// generated via "openssl rand -base64 32"
			SecretBase64:         "1TjsQI3mv84OxUhS55owxZwLDXKMGj2PVUUQIr+E604=",
			TokenDurationSeconds: 3600,
		},
	}
}
