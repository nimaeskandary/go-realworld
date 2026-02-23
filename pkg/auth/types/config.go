package auth_types

import (
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
)

type JwtAuthServiceConfig struct {
	SecretBase64         config_types.SecretString `json:"secret_base_64" validate:"required"`
	TokenDurationSeconds int64                     `json:"token_duration_seconds" validate:"required"`
}
