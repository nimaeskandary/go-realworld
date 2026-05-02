package internal_test

import (
	"testing"

	config_internal "github.com/nimaeskandary/go-realworld/pkg/config/internal"
	config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"
	config_types_mocks "github.com/nimaeskandary/go-realworld/pkg/config/types/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testConfig struct {
	Host   string                    `json:"host" validate:"required"`
	Port   int                       `json:"port" validate:"required,min=1,max=65535"`
	Secret config_types.SecretString `json:"secret"`
}

func Test_YamlConfigLoader(t *testing.T) {
	t.Parallel()

	t.Run("NewYamlConfigLoader", func(t *testing.T) {
		t.Parallel()

		t.Run("returns loader with valid config", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: localhost
port: 8080
secret: my-secret
`)

			loader, err := config_internal.NewYamlConfigLoader[testConfig](config_internal.NewIdentitySecretParser(), yamlBytes)
			require.NoError(t, err)
			require.NotNil(t, loader)

			cfg := loader.GetConfig()
			assert.Equal(t, "localhost", cfg.Host)
			assert.Equal(t, 8080, cfg.Port)
			assert.Equal(t, config_types.SecretString("my-secret"), cfg.Secret)
		})

		t.Run("resolves SecretString via parser", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: localhost
port: 8080
secret: encrypted-raw-value
`)

			mockParser := config_types_mocks.NewMockSecretParser(t)
			mockParser.EXPECT().Parse("encrypted-raw-value").Return("resolved-secret-value", nil)

			loader, err := config_internal.NewYamlConfigLoader[testConfig](mockParser, yamlBytes)
			require.NoError(t, err)

			cfg := loader.GetConfig()
			assert.Equal(t, config_types.SecretString("resolved-secret-value"), cfg.Secret)
		})

		t.Run("bubbles up validation errors", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: ""
port: 0
secret: my-secret
`)

			_, err := config_internal.NewYamlConfigLoader[testConfig](config_internal.NewIdentitySecretParser(), yamlBytes)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "required")
		})

		t.Run("bubbles up YAML parse errors", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: localhost
port: [invalid yaml
secret: my-secret
`)

			_, err := config_internal.NewYamlConfigLoader[testConfig](config_internal.NewIdentitySecretParser(), yamlBytes)

			assert.Error(t, err)
		})

		t.Run("bubbles up secret parser errors", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: localhost
port: 8080
secret: encrypted-raw-value
`)

			mockParser := config_types_mocks.NewMockSecretParser(t)
			mockParser.EXPECT().Parse("encrypted-raw-value").Return("", assert.AnError)

			_, err := config_internal.NewYamlConfigLoader[testConfig](mockParser, yamlBytes)

			assert.Error(t, err)
			assert.Contains(t, err.Error(), "failed to parse secret")
		})
	})

	t.Run("GetConfig", func(t *testing.T) {
		t.Parallel()

		t.Run("returns parsed config values", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: example.com
port: 443
secret: test-secret
`)

			loader, err := config_internal.NewYamlConfigLoader[testConfig](config_internal.NewIdentitySecretParser(), yamlBytes)
			require.NoError(t, err)

			cfg := loader.GetConfig()
			assert.Equal(t, "example.com", cfg.Host)
			assert.Equal(t, 443, cfg.Port)
			assert.Equal(t, config_types.SecretString("test-secret"), cfg.Secret)
		})

		t.Run("returns a copy of the parsed config", func(t *testing.T) {
			t.Parallel()

			yamlBytes := []byte(`
host: original.com
port: 8080
secret: original-secret
`)

			loader, err := config_internal.NewYamlConfigLoader[testConfig](config_internal.NewIdentitySecretParser(), yamlBytes)
			require.NoError(t, err)

			cfg := loader.GetConfig()
			cfg.Host = "modified.com"
			cfg.Port = 9999
			cfg.Secret = "modified-secret"

			cfg2 := loader.GetConfig()
			assert.Equal(t, "original.com", cfg2.Host)
			assert.Equal(t, 8080, cfg2.Port)
			assert.Equal(t, config_types.SecretString("original-secret"), cfg2.Secret)
		})
	})
}
