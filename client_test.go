package vue

import (
	"log/slog"
	"os"
	"testing"
)

func TestNewClientNilConfig(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	_, err := NewClient(logger, nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestNewClientMissingCredentials(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &Config{}
	_, err := NewClient(logger, config)
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestNewClientMissingUsername(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &Config{
		Emporia: Emporia{
			Credentials: Credentials{
				Password: "testpass",
			},
		},
	}
	_, err := NewClient(logger, config)
	if err == nil {
		t.Fatal("expected error for missing username")
	}
}

func TestNewClientMissingPassword(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	config := &Config{
		Emporia: Emporia{
			Credentials: Credentials{
				Username: "testuser",
			},
		},
	}
	_, err := NewClient(logger, config)
	if err == nil {
		t.Fatal("expected error for missing password")
	}
}

func TestTokenStruct(t *testing.T) {
	tok := &token{
		Access:     "access123",
		ID:         "id456",
		Refresh:    "refresh789",
		Expiration: 1234567890,
	}

	if tok.Access != "access123" {
		t.Fatalf("expected Access 'access123', got %s", tok.Access)
	}
	if tok.ID != "id456" {
		t.Fatalf("expected ID 'id456', got %s", tok.ID)
	}
	if tok.Refresh != "refresh789" {
		t.Fatalf("expected Refresh 'refresh789', got %s", tok.Refresh)
	}
	if tok.Expiration != 1234567890 {
		t.Fatalf("expected Expiration 1234567890, got %d", tok.Expiration)
	}
}

func TestCredentialsStruct(t *testing.T) {
	creds := &credentials{
		username: "user@example.com",
		password: "secret",
		token:    &token{},
		cognito: cognito{
			clientID: "client123",
			region:   "us-east-1",
			userPool: "pool123",
		},
	}

	if creds.username != "user@example.com" {
		t.Fatalf("expected username 'user@example.com', got %s", creds.username)
	}
	if creds.password != "secret" {
		t.Fatalf("expected password 'secret', got %s", creds.password)
	}
	if creds.cognito.clientID != "client123" {
		t.Fatalf("expected clientID 'client123', got %s", creds.cognito.clientID)
	}
	if creds.cognito.region != "us-east-1" {
		t.Fatalf("expected region 'us-east-1', got %s", creds.cognito.region)
	}
}

func TestClientStruct(t *testing.T) {
	client := &Client{
		devices:   make(map[int]*VueDevice),
		vehicles:  []*VueVehicle{},
		connected: false,
	}

	if client.connected {
		t.Fatal("expected connected to be false")
	}
	if client.devices == nil {
		t.Fatal("expected devices map to be initialized")
	}
	if client.vehicles == nil {
		t.Fatal("expected vehicles slice to be initialized")
	}
}
