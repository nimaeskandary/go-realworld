package internal

import (
	"context"
	"fmt"

	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"github.com/goccy/go-yaml"
)

type secretParserContextKey struct{}

type YamlConfigLoader[T any] struct {
	parsed *T
}

func NewYamlConfigLoader[T any](secretParser config_types.SecretParser, from []byte) (config_types.ConfigLoader[T], error) {
	parsed, err := parseYaml[T](secretParser, from)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	return &YamlConfigLoader[T]{
		parsed: parsed,
	}, nil
}

func (c *YamlConfigLoader[T]) GetConfig() T {
	return *c.parsed
}

// parseYaml parses the YAML bytes into the target config struct T,
// resolving any SecretString type fields using the provided SecretParser,
// and validates the struct using the "validate" struct tags. , e.g.
//
//	type SomeServiceConfig struct {
//			MyField string `json:"my_field" validate:"required"`
//			MySecretField configtypes.SecretString `json:"my_secret_field"`
//	},
//
// see https://github.com/go-playground/validator for a full list of supported validations
func parseYaml[T any](secretParser config_types.SecretParser, from []byte) (*T, error) {
	cfg := new(T)

	// add secret parser to context
	ctx := context.WithValue(context.Background(), secretParserContextKey{}, secretParser)

	unmarshalOption := yaml.CustomUnmarshalerContext(
		// when unmarshaling a SecretString, use the SecretParser from context to resolve it
		func(ctx context.Context, s *config_types.SecretString, data []byte) error {
			p, ok := ctx.Value(secretParserContextKey{}).(config_types.SecretParser)
			if !ok {
				return fmt.Errorf("secret parser not found in context")
			}

			// decode the raw YAML bytes into a temporary string
			var raw string
			if err := yaml.Unmarshal(data, &raw); err != nil {
				return err
			}

			// resolve via the parser
			resolved, err := p.Parse(raw)
			if err != nil {
				return fmt.Errorf("failed to parse secret: %w", err)
			}

			*s = config_types.SecretString(resolved)
			return nil
		},
	)

	err := yaml.UnmarshalContext(ctx, from, cfg, unmarshalOption, yaml.Validator(util.NewValidator()))
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
