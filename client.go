package vue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"gopkg.in/yaml.v3"
	"resty.dev/v3"
)

// credentials holds authentication data for the Emporia API.
type credentials struct {
	username string
	password string
	token    *token
	cognito  cognito
}

type token struct {
	Access     string `yaml:"access_token"`
	ID         string `yaml:"id_token"`
	Refresh    string `yaml:"refresh_token"`
	Expiration int64  `yaml:"token_expiration"`
}

type cognito struct {
	clientID string
	region   string
	userPool string
}

// credentialsFilePath resolves the token cache location lazily, on each use,
// so that callers which set EMPORIA_CREDENTIALS_FILE at runtime (e.g. from
// main()) are honored. Resolving it in a package-level var initializer would
// capture the value at program init, before such callers run.
func credentialsFilePath() string {
	if v := os.Getenv("EMPORIA_CREDENTIALS_FILE"); v != "" {
		return v
	}
	return "./.credentials.yaml"
}

// Client is a concurrency-safe Emporia Vue API client. Its read methods may be
// called from multiple goroutines simultaneously.
type Client struct {
	mu          sync.RWMutex // guards devices, vehicles, connected
	devices     map[int]*VueDevice
	vehicles    []*VueVehicle
	connected   bool
	tokenMu     sync.Mutex // serializes token read/refresh/persist
	credentials *credentials
	httpClient  *resty.Client
	ctx         context.Context
	log         *slog.Logger
}

// context returns the base context for outbound calls (defaults to Background).
func (c *Client) context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return context.Background()
}

// SetContext sets the base context used for outbound API calls, allowing
// callers to cancel in-flight requests (e.g. on shutdown). Call it before
// issuing concurrent requests.
func (c *Client) SetContext(ctx context.Context) {
	c.ctx = ctx
}

// NewClient creates a new Emporia Vue API client.
// If logger is nil, the logger from config will be used.
func NewClient(logger *slog.Logger, cfg *Config) (*Client, error) {
	if cfg == nil {
		return nil, errors.New("config must not be nil")
	}
	if cfg.Emporia.Credentials.Username == "" || cfg.Emporia.Credentials.Password == "" {
		return nil, errors.New("emporia credentials are required")
	}

	// Use config logger if none provided
	if logger == nil {
		logger = cfg.Logger
	}

	credentials := &credentials{
		username: cfg.Emporia.Credentials.Username,
		password: cfg.Emporia.Credentials.Password,
		token:    &token{},
		cognito: cognito{
			clientID: cfg.Emporia.Cognito.ClientID,
			region:   cfg.Emporia.Cognito.Region,
			userPool: cfg.Emporia.Cognito.UserPool,
		},
	}
	client := &Client{
		credentials: credentials,
		ctx:         context.Background(),
		log:         logger,
	}

	credFile := credentialsFilePath()
	if _, err := os.Stat(credFile); err == nil {
		t, err := os.ReadFile(credFile)
		if err != nil {
			client.log.Error("cannot read .credentials.yaml", "error", err)
		}

		err = yaml.Unmarshal(t, credentials.token)
		if err != nil {
			client.log.Error("cannot unmarshal token from yaml", "error", err)
		}
	}

	// Bootstrap runs single-threaded before the client is shared, so the
	// *Locked workers are called directly.
	if client.credentials.token.Expiration == 0 && client.credentials.token.Access == "" && client.credentials.token.Refresh == "" && client.credentials.token.ID == "" {
		if err := client.authLocked(); err != nil {
			return nil, err
		}
	}

	if client.credentials.token.Expiration < time.Now().Unix() {
		if err := client.refreshTokenLocked(); err != nil {
			return nil, err
		}
	}

	// Retry/backoff and token-refresh-on-401 are handled by execute(); the
	// resty client is intentionally left without its own retry layer to avoid
	// compounding attempts (and to avoid retrying non-retryable 4xx responses).
	httpClient := resty.New()
	httpClient.SetBaseURL(apiURLs["API_ROOT"]).
		SetHeader("authtoken", client.credentials.token.ID)

	client.httpClient = httpClient
	client.GetDevices()

	return client, nil
}

