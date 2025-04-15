package internal

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// ConfigureTracer initializes the tracer with default options
func ConfigureTracer(serviceName, env, agentAddr string) {
	tracer.Start(
		tracer.WithService(serviceName),
		tracer.WithEnv(env),
		tracer.WithAgentAddr(agentAddr),
		tracer.WithGlobalTag("version", "1.0.0"), // Example of a global tag
	)
}
