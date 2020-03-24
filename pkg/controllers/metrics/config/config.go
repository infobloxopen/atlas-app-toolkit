package config

import (
	"github.com/prometheus/client_golang/prometheus"
	ctrl "sigs.k8s.io/controller-runtime"
)

type Configuration struct {
	BindAddress string

	Registry *prometheus.Registry
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
	if c.Registry == nil {
		c.Registry = prometheus.NewRegistry()
	}
	cc := completedConfiguration{c}
	return &CompletedConfiguration{&cc}
}

func (c *Configuration) ApplyToManagerOptions(o *ctrl.Options) {
	o.MetricsBindAddress = c.BindAddress
}
