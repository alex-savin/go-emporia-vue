package vue

import (
	"time"

	"github.com/alex-savin/go-emporia-vue/config"
)

// NoOpMetricsRecorder provides a no-op implementation of config.MetricsRecorder
type NoOpMetricsRecorder struct{}

func (n *NoOpMetricsRecorder) RecordRequest(method, endpoint string, duration time.Duration, success bool) {
}
func (n *NoOpMetricsRecorder) RecordError(errorType string)             {}
func (n *NoOpMetricsRecorder) RecordRetry(endpoint string, attempt int) {}

// MetricsConfig holds metrics configuration
type MetricsConfig struct {
	Recorder config.MetricsRecorder
}

// DefaultMetricsConfig returns a default metrics configuration with no-op recorder
func DefaultMetricsConfig() *MetricsConfig {
	return &MetricsConfig{
		Recorder: &NoOpMetricsRecorder{},
	}
}
