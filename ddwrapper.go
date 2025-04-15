package ddwrapper

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Config defines the global configurations for the library
type Config struct {
	ServiceName string // Service name for Datadog
	Environment string // Environment (e.g., prod, staging)
	DDHost      string // Datadog agent host (optional)
	Enabled     bool   // Enables/disables tracing
}

// DDWrapper is the main structure
type DDWrapper struct {
	config Config
}

// New creates a new instance of DDWrapper
func New(config Config) *DDWrapper {
	if config.Enabled {
		// Initializes the Datadog tracer
		tracer.Start(
			tracer.WithService(config.ServiceName),
			tracer.WithEnv(config.Environment),
			tracer.WithAgentAddr(config.DDHost),
		)
	}
	return &DDWrapper{config: config}
}

// Stop stops the tracer (should be called when the application is shutting down)
func (w *DDWrapper) Stop() {
	if w.config.Enabled {
		tracer.Stop()
	}
}

// Integration methods will be added below
