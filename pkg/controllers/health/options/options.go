package options

import (
	"github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/health/config"
	flagnamer "github.com/infobloxopen/atlas-app-toolkit/pkg/flags/namer"
	"github.com/spf13/pflag"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type Options struct {
	flagnamer.Namer

	BindAddress string
	Healthz     string
	Readyz      string
}

func NewOptions() (*Options, error) {
	return NewOptionsWithGroupName("health")
}

func NewOptionsWithGroupName(groupName string) (*Options, error) {
	o := &Options{
		Namer:       flagnamer.NewNamer(groupName),
		BindAddress: ":10254",
		Healthz:     "/healthz",
		Readyz:      "/readyz",
	}
	return o, nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BindAddress, o.FlagNamer("bind-address"), o.BindAddress, "Address to use for binding monitoring server listener. Exposes healthz and readyz endpoints.")
	fs.StringVar(&o.Healthz, o.FlagNamer("healthz-endpoint"), o.Healthz, "URL endpoint to expose health information.")
	fs.StringVar(&o.Readyz, o.FlagNamer("readiness-endpoint"), o.Readyz, "URL endpoint to expose readiness information.")

}

// ApplyTo fills up Debugging config with options.
func (o *Options) ApplyTo(c *config.Configuration) error {
	if o == nil {
		return nil
	}
	c.BindAddress = o.BindAddress
	c.Healthz = o.Healthz
	c.Readyz = o.Readyz

	return nil
}

// Validate checks validation of DebuggingOptions.
func (o *Options) Validate() error {
	if o == nil {
		return nil
	}

	errs := []error{}
	return utilerrors.NewAggregate(errs)
}

// Config return a controller config objective
func (o *Options) Config() (*config.Configuration, error) {
	var err error
	if err = o.Validate(); err != nil {
		return nil, err
	}

	c := &config.Configuration{}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}

	return c, nil
}
