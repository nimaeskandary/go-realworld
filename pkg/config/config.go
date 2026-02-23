package config

import (
	"github.com/nimaeskandary/go-realworld/pkg/config/internal"
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	"github.com/nimaeskandary/go-realworld/pkg/util"

	"go.uber.org/fx"
)

func NewYamlConfigLoaderModule[T any](from []byte) fx.Option {
	return fx.Options(
		// inject the "from" param with a name tag to look it up later
		fx.Supply(fx.Annotate(from, fx.ResultTags(`name:"from"`))),
		util.NewFxModule[config_types.ConfigLoader[T]](
			"yaml_config_loader",
			func(params struct {
				fx.In
				config_types.SecretParser
				// use the tagged param injected earlier
				From []byte `name:"from"`
			}) (config_types.ConfigLoader[T], error) {
				return internal.NewYamlConfigLoader[T](params.SecretParser, params.From)
			},
		),
	)
}

func NewIdentitySecretParserModule() fx.Option {
	return util.NewFxModule[config_types.SecretParser](
		"identity_secret_parser",
		internal.NewIdentitySecretParser,
	)
}

func NewYamlConfigLoader[T any](secretParser config_types.SecretParser, from []byte) (config_types.ConfigLoader[T], error) {
	return internal.NewYamlConfigLoader[T](secretParser, from)
}

func NewIdentitySecretParser() config_types.SecretParser {
	return internal.NewIdentitySecretParser()
}
