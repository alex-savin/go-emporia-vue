package config

import (
	"testing"
)

func TestFromBytesYAML(t *testing.T) {
	yamlData := []byte(`
emporia:
  credentials:
    username: test@example.com
    password: secret
  auto_reconnect: true
timezone: America/Chicago
logging:
  level: DEBUG
  output: JSON
`)

	cfg, err := FromBytes(yamlData)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	if cfg.Emporia.Credentials.Username != "test@example.com" {
		t.Errorf("expected username test@example.com, got %s", cfg.Emporia.Credentials.Username)
	}
	if cfg.Timezone != "America/Chicago" {
		t.Errorf("expected timezone America/Chicago, got %s", cfg.Timezone)
	}
	if cfg.Logger == nil {
		t.Error("expected logger to be set")
	}
}

func TestFromBytesJSON(t *testing.T) {
	jsonData := []byte(`{
		"emporia": {
			"credentials": {
				"username": "json@example.com",
				"password": "secret"
			}
		},
		"timezone": "UTC"
	}`)

	cfg, err := FromBytes(jsonData)
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	if cfg.Emporia.Credentials.Username != "json@example.com" {
		t.Errorf("expected username json@example.com, got %s", cfg.Emporia.Credentials.Username)
	}
	if cfg.Timezone != "UTC" {
		t.Errorf("expected timezone UTC, got %s", cfg.Timezone)
	}
}

func TestFromBytesEmpty(t *testing.T) {
	cfg, err := FromBytes([]byte{})
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Should return default config
	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected default timezone America/New_York, got %s", cfg.Timezone)
	}
	if cfg.Logger == nil {
		t.Error("expected default logger")
	}
}

func TestLoggingToLogger(t *testing.T) {
	tests := []struct {
		name    string
		logging *Logging
	}{
		{"nil logging", nil},
		{"JSON output", &Logging{Level: "INFO", Output: LoggingOutputJson}},
		{"TEXT output", &Logging{Level: "DEBUG", Output: LoggingOutputText}},
		{"invalid level", &Logging{Level: "INVALID", Output: LoggingOutputText}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := tt.logging.ToLogger()
			if logger == nil {
				t.Error("expected logger, got nil")
			}
		})
	}
}

func TestEnsureDefaults(t *testing.T) {
	cfg := &Config{}
	cfg.ensureDefaults()

	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected default timezone, got %s", cfg.Timezone)
	}
	if len(cfg.Emporia.ScopeOfInterest) != 3 {
		t.Errorf("expected 3 scopes, got %d", len(cfg.Emporia.ScopeOfInterest))
	}
	if cfg.Emporia.Cognito.ClientID != DefaultCognitoClientID {
		t.Errorf("expected default client ID, got %s", cfg.Emporia.Cognito.ClientID)
	}
	if cfg.Logger == nil {
		t.Error("expected default logger")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := defaultConfig()

	if cfg.Emporia.AutoReconnect != true {
		t.Error("expected AutoReconnect to be true")
	}
	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected America/New_York, got %s", cfg.Timezone)
	}
	if cfg.Emporia.Cognito.Region != DefaultCognitoRegion {
		t.Errorf("expected %s, got %s", DefaultCognitoRegion, cfg.Emporia.Cognito.Region)
	}
}
