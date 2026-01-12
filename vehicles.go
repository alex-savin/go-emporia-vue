package vue

// VueVehicle represents a connected electric vehicle linked to the Emporia account.
type VueVehicle struct {
	VehicleGid  int    `json:"vehicleGid"`  //
	Vendor      string `json:"vendor"`      //
	ApiID       string `json:"apiId"`       //
	DisplayName string `json:"displayName"` //
	LoadGid     string `json:"loadGid"`     //
	Make        string `json:"make"`        //
	Model       string `json:"model"`       //
	Year        string `json:"year"`
}

// VehicleStatus contains the current charging and battery state for a vehicle.
type VehicleStatus struct {
	VehicleGid              int         `json:"vehicleGid"`              //
	VehicleState            string      `json:"vehicleState"`            //
	BatteryLevel            int         `json:"batteryLevel"`            //
	BatteryRange            int         `json:"batteryRange"`            //
	ChargingState           string      `json:"chargingState"`           //
	ChargeLimitPercent      int         `json:"chargeLimitPercent"`      //
	MinutesToFullCharge     int         `json:"minutesToFullCharge"`     //
	ChargeCurrentRequest    interface{} `json:"chargeCurrentRequest"`    //
	ChargeCurrentRequestMax interface{} `json:"chargeCurrentRequestMax"` //
}
