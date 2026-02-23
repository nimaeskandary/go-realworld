package app

import (
	auth_types "github.com/nimaeskandary/go-realworld/pkg/auth/types"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
	obs_types "github.com/nimaeskandary/go-realworld/pkg/observability/types"
)

type HttpServerConfig struct {
	Port             int      `json:"port" validate:"required"`
	IsSwaggerEnabled bool     `json:"is_swagger_enabled"`
	AllowedOrigins   []string `json:"allowed_origins"`
}

type Config struct {
	HttpServer     HttpServerConfig                `json:"http_server" validate:"required"`
	JwtAuthService auth_types.JwtAuthServiceConfig `json:"jwt_auth_service" validate:"required"`
	Slog           obs_types.SlogLoggerConfig      `json:"slog" validate:"required"`
	RealWorldAppDb db_types.RealWorldAppDbConfig   `json:"realworld_app_db" validate:"required"`
}
