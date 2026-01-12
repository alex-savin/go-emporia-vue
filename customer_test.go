package vue

import (
	"testing"
	"time"
)

func TestCustomerStruct(t *testing.T) {
	customer := Customer{
		CustomerId: 12345,
		Email:      "test@example.com",
		FirstName:  "John",
		LastName:   "Doe",
		CreatedAt:  time.Now(),
	}

	if customer.CustomerId != 12345 {
		t.Fatalf("expected CustomerId 12345, got %d", customer.CustomerId)
	}
	if customer.Email != "test@example.com" {
		t.Fatalf("expected Email 'test@example.com', got %s", customer.Email)
	}
	if customer.FirstName != "John" {
		t.Fatalf("expected FirstName 'John', got %s", customer.FirstName)
	}
	if customer.LastName != "Doe" {
		t.Fatalf("expected LastName 'Doe', got %s", customer.LastName)
	}
}

func TestPartnerType(t *testing.T) {
	tests := []struct {
		name    string
		partner Partner
	}{
		{"Empty partner", Partner{}},
		{"Partner with key", Partner{"type": "subaru"}},
		{"Partner with multiple keys", Partner{"type": "generic", "id": 123}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.partner == nil {
				t.Fatal("partner should not be nil")
			}
		})
	}
}
