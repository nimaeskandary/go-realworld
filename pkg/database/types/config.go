package db_types

import (
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
)

type SqlDbConfig struct {
	Username string                    `json:"username" validate:"required"`
	Password config_types.SecretString `json:"password" validate:"required"`
	Host     string                    `json:"host" validate:"required"`
	Port     int                       `json:"port" validate:"required"`
	DBName   string                    `json:"db_name" validate:"required"`
	SslMode  string                    `json:"ssl_mode" validate:"required,oneof=disable require verify-ca verify-full"`
}

type RealWorldAppDbConfig SqlDbConfig
