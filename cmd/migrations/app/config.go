package app

import (
	"log"
	"os"

	"github.com/nimaeskandary/go-realworld/pkg/config"
	db_types "github.com/nimaeskandary/go-realworld/pkg/database/types"
)

type Config struct {
	RealWorldAppDb db_types.RealWorldAppDbConfig `json:"realworld_app_db" validate:"required"`
}

func NewConfig(path string) Config {
	configData, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("failed to read config file %v: %v", path, err)
	}

	secretParser := config.NewIdentitySecretParser()
	configLoader, err := config.NewYamlConfigLoader[Config](secretParser, configData)
	if err != nil {
		log.Fatalf("failed to create config loader: %v", err)
	}

	cfg := configLoader.GetConfig()

	return cfg
}
