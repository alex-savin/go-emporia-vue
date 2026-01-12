# Go Emporia Vue

[![CI](https://github.com/alex-savin/go-emporia-vue/actions/workflows/ci.yaml/badge.svg)](https://github.com/alex-savin/go-emporia-vue/actions/workflows/ci.yaml)
[![Go Reference](https://pkg.go.dev/badge/github.com/alex-savin/go-emporia-vue.svg)](https://pkg.go.dev/github.com/alex-savin/go-emporia-vue)
[![Go Report Card](https://goreportcard.com/badge/github.com/alex-savin/go-emporia-vue)](https://goreportcard.com/report/github.com/alex-savin/go-emporia-vue)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go client for the Emporia Vue API. Wraps device, outlet, EV charger, vehicle, battery, and thermostat endpoints with retry-aware HTTP calls and token management via AWS Cognito.

## Features

- **Device Management**: Fetch and control Vue energy monitors, smart outlets, EV chargers, batteries, and thermostats
- **Usage Tracking**: Retrieve real-time and historical energy usage data at various time scales
- **Vehicle Integration**: Monitor connected EV charging status and battery levels
- **Automatic Authentication**: AWS Cognito token management with automatic refresh
- **Retry Logic**: Built-in exponential backoff for resilient API calls
- **Structured Logging**: Uses Go's `log/slog` for configurable logging output

## Requirements

- Go 1.25 or later

## Install

```bash
go get github.com/alex-savin/go-emporia-vue
```

## Quickstart

1. Create `config.yaml` (or copy from `example/config.sample.yaml`) with your Emporia credentials.
2. Instantiate the client and fetch devices/usage:

```go
package main

import (
    "log/slog"
    "os"

    vue "github.com/alex-savin/go-emporia-vue"
)

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
    
    cfg, err := vue.Conf()
    if err != nil {
        logger.Error("failed to load config", "error", err)
        os.Exit(1)
    }

    client, err := vue.NewClient(logger, cfg)
    if err != nil {
        logger.Error("failed to create client", "error", err)
        os.Exit(1)
    }

    // Get customer info
    customer := client.GetCustomer()
    logger.Info("authenticated", "email", customer.Email)

    // List all devices
    devices := client.GetDevices()
    for _, d := range devices {
        logger.Info("device", "gid", d.DeviceGid, "type", d.Type())
    }

    // Get real-time usage
    usage := client.GetDevicesUsage("minute")
    logger.Info("usage", "devices", len(usage.DeviceListUsages.Devices))
}
```

Tokens are cached to `.credentials.yaml` with `0600` permissions and automatically refreshed.

## Configuration

Create a `config.yaml` file:

```yaml
emporia:
  credentials:
    username: "your-email@example.com"
    password: "your-password"
  
  # Optional: AWS Cognito settings (defaults are provided)
  # cognito:
  #   client_id: "4qte47jbstod8apnfic0bunmrq"
  #   region: "us-east-2"
  #   user_pool: "us-east-2_ghlOXVLi1"

  auto_reconnect: true
  
  # Polling scales: minute, hour, day, month
  scope_of_interest:
    - minute
    - day

timezone: "America/New_York"

logging:
  level: "INFO"      # DEBUG, INFO, WARN, ERROR
  output: "TEXT"     # TEXT or JSON
  source: false      # Include source file in logs
```

## API Reference

### Client Methods

| Method | Description |
|--------|-------------|
| `GetCustomer()` | Get authenticated customer info |
| `GetDevices()` | Fetch all devices for the account |
| `GetEnergyMonitors()` | List energy monitor devices |
| `GetOutlets()` | List all smart outlets |
| `GetEvChargers()` | List all EV chargers |
| `GetBatteries()` | List battery storage devices |
| `GetThermostats()` | List thermostat devices |
| `GetVehicles()` | List connected vehicles |
| `GetVehicleStatus(gid)` | Get status for a specific vehicle |
| `GetVehicleStatuses()` | Get charging status for all vehicles |
| `GetDevicesUsage(scope)` | Get usage for all devices |
| `GetDeviceUsage(id, scope)` | Get usage for a specific device |
| `GetChartUsage(device, channel, start, end)` | Get historical usage data |
| `GetDeviceProperties(id)` | Get device configuration |
| `GetChannelTypes()` | List available channel types |
| `UpdateChannel(channel)` | Update channel configuration |

### Device Methods

| Method | Description |
|--------|-------------|
| `device.Info()` | Get device info (manufacturer, model, firmware) |
| `device.Type()` | Returns "outlet", "evcharger", or "energymonitor" |
| `device.IsOutlet()` | Check if device is a smart outlet |
| `device.IsEvCharger()` | Check if device is an EV charger |
| `device.IsBattery()` | Check if device has battery storage |
| `device.IsEnergyMeter()` | Check if device is an energy monitor |
| `device.Usage(scope)` | Get channel usage for this device |
| `device.On()` | Turn on outlet or start EV charging |
| `device.Off()` | Turn off outlet or stop EV charging |

### Usage Scales

| Scale | Description |
|-------|-------------|
| `minute` | Last minute of usage |
| `hour` | Last hour of usage |
| `day` | Current day usage |
| `month` | Current month usage |

## Example

See [`example/example.go`](example/example.go) for a complete runnable example that:
- Displays customer information
- Lists all devices with their properties
- Shows current energy usage
- Runs scheduled polling for usage data

```bash
cd example
cp config.sample.yaml config.yaml
# Edit config.yaml with your credentials
go run example.go
```

## Tests

```bash
go test ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

Please ensure:
- All tests pass (`go test ./...`)
- Code is formatted (`go fmt ./...`)
- No lint errors (`golangci-lint run`)

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
