package tracing

import (
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"go.opencensus.io/trace"
)

//Exporter creates a new OC Agent exporter and configure for tracing
func Exporter(address string, serviceName string, sampler trace.Sampler) error {
	// TRACE: Setup OC agent for tracing
	exporter, err := ocagent.NewExporter(
		ocagent.WithInsecure(),
		ocagent.WithReconnectionPeriod(5*time.Second),
		ocagent.WithAddress(address),
		ocagent.WithServiceName(serviceName))
	if err != nil {
		return err
	}

	trace.RegisterExporter(exporter)
	trace.ApplyConfig(trace.Config{DefaultSampler: sampler})

	return nil
}

//SamplerForFraction init sampler for specified fraction
func SamplerForFraction(fraction float64) trace.Sampler {
	return trace.ProbabilitySampler(fraction)
}
