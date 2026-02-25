package app

import (
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
)

type Config struct {
	RealWorldAppDb db_types.RealWorldAppDbConfig `json:"realworld_app_db" validate:"required"`
}
