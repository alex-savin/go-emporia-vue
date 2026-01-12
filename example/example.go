// Package main demonstrates usage of the go-emporia-vue library.
//
// This example shows how to:
//   - Configure and create a client
//   - Retrieve customer information
//   - List devices and their properties
//   - Get real-time energy usage
//   - Control outlets and EV chargers
//   - Set up scheduled polling for usage data
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	vue "github.com/alex-savin/go-emporia-vue"
	"github.com/alex-savin/go-emporia-vue/config"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Load configuration from config.yaml
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create the Emporia Vue client (uses logger from config)
	client, err := vue.NewClient(nil, cfg)
	if err != nil {
		return fmt.Errorf("failed to create client: %w", err)
	}

	logger := cfg.Logger

	// Display customer information
	printCustomerInfo(client, logger)

	// Display all devices
	printDevices(client, logger)

	// Display current energy usage
	printCurrentUsage(client)

	// Start scheduled polling if configured
	if len(cfg.Emporia.ScopeOfInterest) > 0 {
		runScheduler(ctx, client, cfg, logger)
	}

	return nil
}

// printCustomerInfo displays the authenticated customer's information.
func printCustomerInfo(client *vue.Client, logger *slog.Logger) {
	customer := client.GetCustomer()
	if customer == nil {
		logger.Warn("could not retrieve customer info")
		return
	}

	fmt.Println("\n=== Customer Information ===")
	fmt.Printf("  Customer ID: %d\n", customer.CustomerId)
	fmt.Printf("  Name: %s %s\n", customer.FirstName, customer.LastName)
	fmt.Printf("  Email: %s\n", customer.Email)
	fmt.Printf("  Created: %s\n", customer.CreatedAt.Format(time.RFC3339))
}

// printDevices displays all devices and their details.
func printDevices(client *vue.Client, logger *slog.Logger) {
	devices := client.GetDevices()
	if len(devices) == 0 {
		logger.Warn("no devices found")
		return
	}

	fmt.Println("\n=== Devices ===")
	for i, device := range devices {
		info := device.Info()
		fmt.Printf("\n[%d] %s (GID: %d)\n", i+1, device.Type(), device.DeviceGid)
		fmt.Printf("    Manufacturer: %s\n", info.Manufacturer)
		fmt.Printf("    Model: %s\n", info.Model)
		fmt.Printf("    Firmware: %s\n", info.Firmware)

		// Show outlet-specific info
		if device.IsOutlet() {
			status := "OFF"
			if device.Outlet.OutletOn {
				status = "ON"
			}
			fmt.Printf("    Outlet Status: %s\n", status)
		}

		// Show EV charger-specific info
		if device.IsEvCharger() {
			status := "OFF"
			if device.EvCharger.ChargerOn {
				status = "ON"
			}
			fmt.Printf("    Charger Status: %s\n", status)
			fmt.Printf("    Charging Rate: %d/%d A\n", device.EvCharger.ChargingRate, device.EvCharger.MaxChargingRate)
		}

		// Show battery-specific info
		if device.IsBattery() {
			fmt.Printf("    Battery: present\n")
		}

		// Show channel count
		if len(device.Channels) > 0 {
			fmt.Printf("    Channels: %d\n", len(device.Channels))
		}
	}

	// Summary by device type
	fmt.Println("\n=== Device Summary ===")
	fmt.Printf("  Energy Monitors: %d\n", len(client.GetEnergyMonitors()))
	fmt.Printf("  Outlets: %d\n", len(client.GetOutlets()))
	fmt.Printf("  EV Chargers: %d\n", len(client.GetEvChargers()))
	fmt.Printf("  Batteries: %d\n", len(client.GetBatteries()))

	// Show vehicles if any
	vehicles := client.GetVehicles()
	if len(vehicles) > 0 {
		fmt.Printf("  Vehicles: %d\n", len(vehicles))
		for _, v := range vehicles {
			fmt.Printf("    - %s %s %s (%s)\n", v.Year, v.Make, v.Model, v.DisplayName)
		}
	}
}

// printCurrentUsage displays current energy usage for all devices.
func printCurrentUsage(client *vue.Client) {
	devices := client.GetDevices()
	if len(devices) == 0 {
		return
	}

	fmt.Println("\n=== Current Energy Usage (1 minute) ===")
	for _, device := range devices {
		channelUsages := device.Usage("minute")
		if len(channelUsages) == 0 {
			continue
		}

		fmt.Printf("\nDevice: %d (%s)\n", device.DeviceGid, device.Type())
		for _, ch := range channelUsages {
			if ch.Usage != 0 {
				fmt.Printf("  %s: %.2f kWh (%.1f%%)\n", ch.Name, ch.Usage, ch.Percentage)
			}
		}
	}
}

// runScheduler starts the scheduled polling for usage data using standard library tickers.
func runScheduler(ctx context.Context, client *vue.Client, cfg *vue.Config, logger *slog.Logger) {
	// Create tickers for each scope of interest
	for _, scale := range cfg.Emporia.ScopeOfInterest {
		interval := getInterval(scale)
		if interval == 0 {
			logger.Warn("unknown scale, skipping", "scale", scale)
			continue
		}

		// Start a goroutine for each polling interval
		go func(scale string, interval time.Duration) {
			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			logger.Info("started polling", "scale", scale, "interval", interval)

			// Poll immediately on start
			pollUsage(client, logger, scale)

			for {
				select {
				case <-ctx.Done():
					logger.Info("stopping poller", "scale", scale)
					return
				case <-ticker.C:
					pollUsage(client, logger, scale)
				}
			}
		}(scale, interval)
	}

	logger.Info("scheduler started, press Ctrl+C to exit")

	// Wait for shutdown signal
	<-ctx.Done()
	logger.Info("shutting down...")
}

// getInterval returns the polling interval for a given scale.
func getInterval(scale string) time.Duration {
	switch scale {
	case "minute":
		return 1 * time.Minute
	case "hour":
		return 1 * time.Hour
	case "day":
		return 24 * time.Hour
	default:
		return 0
	}
}

// pollUsage fetches and logs usage data.
func pollUsage(client *vue.Client, logger *slog.Logger, scale string) {
	usages := client.GetDevicesUsage(scale)
	logger.Info("polled usage data",
		"scale", scale,
		"devices", len(usages.DeviceListUsages.Devices),
		"timestamp", time.Now().Format(time.RFC3339),
	)
}
