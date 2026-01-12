package vue

import (
	"encoding/json"
	"strconv"
	"time"
)

// VueDevice represents an Emporia energy monitoring device such as a Vue monitor,
// smart outlet, EV charger, or battery.
type VueDevice struct {
	DeviceGid            int               `json:"deviceGid"`                    //
	ManufacturerDeviceID string            `json:"manufacturerDeviceId"`         //
	Model                string            `json:"model"`                        //
	Firmware             string            `json:"firmware,omitempty"`           //
	Devices              []VueDevice       `json:"devices"`                      //
	Channels             []Channel         `json:"channels"`                     //
	ConnectionStatus     *ConnectionStatus `json:"deviceConnected,omitempty"`    //
	Properties           *Properties       `json:"locationProperties,omitempty"` //
	Outlet               *Outlet           `json:"outlet,omitempty"`             //
	EvCharger            *EvCharger        `json:"evCharger,omitempty"`          //
	Battery              *Battery          `json:"battery,omitempty"`            //
	ParentDeviceGid      *int              `json:"parentDeviceGid,omitempty"`    //
	ParentChannelNum     *int              `json:"parentChannelNum,omitempty"`   //
	vue                  *Client
}

// Info contains manufacturer and firmware details for a device.
type Info struct {
	ID           int
	Manufacturer string
	Model        string
	Firmware     string
}

// On turns on the device. For outlets, this enables power output.
// For EV chargers, this starts charging at the configured rate.
func (d *VueDevice) On() {
	defer d.vue.timeTrack("[TIMETRK] Executing GetDeviceUsage Request")()

	var resp []byte
	if d.IsOutlet() {
		params := map[string]string{
			"deviceGid": strconv.Itoa(d.DeviceGid),
			"outletOn":  "true",
		}
		resp = d.vue.execute(apiURLs["API_OUTLET"], "PUT", params, true)
	}
	if d.IsEvCharger() {
		params := map[string]string{
			"deviceGid":               strconv.Itoa(d.DeviceGid),
			"chargerOn":               "true",
			"chargingRate":            "32",
			"maxChargingRate":         "40",
			"offPeakSchedulesEnabled": "false",
		}
		resp = d.vue.execute(apiURLs["API_CHARGER"], "PUT", params, true)
	}
	if resp == nil {
		d.vue.log.Error("toggle on request failed", "device", d.DeviceGid)
		return
	}

	switch {
	case d.IsOutlet():
		var outlet Outlet
		if err := json.Unmarshal(resp, &outlet); err != nil {
			d.vue.log.Error("cannot parse outlet toggle response", "error", err)
		}
	case d.IsEvCharger():
		var charger EvCharger
		if err := json.Unmarshal(resp, &charger); err != nil {
			d.vue.log.Error("cannot parse charger toggle response", "error", err)
		}
	}
}

// Off turns off the device. For outlets, this disables power output.
// For EV chargers, this stops charging.
func (d *VueDevice) Off() {
	defer d.vue.timeTrack("[TIMETRK] Executing GetDeviceUsage Request")()

	var resp []byte
	if d.IsOutlet() {
		params := map[string]string{
			"deviceGid": strconv.Itoa(d.DeviceGid),
			"outletOn":  "false",
		}
		resp = d.vue.execute(apiURLs["API_OUTLET"], "PUT", params, true)
	}
	if d.IsEvCharger() {
		params := map[string]string{
			"deviceGid":               strconv.Itoa(d.DeviceGid),
			"chargerOn":               "false",
			"chargingRate":            "32",
			"maxChargingRate":         "40",
			"offPeakSchedulesEnabled": "false",
		}
		resp = d.vue.execute(apiURLs["API_CHARGER"], "PUT", params, true)
	}
	if resp == nil {
		d.vue.log.Error("toggle off request failed", "device", d.DeviceGid)
		return
	}

	switch {
	case d.IsOutlet():
		var outlet Outlet
		if err := json.Unmarshal(resp, &outlet); err != nil {
			d.vue.log.Error("cannot parse outlet toggle response", "error", err)
		}
	case d.IsEvCharger():
		var charger EvCharger
		if err := json.Unmarshal(resp, &charger); err != nil {
			d.vue.log.Error("cannot parse charger toggle response", "error", err)
		}
	}
}

