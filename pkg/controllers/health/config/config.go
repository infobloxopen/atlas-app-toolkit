package config

import (
	ctrl "sigs.k8s.io/controller-runtime"
)

type Configuration struct {
	BindAddress string
	Healthz     string
	Readyz      string
}

type completedConfiguration struct {
	*Configuration
}

// CompletedConfiguration same as Configuration, just to swap private object.
type CompletedConfiguration struct {
	// Embed a private pointer that cannot be instantiated outside of this package.
	*completedConfiguration
}

// Complete fills in any fields not set that are required to have valid data. It's mutating the receiver.
func (c *Configuration) Complete() *CompletedConfiguration {
	cc := completedConfiguration{c}
	return &CompletedConfiguration{&cc}
}

func (c *Configuration) ApplyToManagerOptions(o *ctrl.Options) {
	o.HealthProbeBindAddress = c.BindAddress
	o.LivenessEndpointName = c.Healthz
	o.ReadinessEndpointName = c.Readyz
}
