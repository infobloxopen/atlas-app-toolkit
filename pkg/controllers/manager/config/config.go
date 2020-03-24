package config

import (
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"

	debugconfig "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/debug/config"
	healthconfig "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/health/config"
	metricsconfig "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/metrics/config"
)

type Configuration struct {
	Debug   *debugconfig.Configuration
	Health  *healthconfig.Configuration
	Metrics *metricsconfig.Configuration

	LeaderElection bool
	WebhookEnabled bool
	WebhookHost    string
	WebhookPort    int

	MasterURL  string
	KubeConfig *rest.Config
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
	o.LeaderElection = c.LeaderElection
	o.Host = c.WebhookHost
	o.Port = c.WebhookPort
}
