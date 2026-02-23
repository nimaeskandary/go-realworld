package obs_types

type SlogLoggerConfig struct {
	Level string `json:"level" validate:"required,oneof=DEBUG INFO WARN ERROR"`
}
