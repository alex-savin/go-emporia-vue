package vue

import "testing"

func TestFromBytesWrapper(t *testing.T) {
	yamlData := []byte(`
emporia:
  credentials:
    username: test@example.com
    password: secret
timezone: America/Chicago
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
}

func TestFromBytesWrapperEmpty(t *testing.T) {
	cfg, err := FromBytes([]byte{})
	if err != nil {
		t.Fatalf("FromBytes failed: %v", err)
	}

	// Should return default config
	if cfg.Timezone != "America/New_York" {
		t.Errorf("expected default timezone America/New_York, got %s", cfg.Timezone)
	}
}

func TestTypeAliases(t *testing.T) {
	// Verify type aliases work correctly
	cfg := &Config{
		Emporia: Emporia{
			Credentials: Credentials{
				Username: "test",
				Password: "pass",
			},
			Cognito: Cognito{
				ClientID: "client",
				Region:   "us-east-2",
				UserPool: "pool",
			},
		},
		Timezone: "UTC",
	}

	if cfg.Emporia.Credentials.Username != "test" {
		t.Errorf("expected username test, got %s", cfg.Emporia.Credentials.Username)
	}
}