// authLocked performs Cognito user/password authentication and stores the
// token locally. The caller must hold tokenMu (or guarantee single-threaded
// use, as in NewClient).
func (c *Client) authLocked() error {
	defer c.timeTrack("Authentication")()

	ctx := c.context()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.credentials.cognito.region))
	if err != nil {
		c.log.Error("cannot load AWS config", "error", err)
		return err
	}

	cip := cognitoidentityprovider.NewFromConfig(cfg)
	authResp, err := cip.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeUserPasswordAuth,
		AuthParameters: map[string]string{
			"USERNAME": c.credentials.username,
			"PASSWORD": c.credentials.password,
		},
		ClientId: aws.String(c.credentials.cognito.clientID),
	})
	if err != nil {
		c.log.Error("cannot initiate cognito auth", "error", err)
		return err
	}

	c.credentials.token.Access = aws.ToString(authResp.AuthenticationResult.AccessToken)
	c.credentials.token.ID = aws.ToString(authResp.AuthenticationResult.IdToken)
	c.credentials.token.Refresh = aws.ToString(authResp.AuthenticationResult.RefreshToken)
	c.credentials.token.Expiration = time.Now().Add(time.Second * time.Duration(authResp.AuthenticationResult.ExpiresIn)).Unix()

	if err := c.persistToken(); err != nil {
		return err
	}

	c.log.Info("successfully received tokens")
	return nil
}

// refreshTokenLocked refreshes the access token using the stored refresh token.
// The caller must hold tokenMu.
func (c *Client) refreshTokenLocked() error {
	defer c.timeTrack("RefreshToken")()

	ctx := c.context()
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(c.credentials.cognito.region))
	if err != nil {
		c.log.Error("cannot load AWS config", "error", err)
		return err
	}

	cip := cognitoidentityprovider.NewFromConfig(cfg)
	authResp, err := cip.InitiateAuth(ctx, &cognitoidentityprovider.InitiateAuthInput{
		AuthFlow: types.AuthFlowTypeRefreshTokenAuth,
		AuthParameters: map[string]string{
			"REFRESH_TOKEN": c.credentials.token.Refresh,
		},
		ClientId: aws.String(c.credentials.cognito.clientID),
	})
	if err != nil {
		c.log.Error("cannot initiate cognito refresh", "error", err)
		return err
	}

	c.credentials.token.Access = aws.ToString(authResp.AuthenticationResult.AccessToken)
	c.credentials.token.ID = aws.ToString(authResp.AuthenticationResult.IdToken)
	c.credentials.token.Expiration = time.Now().Add(time.Second * time.Duration(authResp.AuthenticationResult.ExpiresIn)).Unix()

	if err := c.persistToken(); err != nil {
		return err
	}

	c.log.Info("successfully refreshed token")
	return nil
}

// ensureValidToken refreshes the token if it has expired. Concurrent callers
// serialize on tokenMu and the double-check ensures only one refresh happens.
func (c *Client) ensureValidToken() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	if c.credentials.token.Expiration < time.Now().Unix() {
		return c.refreshTokenLocked()
	}
	return nil
}

// refreshToken refreshes the token, serializing against other token operations.
func (c *Client) refreshToken() error {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	return c.refreshTokenLocked()
}

// tokenID returns the current ID token under tokenMu.
func (c *Client) tokenID() string {
	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	return c.credentials.token.ID
}

// persistToken writes the current token to disk. The caller must hold tokenMu.
func (c *Client) persistToken() error {
	token, err := yaml.Marshal(c.credentials.token)
	if err != nil {
		c.log.Error("cannot marshal token to yaml from struct", "error", err)
		return err
	}

	if err := os.WriteFile(credentialsFilePath(), token, 0o600); err != nil {
		c.log.Error("cannot write credentials file", "error", err)
		return err
	}

	return nil
}

// GetCustomer retrieves the customer profile associated with the credentials.
func (c *Client) GetCustomer() *Customer {
	defer c.timeTrack("GetCustomer")()

	params := map[string]string{"email": c.credentials.username}
	resp := c.execute(apiURLs["API_CUSTOMER"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetCustomer")
		return nil
	}

	var customer Customer
	err := json.Unmarshal(resp, &customer)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetCustomer", "error", err)
	}
	customer.vue = c

	return &customer
}

