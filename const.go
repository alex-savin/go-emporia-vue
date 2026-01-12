package vue

// /Notifications/GetAll?customerGid=0000

var apiURLs = map[string]string{
	"API_ROOT":              "https://api.emporiaenergy.com",
	"API_CUSTOMER":          "/customers?email={email}",
	"API_CUSTOMER_DEVICES":  "/customers/devices",
	"API_CHANNELS":          "/customers/devices/{deviceGid}/channels",
	"API_CHANNEL_TYPES":     "/customers/devices/channelTypes",
	"API_DEVICES_USAGE":     "/AppAPI?apiMethod=getDeviceListUsages&deviceGids={deviceGids}&instant={instant}&scale={scale}&energyUnit={unit}",
	"API_CHART_USAGE":       "/AppAPI?apiMethod=getChartUsage&deviceGid={deviceGid}&channel={channel}&start={start}&end={end}&scale={scale}&energyUnit={unit}",
	"API_DEVICE_PROPERTIES": "/devices/{deviceGid}/locationProperties",
	"API_OUTLET":            "/devices/outlet",
	"API_GET_OUTLETS":       "/customers/outlets",
	"API_CHARGER":           "/devices/evcharger",
	"API_GET_CHARGERS":      "/customers/evchargers",
	"API_THERMOSTAT":        "/devices/thermostat",
	"API_GET_THERMOSTATS":   "/customers/thermostats?customerGid={customerGid}",
	"API_VEHICLES":          "/customers/vehicles",
	"API_VEHICLE_STATUS":    "/vehicles/v2/settings?vehicleGid={vehicleGid}",
	"API_PARTNERS":          "/customers/partners?customerGid={customerGid}",
	"API_REMOTE_CONFIG":     "/remoteconfig?appVersion={appVersion}",
	"API_BATTERY":           "/devices/battery",
	"API_GET_BATTERIES":     "/customers/batteries",
	"API_MAINTENANCE":       "https://s3.amazonaws.com/com.emporiaenergy.manual.ota/maintenance/maintenance.json",
}

var Durations = map[string]string{
	"second":   "1S",
	"minute":   "1MIN",
	"15minute": "15MIN",
	"hour":     "1H",
	"day":      "1D",
	"week":     "1W",
	"month":    "1MON",
	"year":     "1Y",
}

var Units = map[string]string{
	"kwh":     "KilowattHours",
	"volts":   "Voltage",
	"dollars": "Dollars",
	"ah":      "AmpHours",
	"trees":   "Trees",
	"gallons": "GallonsOfGas",
	"miles":   "MilesDriven",
	"carbon":  "Carbon",
}

var Models = map[string]map[string]string{
	"VUE001": {
		"manufacturer": "Emporia Vue",
		"model":        "Energy Monitor (Gen 1)",
	},
	"VUE002": {
		"manufacturer": "Emporia Vue",
		"model":        "Energy Monitor (Gen 2)",
	},
	"VUE02": { // alias seen in simulator
		"manufacturer": "Emporia Vue",
		"model":        "Energy Monitor (Gen 2)",
	},
	"VUE003": {
		"manufacturer": "Emporia Vue",
		"model":        "Energy Monitor (Gen 3)",
	},
	"WAT001": {
		"manufacturer": "Emporia Vue",
		"model":        "Energy Monitor (Gen 1) Internal Module",
	},
	"SSO001": {
		"manufacturer": "Emporia Vue",
		"model":        "Smart Plug",
	},
	"EVSE001": {
		"manufacturer": "Emporia Vue",
		"model":        "EV Charger",
	},
	"BATT001": {
		"manufacturer": "Emporia Vue",
		"model":        "Battery",
	},
	"THER001": {
		"manufacturer": "Emporia Vue",
		"model":        "Thermostat",
	},
}
