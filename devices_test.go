package vue

import (
	"testing"
)

func TestVueDeviceIsOutlet(t *testing.T) {
	tests := []struct {
		name     string
		outlet   *Outlet
		expected bool
	}{
		{"With outlet", &Outlet{DeviceGid: 1}, true},
		{"Without outlet", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{Outlet: tt.outlet}
			if device.IsOutlet() != tt.expected {
				t.Fatalf("expected IsOutlet=%v", tt.expected)
			}
		})
	}
}

func TestVueDeviceIsEvCharger(t *testing.T) {
	tests := []struct {
		name      string
		evCharger *EvCharger
		expected  bool
	}{
		{"With EV charger", &EvCharger{DeviceGid: 1}, true},
		{"Without EV charger", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{EvCharger: tt.evCharger}
			if device.IsEvCharger() != tt.expected {
				t.Fatalf("expected IsEvCharger=%v", tt.expected)
			}
		})
	}
}

func TestVueDeviceIsBattery(t *testing.T) {
	tests := []struct {
		name     string
		battery  *Battery
		expected bool
	}{
		{"With battery", &Battery{}, true},
		{"Without battery", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{Battery: tt.battery}
			if device.IsBattery() != tt.expected {
				t.Fatalf("expected IsBattery=%v", tt.expected)
			}
		})
	}
}

func TestVueDeviceIsEnergyMeter(t *testing.T) {
	tests := []struct {
		name      string
		outlet    *Outlet
		evCharger *EvCharger
		battery   *Battery
		expected  bool
	}{
		{"Neither outlet nor charger", nil, nil, nil, true},
		{"With outlet", &Outlet{}, nil, nil, false},
		{"With charger", nil, &EvCharger{}, nil, false},
		{"With both", &Outlet{}, &EvCharger{}, nil, false},
		{"With battery", nil, nil, &Battery{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{Outlet: tt.outlet, EvCharger: tt.evCharger, Battery: tt.battery}
			if device.IsEnergyMeter() != tt.expected {
				t.Fatalf("expected IsEnergyMeter=%v", tt.expected)
			}
		})
	}
}

func TestVueDeviceChargerRates(t *testing.T) {
	tests := []struct {
		name        string
		charger     *EvCharger
		wantRate    int
		wantMaxRate int
	}{
		{"configured", &EvCharger{ChargingRate: 16, MaxChargingRate: 48}, 16, 48},
		{"unset uses defaults", &EvCharger{}, defaultChargingRate, defaultMaxChargingRate},
		{"zero uses defaults", &EvCharger{ChargingRate: 0, MaxChargingRate: 0}, defaultChargingRate, defaultMaxChargingRate},
		{"nil charger uses defaults", nil, defaultChargingRate, defaultMaxChargingRate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{EvCharger: tt.charger}
			rate, maxRate := device.chargerRates()
			if rate != tt.wantRate || maxRate != tt.wantMaxRate {
				t.Fatalf("chargerRates() = (%d,%d), want (%d,%d)", rate, maxRate, tt.wantRate, tt.wantMaxRate)
			}
		})
	}
}

func TestVueDeviceType(t *testing.T) {
	tests := []struct {
		name      string
		outlet    *Outlet
		evCharger *EvCharger
		expected  string
	}{
		{"Outlet device", &Outlet{}, nil, "outlet"},
		{"EV charger device", nil, &EvCharger{}, "evcharger"},
		{"Energy monitor device", nil, nil, "energymonitor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{Outlet: tt.outlet, EvCharger: tt.evCharger}
			if device.Type() != tt.expected {
				t.Fatalf("expected Type=%s, got %s", tt.expected, device.Type())
			}
		})
	}
}

func TestVueDeviceHasNested(t *testing.T) {
	tests := []struct {
		name     string
		devices  []VueDevice
		expected bool
	}{
		{"No nested devices", nil, false},
		{"Empty nested devices", []VueDevice{}, true}, // Empty slice is non-nil, so HasNested returns true
		{"With nested devices", []VueDevice{{}}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			device := &VueDevice{Devices: tt.devices}
			if device.HasNested() != tt.expected {
				t.Fatalf("expected HasNested=%v", tt.expected)
			}
		})
	}
}

func TestChannelStruct(t *testing.T) {
	channel := Channel{
		DeviceGid:  1,
		Name:       "Test Channel",
		Number:     "1",
		Multiplier: 1.0,
		TypeGid:    1,
	}

	if channel.DeviceGid != 1 {
		t.Fatalf("expected DeviceGid 1, got %d", channel.DeviceGid)
	}
	if channel.Name != "Test Channel" {
		t.Fatalf("expected Name 'Test Channel', got %s", channel.Name)
	}
	if channel.Number != "1" {
		t.Fatalf("expected Number '1', got %s", channel.Number)
	}
}