// GetDevices fetches all devices associated with the customer account.
func (c *Client) GetDevices() []*VueDevice {
	defer c.timeTrack("GetDevices")()

	resp := c.execute(apiURLs["API_CUSTOMER_DEVICES"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetDevices")
		return nil
	}

	raw := sanitizeJSON(resp)

	var customer Customer
	err := json.Unmarshal(raw, &customer)
	if err != nil {
		if syn, ok := err.(*json.SyntaxError); ok {
			c.log.Error("cannot unmarshal json", "func", "GetDevices", "error", err, "offset", syn.Offset, "context", jsonErrorContext(raw, syn.Offset))
		} else {
			c.log.Error("cannot unmarshal json", "func", "GetDevices", "error", err)
		}
	}

	devices := make(map[int]*VueDevice)
	for _, device := range customer.Devices {
		device.vue = c
		devices[device.DeviceGid] = device
	}

	c.mu.Lock()
	c.devices = devices
	c.mu.Unlock()

	return customer.Devices
}

// GetEnergyMonitors returns only Vue energy monitor devices (excludes outlets, chargers, batteries).
func (c *Client) GetEnergyMonitors() []*VueDevice {
	defer c.timeTrack("GetEnergyMonitors")()

	c.mu.RLock()
	empty := c.devices == nil
	c.mu.RUnlock()
	if empty {
		c.GetDevices()
	}

	var monitors []*VueDevice
	for _, d := range c.snapshotDevices() {
		if !d.IsOutlet() && !d.IsEvCharger() && !d.IsBattery() {
			monitors = append(monitors, d)
		}
	}

	return monitors
}

// snapshotDevices returns a copy of the current device slice under the read lock.
func (c *Client) snapshotDevices() []*VueDevice {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]*VueDevice, 0, len(c.devices))
	for _, d := range c.devices {
		out = append(out, d)
	}
	return out
}

// GetOutlets retrieves all Emporia smart outlets for the customer.
func (c *Client) GetOutlets() []*Outlet {
	defer c.timeTrack("GetOutlets")()

	resp := c.execute(apiURLs["API_GET_OUTLETS"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetOutlets")
		return nil
	}

	var outlets []*Outlet
	err := json.Unmarshal(resp, &outlets)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetOutlets", "error", err)
	}

	return outlets
}

// GetEvChargers retrieves all Emporia EV chargers for the customer.
func (c *Client) GetEvChargers() []*EvCharger {
	defer c.timeTrack("GetEvChargers")()

	resp := c.execute(apiURLs["API_GET_CHARGERS"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetEvChargers")
		return nil
	}

	var chargers []*EvCharger
	err := json.Unmarshal(resp, &chargers)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetEvChargers", "error", err)
	}

	return chargers
}

// GetBatteries retrieves all Emporia home battery systems for the customer.
func (c *Client) GetBatteries() []*Battery {
	defer c.timeTrack("GetBatteries")()

	resp := c.execute(apiURLs["API_GET_BATTERIES"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetBatteries")
		return nil
	}

	var batteries []*Battery
	err := json.Unmarshal(resp, &batteries)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetBatteries", "error", err)
	}

	return batteries
}

// GetThermostats retrieves all Emporia smart thermostats for the customer.
func (c *Client) GetThermostats() []*Thermostat {
	defer c.timeTrack("GetThermostats")()

	params := map[string]string{"customerGid": strconv.Itoa(c.GetCustomer().CustomerId)}
	resp := c.execute(apiURLs["API_GET_THERMOSTATS"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetThermostats")
		return nil
	}

	var thermostats []*Thermostat
	err := json.Unmarshal(resp, &thermostats)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetThermostats", "error", err)
	}

	return thermostats
}

// GetVehicles retrieves all connected electric vehicles for the customer.
func (c *Client) GetVehicles() []*VueVehicle {
	defer c.timeTrack("GetVehicles")()

	resp := c.execute(apiURLs["API_VEHICLES"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetVehicles")
		return nil
	}

	var vehicles []*VueVehicle
	err := json.Unmarshal(resp, &vehicles)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetVehicles", "error", err)
	}

	c.mu.Lock()
	c.vehicles = vehicles
	c.mu.Unlock()

	return vehicles
}

// GetVehicleStatus returns the current status for a specific vehicle.
func (c *Client) GetVehicleStatus(vehicleGid int) *VehicleStatus {
	defer c.timeTrack("GetVehicleStatus")()

	params := map[string]string{"vehicleGid": strconv.Itoa(vehicleGid)}
	resp := c.execute(apiURLs["API_VEHICLE_STATUS"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetVehicleStatus")
		return nil
	}

	var status VehicleStatus
	if err := json.Unmarshal(resp, &status); err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetVehicleStatus", "error", err)
		return nil
	}

	return &status
}

