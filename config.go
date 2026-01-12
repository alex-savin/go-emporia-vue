package vue

import (
	"github.com/alex-savin/go-emporia-vue/config"
)

// Re-export types from config package for backward compatibility.
type (
	// Config represents the runtime configuration.
	Config = config.Config
	// Emporia contains Emporia-specific configuration.
	Emporia = config.Emporia
	// Credentials contains authentication credentials.
	Credentials = config.Credentials
	// Cognito contains AWS Cognito configuration.
	Cognito = config.Cognito
	// Logging contains logging configuration.
	Logging = config.Logging
	// MetricsRecorder defines the interface for recording metrics.
	MetricsRecorder = config.MetricsRecorder
)

// Re-export constants from config package.
const (
	Version           = config.Version
	LoggingOutputJson = config.LoggingOutputJson
	LoggingOutputText = config.LoggingOutputText
)

// Conf loads configuration from standard locations.
// Deprecated: Use config.Load() instead.
func Conf() (*Config, error) {
	return config.Load()
}

// FromBytes parses configuration from a byte slice.
func FromBytes(b []byte) (*Config, error) {
	return config.FromBytes(b)
}

// FromFile reads configuration from a file path.
func FromFile(path string) (*Config, error) {
	return config.FromFile(path)
}
