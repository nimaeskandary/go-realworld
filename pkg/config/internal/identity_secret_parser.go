package internal

import config_types "github.com/nimaeskandary/go-realworld/pkg/config/types"

// IdentitySecretParser An example implementation of SecretParser that returns the raw string as is,
// in real life, this may be something like a AwsSsmSecretParser
type identitySecretParser struct{}

func NewIdentitySecretParser() config_types.SecretParser {
	return &identitySecretParser{}
}

func (p *identitySecretParser) Parse(raw string) (string, error) {
	return raw, nil
}
