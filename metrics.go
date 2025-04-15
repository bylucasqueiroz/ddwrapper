package ddwrapper

import (
	"context"
	"fmt"

	dogstatsd "github.com/DataDog/datadog-go/v5/statsd"
)

// CustomMetric defines a custom metric
type CustomMetric struct {
	Name  string
	Tags  []string
	Value float64
	Type  string // "gauge", "count", "histogram"
}

// RecordCustomMetric sends a custom metric
func (w *DDWrapper) RecordCustomMetric(ctx context.Context, metric CustomMetric) error {
	if !w.config.Enabled {
		return nil
	}

	// Creates a DogStatsD client
	statsd, err := dogstatsd.New("127.0.0.1:8125")
	if err != nil {
		return err
	}
	defer statsd.Close()

	// Adds global tags (e.g., environment)
	tags := append(metric.Tags, "env:"+w.config.Environment)

	switch metric.Type {
	case "gauge":
		return statsd.Gauge(metric.Name, metric.Value, tags, 1.0)
	case "count":
		return statsd.Count(metric.Name, int64(metric.Value), tags, 1.0)
	case "histogram":
		return statsd.Histogram(metric.Name, metric.Value, tags, 1.0)
	default:
		return fmt.Errorf("metric type %s not supported", metric.Type)
	}
}
