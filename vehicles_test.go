package vue

import (
	"testing"
)

func TestVueVehicleStruct(t *testing.T) {
	vehicle := VueVehicle{
		VehicleGid:  1,
		DisplayName: "Test Vehicle",
		Make:        "Tesla",
		Model:       "Model 3",
		Year:        "2023",
		Vendor:      "tesla",
		ApiID:       "abc123",
		LoadGid:     "load1",
	}

	if vehicle.VehicleGid != 1 {
		t.Fatalf("expected VehicleGid 1, got %d", vehicle.VehicleGid)
	}
	if vehicle.DisplayName != "Test Vehicle" {
		t.Fatalf("expected DisplayName 'Test Vehicle', got %s", vehicle.DisplayName)
	}
	if vehicle.Make != "Tesla" {
		t.Fatalf("expected Make 'Tesla', got %s", vehicle.Make)
	}
	if vehicle.Model != "Model 3" {
		t.Fatalf("expected Model 'Model 3', got %s", vehicle.Model)
	}
	if vehicle.Year != "2023" {
		t.Fatalf("expected Year '2023', got %s", vehicle.Year)
	}
}

func TestVehicleStatusStruct(t *testing.T) {
	status := VehicleStatus{
		VehicleGid:          1,
		BatteryLevel:        80,
		ChargingState:       "Charging",
		VehicleState:        "online",
		ChargeLimitPercent:  90,
		MinutesToFullCharge: 60,
		BatteryRange:        200,
	}

	if status.VehicleGid != 1 {
		t.Fatalf("expected VehicleGid 1, got %d", status.VehicleGid)
	}
	if status.BatteryLevel != 80 {
		t.Fatalf("expected BatteryLevel 80, got %d", status.BatteryLevel)
	}
	if status.ChargingState != "Charging" {
		t.Fatalf("expected ChargingState 'Charging', got %s", status.ChargingState)
	}
	if status.VehicleState != "online" {
		t.Fatalf("expected VehicleState 'online', got %s", status.VehicleState)
	}
}

func TestVehicleStatusChargingStates(t *testing.T) {
	tests := []struct {
		name          string
		chargingState string
	}{
		{"Not charging", "Disconnected"},
		{"Charging", "Charging"},
		{"Complete", "Complete"},
		{"Stopped", "Stopped"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := VehicleStatus{
				VehicleGid:    1,
				ChargingState: tt.chargingState,
			}
			if status.ChargingState != tt.chargingState {
				t.Fatalf("expected ChargingState '%s', got '%s'", tt.chargingState, status.ChargingState)
			}
		})
	}
}

func TestVehicleStatusVehicleStates(t *testing.T) {
	tests := []struct {
		name         string
		vehicleState string
	}{
		{"Online", "online"},
		{"Offline", "offline"},
		{"Asleep", "asleep"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := VehicleStatus{
				VehicleGid:   1,
				VehicleState: tt.vehicleState,
			}
			if status.VehicleState != tt.vehicleState {
				t.Fatalf("expected VehicleState '%s', got '%s'", tt.vehicleState, status.VehicleState)
			}
		})
	}
}

func TestVehicleStatusBatteryEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		batteryLevel int
	}{
		{"Empty battery", 0},
		{"Full battery", 100},
		{"Half battery", 50},
		{"Low battery", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := VehicleStatus{
				VehicleGid:   1,
				BatteryLevel: tt.batteryLevel,
			}
			if status.BatteryLevel != tt.batteryLevel {
				t.Fatalf("expected BatteryLevel %d, got %d", tt.batteryLevel, status.BatteryLevel)
			}
		})
	}
}
