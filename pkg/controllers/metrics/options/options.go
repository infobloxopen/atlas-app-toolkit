package options

import (
	"github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/metrics/config"
	flagnamer "github.com/infobloxopen/atlas-app-toolkit/pkg/flags/namer"
	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
)

type Options struct {
	flagnamer.Namer
	BindAddress string
}

func NewOptions() (*Options, error) {
	return NewOptionsWithGroupName("metrics")
}

func NewOptionsWithGroupName(groupName string) (*Options, error) {
	o := &Options{
		Namer:       flagnamer.NewNamer(groupName),
		BindAddress: ":9000",
	}
	return o, nil
}

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.BindAddress, o.FlagNamer("bind-address"), o.BindAddress, "Address to use for binding metrics server listener.")
}

// ApplyTo fills up Debugging config with options.
func (o *Options) ApplyTo(c *config.Configuration) error {
	if o == nil {
		return nil
	}
	c.BindAddress = o.BindAddress

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
