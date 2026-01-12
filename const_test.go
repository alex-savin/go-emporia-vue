package vue

import "testing"

func TestBatteryEndpointConstant(t *testing.T) {
	if apiURLs["API_GET_BATTERIES"] == "" {
		t.Fatalf("expected API_GET_BATTERIES to be set")
	}
}
