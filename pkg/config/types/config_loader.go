package config_types

// ConfiLoader is a generic interface for loading configuration of type T, e.g.
// { JwtAuthServiceConfig: JwtAuthServiceConfig, ClientDatabaseConfig: DatabaseConfig, etc. }
type ConfigLoader[T any] interface {
	GetConfig() T
}

// SecretString is a type that represents a secret string that will need to be parsed using a SecretParser.
// The ConfigLoader implementation must use a SecretParser to resolve the actual secret value
// when it encounters a SecretString field in the config struct
type SecretString string

type SecretParser interface {
	Parse(raw string) (string, error)
}
