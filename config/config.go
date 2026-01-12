// Package config provides configuration management for the go-emporia-vue library.
package config

import (
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	// Version is the current library version.
	Version = "1.0.0"

	// LoggingOutputJson configures JSON log output.
	LoggingOutputJson = "JSON"

	// LoggingOutputText configures text log output.
	LoggingOutputText = "TEXT"

	// Default Cognito settings for Emporia API.
	DefaultCognitoClientID = "4qte47jbstod8apnfic0bunmrq"
	DefaultCognitoRegion   = "us-east-2"
	DefaultCognitoUserPool = "us-east-2_ghlOXVLi1"
)

// MetricsRecorder defines the interface for recording metrics.
type MetricsRecorder interface {
	// RecordRequest records an API request.
	RecordRequest(method, endpoint string, duration time.Duration, success bool)

	// RecordError records an error occurrence.
	RecordError(errorType string)

	// RecordRetry records a retry attempt.
	RecordRetry(endpoint string, attempt int)
}

// Config represents the runtime configuration with initialized logger.
type Config struct {
	Emporia  Emporia
	Timezone string
	Logger   *slog.Logger
	Metrics  MetricsRecorder
}

// config defines the structure of configuration data to be parsed from a config source.
type config struct {
	Emporia  Emporia  `json:"emporia" yaml:"emporia"`
	Timezone string   `json:"timezone" yaml:"timezone"`
	Logging  *Logging `json:"logging" yaml:"logging"`
}

// Emporia contains Emporia-specific configuration.
type Emporia struct {
	Credentials     Credentials `json:"credentials" yaml:"credentials"`
	AutoReconnect   bool        `json:"auto_reconnect" yaml:"auto_reconnect"`
	ScopeOfInterest []string    `json:"scope_of_interest" yaml:"scope_of_interest"`
	Cognito         Cognito     `json:"cognito" yaml:"cognito"`
}

// Credentials contains authentication credentials.
type Credentials struct {
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
}

// Cognito contains AWS Cognito configuration.
type Cognito struct {
	ClientID string `json:"client_id" yaml:"client_id"`
	Region   string `json:"region" yaml:"region"`
	UserPool string `json:"user_pool" yaml:"user_pool"`
}

// Logging contains logging configuration.
type Logging struct {
	Level  string `json:"level" yaml:"level"`
	Output string `json:"output" yaml:"output"`
	Source bool   `json:"source,omitempty" yaml:"source,omitempty"`
}

// ToLogger creates a slog.Logger from the logging configuration.
func (l *Logging) ToLogger() *slog.Logger {
	if l == nil {
		return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(l.Level)); err != nil {
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		AddSource: l.Source,
		Level:     level,
	}

	var handler slog.Handler
	switch l.Output {
	case LoggingOutputJson:
		handler = slog.NewJSONHandler(os.Stdout, opts)
	case LoggingOutputText:
		handler = slog.NewTextHandler(os.Stdout, opts)
	default:
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

// FromBytes unmarshals a byte slice of JSON or YAML config data into a Config.
func FromBytes(b []byte) (*Config, error) {
	if len(b) == 0 {
		return defaultConfig(), nil
	}

	c := new(config)
	if b[0] == '{' {
		if err := json.Unmarshal(b, c); err != nil {
			return nil, err
		}
	} else {
		if err := yaml.Unmarshal(b, c); err != nil {
			return nil, err
		}
	}

	return c.toConfig(), nil
}

// FromFile reads configuration from a file path.
func FromFile(path string) (*Config, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return FromBytes(b)
}

// Load searches for config.yaml in standard locations and loads it.
func Load() (*Config, error) {
	paths := []string{
		"config.yaml",
		"config.yml",
		"/etc/emporia/config.yaml",
		"/etc/emporia/config.yml",
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			return FromFile(path)
		}
	}

	// Return default config if no file found
	return defaultConfig(), nil
}

// toConfig converts the parsed config to the runtime Config with logger.
func (c *config) toConfig() *Config {
	cfg := &Config{
		Emporia:  c.Emporia,
		Timezone: c.Timezone,
		Logger:   c.Logging.ToLogger(),
	}

	cfg.ensureDefaults()
	return cfg
}

// ensureDefaults applies default values to missing configuration.
func (c *Config) ensureDefaults() {
	if c.Timezone == "" {
		c.Timezone = "America/New_York"
	}
	if len(c.Emporia.ScopeOfInterest) == 0 {
		c.Emporia.ScopeOfInterest = []string{"minute", "day", "month"}
	}
	if c.Emporia.Cognito.ClientID == "" {
		c.Emporia.Cognito.ClientID = DefaultCognitoClientID
	}
	if c.Emporia.Cognito.Region == "" {
		c.Emporia.Cognito.Region = DefaultCognitoRegion
	}
	if c.Emporia.Cognito.UserPool == "" {
		c.Emporia.Cognito.UserPool = DefaultCognitoUserPool
	}
	if c.Logger == nil {
		c.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}
}

// defaultConfig returns a Config with all default values.
func defaultConfig() *Config {
	cfg := &Config{
		Emporia: Emporia{
			AutoReconnect:   true,
			ScopeOfInterest: []string{"minute", "day", "month"},
			Cognito: Cognito{
				ClientID: DefaultCognitoClientID,
				Region:   DefaultCognitoRegion,
				UserPool: DefaultCognitoUserPool,
			},
		},
		Timezone: "America/New_York",
		Logger:   slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})),
	}
	return cfg
}
