package vue

import (
	"time"
)

// Customer represents an Emporia account holder with associated devices.
type Customer struct {
	CustomerId int          `json:"customerGid"`
	FirstName  string       `json:"firstName"`
	LastName   string       `json:"lastName"`
	Email      string       `json:"email"`
	CreatedAt  time.Time    `json:"createdAt"`
	Devices    []*VueDevice `json:"devices"`
	vue        *Client
}

// Partner holds loosely-typed partner integration data (e.g., Tesla, Subaru)
// returned by the Emporia API.
type Partner map[string]interface{}

// Update refreshes the customer's device list from the API.
func (c *Customer) Update() {

}