func TestChannelTypeStruct(t *testing.T) {
	ct := ChannelType{
		ChannelTypeGid: 1,
		Description:    "Main",
		Selectable:     true,
	}

	if ct.ChannelTypeGid != 1 {
		t.Fatalf("expected ChannelTypeGid 1, got %d", ct.ChannelTypeGid)
	}
	if ct.Description != "Main" {
		t.Fatalf("expected Description 'Main', got %s", ct.Description)
	}
	if !ct.Selectable {
		t.Fatal("expected Selectable to be true")
	}
}

func TestConnectionStatusStruct(t *testing.T) {
	status := ConnectionStatus{
		DeviceGid:    1,
		Connected:    true,
		OfflineSince: "",
	}

	if !status.Connected {
		t.Fatal("expected Connected to be true")
	}
	if status.DeviceGid != 1 {
		t.Fatalf("expected DeviceGid 1, got %d", status.DeviceGid)
	}
}

func TestOutletStruct(t *testing.T) {
	outlet := Outlet{
		DeviceGid:        1,
		OutletOn:         true,
		ParentDeviceGid:  100,
		ParentChannelNum: "1",
	}

	if !outlet.OutletOn {
		t.Fatal("expected OutletOn to be true")
	}
	if outlet.DeviceGid != 1 {
		t.Fatalf("expected DeviceGid 1, got %d", outlet.DeviceGid)
	}
}

func TestEvChargerStruct(t *testing.T) {
	charger := EvCharger{
		DeviceGid:       1,
		ChargerOn:       true,
		ChargingRate:    32,
		MaxChargingRate: 48,
		Status:          "charging",
	}

	if !charger.ChargerOn {
		t.Fatal("expected ChargerOn to be true")
	}
	if charger.MaxChargingRate != 48 {
		t.Fatalf("expected MaxChargingRate 48, got %d", charger.MaxChargingRate)
	}
}

func TestChannelUsageStruct(t *testing.T) {
	usage := ChannelUsage{
		Name:          "Main",
		Percentage:    50.5,
		Usage:         100.0,
		DeviceGid:     1,
		ChannelNumber: "1",
	}

	if usage.Name != "Main" {
		t.Fatalf("expected Name 'Main', got %s", usage.Name)
	}
	if usage.Percentage != 50.5 {
		t.Fatalf("expected Percentage 50.5, got %f", usage.Percentage)
	}
}

func TestDeviceUsageStruct(t *testing.T) {
	usage := DeviceUsage{
		DeviceGid:     1,
		ChannelUsages: []ChannelUsage{},
	}

	if usage.DeviceGid != 1 {
		t.Fatalf("expected DeviceGid 1, got %d", usage.DeviceGid)
	}
}

func TestPropertiesStruct(t *testing.T) {
	props := Properties{
		DeviceGid:            1,
		DeviceName:           "Test Device",
		TimeZone:             "America/New_York",
		ZipCode:              "12345",
		BillingCycleStartDay: 1,
		UsageCentPerKwHour:   12.5,
	}

	if props.DeviceGid != 1 {
		t.Fatalf("expected DeviceGid 1, got %d", props.DeviceGid)
	}
	if props.DeviceName != "Test Device" {
		t.Fatalf("expected DeviceName 'Test Device', got %s", props.DeviceName)
	}
}

func TestScheduleStruct(t *testing.T) {
	schedule := Schedule{
		ScheduleGid: 1,
		DeviceGid:   1,
		Enabled:     true,
		DeviceOn:    true,
		Anchor:      "MIDNIGHT",
	}

	if schedule.ScheduleGid != 1 {
		t.Fatalf("expected ScheduleGid 1, got %d", schedule.ScheduleGid)
	}
	if !schedule.Enabled {
		t.Fatal("expected Enabled to be true")
	}
}

func TestInfoStruct(t *testing.T) {
	info := Info{
		ID:           12345,
		Manufacturer: "Emporia",
		Model:        "Vue 2",
		Firmware:     "1.2.3",
	}

	if info.ID != 12345 {
		t.Fatalf("expected ID 12345, got %d", info.ID)
	}
	if info.Manufacturer != "Emporia" {
		t.Fatalf("expected Manufacturer 'Emporia', got %s", info.Manufacturer)
	}
	if info.Model != "Vue 2" {
		t.Fatalf("expected Model 'Vue 2', got %s", info.Model)
	}
	if info.Firmware != "1.2.3" {
		t.Fatalf("expected Firmware '1.2.3', got %s", info.Firmware)
	}
}