// Usage retrieves channel usage data for this device at the specified time scale.
func (d *VueDevice) Usage(scope string) []ChannelUsage {
	usage := d.vue.GetDeviceUsage(d.DeviceGid, scope)
	if len(usage.DeviceListUsages.Devices) == 0 {
		return nil
	}
	return usage.DeviceListUsages.Devices[0].ChannelUsages
}

// RefreshProperties reloads location properties for the device.
func (d *VueDevice) RefreshProperties() *Properties {
	if d.vue == nil {
		return nil
	}
	props := d.vue.GetDeviceProperties(d.DeviceGid)
	d.Properties = props
	return props
}

// Info returns manufacturer and firmware details for this device.
func (d *VueDevice) Info() Info {
	var info Info
	if val, ok := Models[d.Model]; ok {
		info = Info{
			ID:           d.DeviceGid,
			Manufacturer: val["manufacturer"],
			Model:        val["model"],
			Firmware:     d.Firmware,
		}
	}

	return info
}

// Properties contains location and billing configuration for a device.
type Properties struct {
	TimeZone              string       `json:"timeZone"`
	LatitudeLongitude     interface{}  `json:"latitudeLongitude,omitempty"`
	UtilityRateGid        *interface{} `json:"utilityRateGid,omitempty"`
	DeviceGid             int          `json:"deviceGid"`
	DeviceName            string       `json:"deviceName"`
	ZipCode               string       `json:"zipCode,omitempty"`
	BillingCycleStartDay  int          `json:"billingCycleStartDay"`
	UsageCentPerKwHour    float32      `json:"usageCentPerKwHour"`
	PeakDemandDollarPerKw float32      `json:"peakDemandDollarPerKw"`
	LocationInformation   struct {
		AirConditioning bool   `json:"airConditioning,string,omitempty"`
		HeatSource      string `json:"heatSource"`
		LocationSqFt    int    `json:"locationSqFt,string"`
		NumElectricCars int    `json:"numElectricCars,string"`
		LocationType    string `json:"locationType"`
		NumPeople       int    `json:"numPeople,string"`
		SwimmingPool    bool   `json:"swimmingPool,string"`
		HotTub          bool   `json:"hotTub,string"`
	} `json:"locationInformation"`
}

// Channel represents an individual circuit or sensor channel on a device.
type Channel struct {
	DeviceGid  int     `json:"deviceGid"`
	Name       string  `json:"name"`
	Number     string  `json:"channelNum"`
	Multiplier float64 `json:"channelMultiplier"`
	TypeGid    int     `json:"channelTypeGid"`
	Type       string  `json:"type,omitempty"`
}

// ChannelType describes available channel type metadata.
type ChannelType struct {
	ChannelTypeGid int    `json:"channelTypeGid"`
	Description    string `json:"description"`
	Selectable     bool   `json:"selectable"`
}

// Usage returns empty usage data for this channel (placeholder implementation).
func (c *Channel) Usage() *ChannelUsage {
	return &ChannelUsage{}
}

// ConnectionStatus indicates whether a device is currently online.
type ConnectionStatus struct {
	DeviceGid    int    `json:"deviceGid"`
	Connected    bool   `json:"connected"`
	OfflineSince string `json:"offlineSince,omitempty"`
}

// Schedule defines an on/off schedule for an outlet or charger.
type Schedule struct {
	ScheduleGid      int      // "scheduleGid": 30303,
	LoadGid          int      // "loadGid": 50505,
	DeviceOn         bool     // "deviceOn": true,
	Enabled          bool     // "enabled": true,
	Anchor           string   // "anchor": "MIDNIGHT",
	OffsetFromAnchor string   // "offsetFromAnchor": "06:00:00",
	DaysOfWeek       []string // "daysOfWeek": [ "MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY", "SATURDAY", "SUNDAY"],
	DeviceGid        int      // Device GID (may be null for load-level schedules)
}