// GetVehicleStatuses fetches status for all known vehicles.
func (c *Client) GetVehicleStatuses() []*VehicleStatus {
	var statuses []*VehicleStatus
	for _, v := range c.GetVehicles() {
		if status := c.GetVehicleStatus(v.VehicleGid); status != nil {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

// GetPartners retrieves partner integrations for the customer.
func (c *Client) GetPartners(customerGid int) []Partner {
	defer c.timeTrack("GetPartners")()

	params := map[string]string{"customerGid": strconv.Itoa(customerGid)}
	resp := c.execute(apiURLs["API_PARTNERS"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetPartners")
		return nil
	}

	var partners []Partner
	if err := json.Unmarshal(resp, &partners); err != nil {
		c.log.Error("cannot unmarshal partners", "error", err)
		return nil
	}

	return partners
}

// GetRemoteConfig fetches remote configuration for the app version.
func (c *Client) GetRemoteConfig(appVersion string) map[string]interface{} {
	defer c.timeTrack("GetRemoteConfig")()

	params := map[string]string{"appVersion": appVersion}
	resp := c.execute(apiURLs["API_REMOTE_CONFIG"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetRemoteConfig")
		return nil
	}

	remoteConfig := make(map[string]interface{})
	if err := json.Unmarshal(resp, &remoteConfig); err != nil {
		c.log.Error("cannot unmarshal remote config", "error", err)
		return nil
	}

	return remoteConfig
}

// GetDeviceUsage retrieves energy usage for a single device at the specified time scale.
func (c *Client) GetDeviceUsage(id int, scope string) DeviceListUsages {
	defer c.timeTrack("GetDeviceUsage")()

	params := map[string]string{
		"deviceGids": strconv.Itoa(id),
		"instant":    time.Now().UTC().Format(time.RFC3339),
		"scale":      Durations[scope],
		"unit":       Units["kwh"],
	}
	resp := c.execute(apiURLs["API_DEVICES_USAGE"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetDeviceUsage")
		return DeviceListUsages{}
	}

	var usage DeviceListUsages
	err := json.Unmarshal(resp, &usage)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetDeviceUsage", "error", err)
	}

	return usage
}

// GetDevicesUsage retrieves energy usage for all devices at the specified time scale.
func (c *Client) GetDevicesUsage(scope string) DeviceListUsages {
	defer c.timeTrack("GetDevicesUsage")()

	params := map[string]string{
		"deviceGids": intArrayToString(c.getDeviceIDList()),
		"instant":    time.Now().UTC().Format(time.RFC3339),
		"scale":      Durations[scope],
		"unit":       Units["kwh"],
	}
	resp := c.execute(apiURLs["API_DEVICES_USAGE"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetDevicesUsage")
		return DeviceListUsages{}
	}

	var usage DeviceListUsages
	err := json.Unmarshal(resp, &usage)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetDevicesUsage", "error", err)
	}

	return usage
}

// GetChartUsage retrieves historical usage data for charting a specific channel.
func (c *Client) GetChartUsage(d *VueDevice, channel string, start, end time.Time) ChannelChart {
	defer c.timeTrack("GetChartUsage")()

	params := map[string]string{
		"deviceGid": strconv.Itoa(d.DeviceGid),
		"channel":   channel,
		"start":     start.UTC().Format(time.RFC3339),
		"end":       end.UTC().Format(time.RFC3339),
		"scale":     Durations["hour"],
		"unit":      Units["kwh"],
	}
	resp := c.execute(apiURLs["API_CHART_USAGE"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetChartUsage")
		return ChannelChart{}
	}

	var chart ChannelChart
	err := json.Unmarshal(resp, &chart)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetChartUsage", "error", err)
	}

	return chart
}

// GetDeviceProperties retrieves location and billing configuration for a device.
func (c *Client) GetDeviceProperties(id int) *Properties {
	defer c.timeTrack("GetDeviceProperty")()

	params := map[string]string{
		"deviceGid": strconv.Itoa(id),
	}
	resp := c.execute(apiURLs["API_DEVICE_PROPERTIES"], "GET", params, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetDeviceProperty")
		return nil
	}

	var properties Properties
	err := json.Unmarshal(resp, &properties)
	if err != nil {
		c.log.Error("cannot unmarshal json", "func", "GetDeviceProperty", "error", err)
	}

	return &properties
}

// GetChannelTypes returns metadata about available channel types.
func (c *Client) GetChannelTypes() []ChannelType {
	defer c.timeTrack("GetChannelTypes")()

	resp := c.execute(apiURLs["API_CHANNEL_TYPES"], "GET", map[string]string{}, false)
	if resp == nil {
		c.log.Error("request failed", "func", "GetChannelTypes")
		return nil
	}

	var channelTypes []ChannelType
	if err := json.Unmarshal(resp, &channelTypes); err != nil {
		c.log.Error("cannot unmarshal channel types", "error", err)
		return nil
	}

	return channelTypes
}

// UpdateChannel updates a channel configuration (e.g., type or multiplier).
func (c *Client) UpdateChannel(ch *Channel) error {
	if ch == nil {
		return errors.New("channel is nil")
	}
	defer c.timeTrack("UpdateChannel")()

	params := map[string]string{"deviceGid": strconv.Itoa(ch.DeviceGid)}
	resp := c.execute(apiURLs["API_CHANNELS"], "PUT", params, true)
	if resp == nil {
		return errors.New("request failed")
	}

	if err := json.Unmarshal(resp, ch); err != nil {
		c.log.Error("cannot unmarshal channel update response", "error", err)
		return err
	}

	return nil
}

// IsConnected returns true if the client has an active connection to the API.
func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// device returns the device with the given ID under the read lock.
func (c *Client) device(id int) (*VueDevice, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	d, ok := c.devices[id]
	return d, ok
}

// IsEnergyMeter checks if the device with the given ID is a Vue energy monitor
// (not an outlet, EV charger, or battery).
func (c *Client) IsEnergyMeter(id int) (bool, error) {
	if d, ok := c.device(id); ok {
		return isNil(d.Outlet) && isNil(d.EvCharger) && isNil(d.Battery), nil
	}
	return false, errors.New("device not found")
}

// IsOutlet checks if the device with the given ID is a smart outlet.
func (c *Client) IsOutlet(id int) (bool, error) {
	if d, ok := c.device(id); ok {
		return !isNil(d.Outlet), nil
	}
	return false, errors.New("device not found")
}

// IsEvCharger checks if the device with the given ID is an EV charger.
func (c *Client) IsEvCharger(id int) (bool, error) {
	if d, ok := c.device(id); ok {
		return !isNil(d.EvCharger), nil
	}
	return false, errors.New("device not found")
}

// execute performs an HTTP request against the Emporia API with retry logic and token refresh.
func (c *Client) execute(url, method string, params map[string]string, useJSON bool) []byte {
	const maxAttempts = 3
	backoff := 500 * time.Millisecond

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := c.ensureValidToken(); err != nil {
			return nil
		}

		resp, err := c.doRequest(url, method, params, useJSON)
		if err != nil {
			c.log.Error("request attempt failed", "attempt", attempt, "error", err)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		result, shouldRetry := c.handleResponse(resp, attempt, maxAttempts)
		if shouldRetry {
			time.Sleep(backoff)
			backoff *= 2
			continue
		}

		return result
	}

	c.setConnected(false)
	return nil
}

// doRequest builds and executes an HTTP request.
func (c *Client) doRequest(url, method string, params map[string]string, useJSON bool) (*resty.Response, error) {
	req := c.httpClient.R().
		SetContext(c.context()).
		SetHeader("authtoken", c.tokenID())
	c.prepareRequest(req, method, params, useJSON)

	c.logRequest(method, url, params)

	return c.executeMethod(req, method, url)
}

// prepareRequest configures the request based on HTTP method.
func (c *Client) prepareRequest(req *resty.Request, method string, params map[string]string, useJSON bool) {
	req.SetPathParams(params)
	switch method {
	case "POST":
		if useJSON {
			req.SetBody(params)
		} else {
			req.SetFormData(params)
		}
	case "PUT":
		req.SetBody(params)
	}
}

// executeMethod performs the actual HTTP call.
func (c *Client) executeMethod(req *resty.Request, method, url string) (*resty.Response, error) {
	switch method {
	case "GET":
		return req.Get(url)
	case "POST":
		return req.Post(url)
	case "PUT":
		return req.Put(url)
	default:
		return nil, errors.New("unsupported method")
	}
}

// logRequest logs request details if debug logging is enabled.
func (c *Client) logRequest(method, url string, params map[string]string) {
	if c.log != nil && c.log.Enabled(context.Background(), slog.LevelDebug) {
		c.log.Debug("emporia api request", "method", method, "url", url, "params", params)
	}
}

// handleResponse processes the response and determines if retry is needed.
// Returns the response body (or nil) and whether to retry.
func (c *Client) handleResponse(resp *resty.Response, attempt, maxAttempts int) ([]byte, bool) {
	status := resp.StatusCode()

	if c.log != nil && c.log.Enabled(context.Background(), slog.LevelDebug) {
		c.log.Debug("emporia api response", "status", status, "attempt", attempt, "body", string(resp.Bytes()))
	}

	if status == http.StatusUnauthorized {
		c.log.Debug("request unauthorized, attempting token refresh", "attempt", attempt)
		if err := c.refreshToken(); err != nil {
			c.setConnected(false)
			return nil, false
		}
		return nil, true
	}

	// Retry only transient failures (5xx and 429). Other 4xx responses are
	// client errors that won't succeed on retry, so return immediately.
	if status == http.StatusTooManyRequests || status >= http.StatusInternalServerError {
		c.log.Debug("request failed (retryable)", "status", status, "attempt", attempt)
		return nil, attempt < maxAttempts
	}

	if status >= http.StatusBadRequest {
		c.log.Debug("request failed (non-retryable)", "status", status, "attempt", attempt)
		return nil, false
	}

	c.setConnected(true)
	return resp.Bytes(), false
}

// setConnected updates the connection state under the client lock.
func (c *Client) setConnected(v bool) {
	c.mu.Lock()
	c.connected = v
	c.mu.Unlock()
}

// getDeviceIDList returns a slice of all known device GIDs.
func (c *Client) getDeviceIDList() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	var ids []int
	for _, device := range c.devices {
		ids = append(ids, device.DeviceGid)
	}

	return ids
}

// timeTrack returns a function that logs the elapsed time when called (for defer patterns).
func (c *Client) timeTrack(name string) func() {
	start := time.Now()
	return func() {
		c.log.Debug("http request ("+name+") execution time", "duration", time.Since(start))
	}
}

// sanitizeJSON tries to repair common malformed patterns (double commas, trailing commas) seen in upstream payloads.
func sanitizeJSON(in []byte) []byte {
	prev := bytes.TrimSpace(in)
	for i := 0; i < 5; i++ { // bounded passes to avoid spinning
		out := bytes.ReplaceAll(prev, []byte(",}"), []byte("}"))
		out = bytes.ReplaceAll(out, []byte(",]"), []byte("]"))
		out = bytes.ReplaceAll(out, []byte(", ,"), []byte(","))
		out = bytes.ReplaceAll(out, []byte(",,"), []byte(","))
		out = bytes.ReplaceAll(out, []byte("{,"), []byte("{"))
		if bytes.Equal(out, prev) {
			break
		}
		prev = out
	}
	return prev
}

// jsonErrorContext returns a short snippet around the syntax error offset to aid debugging.
func jsonErrorContext(data []byte, offset int64) string {
	if offset <= 0 {
		return string(data)
	}
	start := int(offset) - 40
	if start < 0 {
		start = 0
	}
	end := int(offset) + 40
	if end > len(data) {
		end = len(data)
	}
	return string(data[start:end])
}

// IsUnderMaintenance checks the maintenance sentinel and returns true when the service is reported down.
func (c *Client) IsUnderMaintenance() bool {
	resp, err := c.httpClient.R().SetHeader("authtoken", "").Get(apiURLs["API_MAINTENANCE"])
	if err != nil {
		c.log.Debug("maintenance check failed", "error", err)
		return false
	}

	// Normal operation responds with 403; maintenance returns a small JSON body.
	if resp.StatusCode() == http.StatusForbidden {
		return false
	}

	var payload map[string]string
	if err := json.Unmarshal(resp.Bytes(), &payload); err != nil {
		c.log.Debug("maintenance response not JSON", "status", resp.StatusCode())
		return false
	}

	return strings.EqualFold(payload["msg"], "down")
}
