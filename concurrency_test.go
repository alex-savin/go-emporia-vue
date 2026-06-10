package vue

import (
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"
)

func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestClientConcurrentAccess exercises the lock-guarded fields (devices,
// connected, token) from many goroutines at once. Run with -race it guards the
// data races that previously existed around c.devices/c.connected/c.token.
func TestClientConcurrentAccess(t *testing.T) {
	c := &Client{
		devices:     map[int]*VueDevice{1: {DeviceGid: 1}},
		credentials: &credentials{token: &token{Expiration: time.Now().Add(time.Hour).Unix(), ID: "tok"}},
		ctx:         nil,
		log:         quietLogger(),
	}

	const iterations = 2000
	var wg sync.WaitGroup

	// Writer: simulate GetDevices reassigning the map under the lock.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < iterations; i++ {
			c.mu.Lock()
			c.devices = map[int]*VueDevice{i: {DeviceGid: i}}
			c.mu.Unlock()
		}
	}()

	// Readers across all the guarded accessors.
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < iterations; i++ {
				_ = c.snapshotDevices()
				_ = c.getDeviceIDList()
				_, _ = c.device(1)
				_ = c.IsConnected()
				_ = c.tokenID()
				c.setConnected(i%2 == 0)
			}
		}()
	}

	wg.Wait()
}

func TestCredentialsFilePathHonorsEnv(t *testing.T) {
	t.Setenv("EMPORIA_CREDENTIALS_FILE", "/tmp/custom-creds.yaml")
	if got := credentialsFilePath(); got != "/tmp/custom-creds.yaml" {
		t.Errorf("credentialsFilePath() = %q, want /tmp/custom-creds.yaml", got)
	}

	t.Setenv("EMPORIA_CREDENTIALS_FILE", "")
	if got := credentialsFilePath(); got != "./.credentials.yaml" {
		t.Errorf("credentialsFilePath() default = %q, want ./.credentials.yaml", got)
	}
}