// ChannelChart contains historical usage data for charting.
type ChannelChart struct {
	UsageList         []float64 `json:"usageList"`
	FirstUsageInstant string    `json:"firstUsageInstant"`
}

// DeviceListUsages is the top-level response wrapper for device usage queries.
type DeviceListUsages struct {
	DeviceListUsages DevicesUsage `json:"deviceListUsages"`
}

// DevicesUsage contains usage data for multiple devices at a point in time.
type DevicesUsage struct {
	Instant    time.Time     `json:"instant"`
	Scale      string        `json:"scale"`
	EnergyUnit string        `json:"energyUnit"`
	Devices    []DeviceUsage `json:"devices"`
}

// DeviceUsage contains per-channel usage breakdown for a single device.
type DeviceUsage struct {
	DeviceGid     int            `json:"deviceGid"`
	ChannelUsages []ChannelUsage `json:"channelUsages"`
}

// ChannelUsage represents energy consumption for a single channel.
type ChannelUsage struct {
	Name          string         `json:"name"`
	Usage         float64        `json:"usage"`
	DeviceGid     int            `json:"deviceGid"`
	ChannelNumber string         `json:"channelNum"`
	Percentage    float64        `json:"percentage"`
	NestedDevices *[]interface{} `json:"nestedDevices,omitempty"`
}

// HasNested returns true if this device has nested sub-devices.
func (v *VueDevice) HasNested() bool {
	return !isNil(v.Devices)
}

// Outlet represents an Emporia smart plug/outlet.
type Outlet struct {
	DeviceGid        int         `json:"deviceGid"`
	OutletOn         bool        `json:"outletOn"`
	ParentDeviceGid  int         `json:"parentDeviceGid"`
	ParentChannelNum string      `json:"parentChannelNum"`
	Schedules        *[]Schedule `json:"schedules,omitempty"`
}

// IsEnergyMeter returns true if this device is a Vue energy monitor (not an outlet or charger).
func (v *VueDevice) IsEnergyMeter() bool {
	if isNil(v.Outlet) && isNil(v.EvCharger) {
		return true
	}
	return false
}

// IsOutlet returns true if this device is an Emporia smart outlet.
func (v *VueDevice) IsOutlet() bool {
	return !isNil(v.Outlet)
}

// EvCharger represents an Emporia EV charging station.
type EvCharger struct {
	DeviceGid               int     `json:"deviceGid"`
	Message                 string  `json:"message"`
	Status                  string  `json:"status"`
	Icon                    string  `json:"icon"`
	IconLabel               string  `json:"iconLabel"`
	IconDetailText          string  `json:"iconDetailText"`
	FaultText               *string `json:"faultText,omitempty"`
	ChargerOn               bool    `json:"chargerOn"`
	ChargingRate            int     `json:"chargingRate"`
	MaxChargingRate         int     `json:"maxChargingRate"`
	OffPeakSchedulesEnabled bool    `json:"offPeakSchedulesEnabled"`
}

// IsEvCharger returns true if this device is an EV charger.
func (v *VueDevice) IsEvCharger() bool {
	return !isNil(v.EvCharger)
}

// Thermostat represents an Emporia smart thermostat (placeholder).
type Thermostat struct{}

// Battery represents an Emporia home battery system (placeholder).
type Battery struct{}

// IsBattery returns true if this device is a home battery.
func (v *VueDevice) IsBattery() bool {
	return !isNil(v.Battery)
}

// Type returns a human-readable device type string.
func (d *VueDevice) Type() string {
	if d.IsOutlet() {
		return "outlet"
	}
	if d.IsEvCharger() {
		return "evcharger"
	}
	return "energymonitor"
}
