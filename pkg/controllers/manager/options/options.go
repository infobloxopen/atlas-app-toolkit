package options

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	ctrl "sigs.k8s.io/controller-runtime"

	flagnamer "github.com/infobloxopen/atlas-app-toolkit/pkg/flags/namer"

	debugoptions "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/debug/options"
	healthoptions "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/health/options"
	metricsoptions "github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/metrics/options"

	"github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/manager/config"
)

// Options contains the fields which map to
// controller-runtime/manager.Options that is embedded in config.Configuration
type Options struct {
	flagnamer.Namer

	// Debug options
	Debug *debugoptions.Options

	// Health options
	Health *healthoptions.Options

	// Metrics options
	Metrics *metricsoptions.Options

	// LeaderElection determines whether or not to use leader election when
	// starting the manager.
	LeaderElection bool

	// LeaderElectionNamespace determines the namespace in which the leader
	// election configmap will be created.
	LeaderElectionNamespace string

	// LeaderElectionID determines the name of the configmap that leader election
	// will use for holding the leader lock.
	LeaderElectionID string

	// LeaseDuration is the duration that non-leader candidates will
	// wait to force acquire leadership. This is measured against time of
	// last observed ack. Default is 15 seconds.
	LeaseDuration time.Duration
	// RenewDeadline is the duration that the acting master will retry
	// refreshing leadership before giving up. Default is 10 seconds.
	RenewDeadline time.Duration
	// RetryPeriod is the duration the LeaderElector clients should wait
	// between tries of actions. Default is 2 seconds.
	RetryPeriod time.Duration

	// Namespace if specified restricts the manager's cache to watch objects in
	// the desired namespace Defaults to all namespaces
	//
	// Note: If a namespace is specified, controllers can still Watch for a
	// cluster-scoped resource (e.g Node).  For namespaced resources the cache
	// will only hold objects from the desired namespace.
	Namespace string

	WebhookEnabled     bool
	WebhookBindAddress string

	// these are indirectly set by WebhookBindAddress during validation
	webhookHost string
	webhookPort int

	// CertDir is the directory that contains the server key and certificate.
	// if not set, webhook server would look up the server key and certificate in
	// {TempDir}/k8s-webhook-server/serving-certs. The server key and certificate
	// must be named tls.key and tls.crt, respectively.
	CertDir string

	KubeConfig *genericclioptions.ConfigFlags
}

func NewOptions() (*Options, error) {
	return NewOptionsWithGroupName("")
}

func NewOptionsWithGroupName(groupName string) (*Options, error) {
	debugOptions, err := debugoptions.NewOptionsWithGroupName(groupName)
	if err != nil {
		return nil, err
	}
	healthOptions, err := healthoptions.NewOptionsWithGroupName(groupName)
	if err != nil {
		return nil, err
	}
	metricsOptions, err := metricsoptions.NewOptionsWithGroupName(groupName)
	if err != nil {
		return nil, err
	}

	o := &Options{
		Namer:   flagnamer.NewNamer(groupName),
		Debug:   debugOptions,
		Health:  healthOptions,
		Metrics: metricsOptions,

		LeaderElection: false,
		LeaseDuration:  time.Second * 15,
		RenewDeadline:  time.Second * 10,
		RetryPeriod:    time.Second * 2,

		WebhookEnabled: true,
		webhookPort:    9443,
	}
	kubeConfigFlags := genericclioptions.NewConfigFlags(true).WithDeprecatedPasswordFlag()
	o.KubeConfig = kubeConfigFlags
	return o, nil
}

var certDirectoryHelp = `
Directory that contains the server key and certificate.
if not set, webhook server would look up the server key and certificate in 
{TempDir}/k8s-webhook-server/serving-certs. The server key and certificate 
must be named tls.key and tls.crt, respectively
`

func (o *Options) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&o.CertDir, o.FlagNamer("certificate-directory"), o.CertDir, certDirectoryHelp)

	fs.BoolVar(&o.LeaderElection, o.FlagNamer("leader-election"), o.LeaderElection, "Enable leader-election for app-def-controllers")
	fs.StringVar(&o.LeaderElectionNamespace, o.FlagNamer("leader-election-namespace"), o.LeaderElectionNamespace, "Namespace where leader election configmap is created.")
	fs.StringVar(&o.LeaderElectionID, o.FlagNamer("leader-election-id"), o.LeaderElectionID, "Name of the configmap used for leader election lock.")
	fs.DurationVar(&o.LeaseDuration, o.FlagNamer("leader-election-lease-duration"), o.LeaseDuration, "Duration that non-leaders will wait to force acquire leadership.")
	fs.DurationVar(&o.RenewDeadline, o.FlagNamer("leader-election-renew-deadline"), o.RenewDeadline, "Duration that the acting master will retry refreshing leadership before giving up.")
	fs.DurationVar(&o.RetryPeriod, o.FlagNamer("leader-election-retry-period"), o.RetryPeriod, "Duration that leader-election cliets will wait between actions.")

	fs.StringVar(&o.Namespace, o.FlagNamer("namespace"), o.Namespace, "Namespace to limit resource watch caching.")

	fs.BoolVar(&o.WebhookEnabled, o.FlagNamer("webhook-enabled"), o.WebhookEnabled, "Enable webhook for app-def-controllers")
	fs.StringVar(&o.WebhookBindAddress, o.FlagNamer("webhook-bind-address"), o.WebhookBindAddress, "Address to use for binding webhook server listener.")

	o.KubeConfig.AddFlags(fs)
	if o.Debug != nil {
		o.Debug.AddFlags(fs)
	}
	if o.Health != nil {
		o.Health.AddFlags(fs)
	}
	if o.Metrics != nil {
		o.Metrics.AddFlags(fs)
	}
}

// Config return a controller config objective
func (o *Options) Config() (*config.Configuration, error) {
	if err := o.Validate(); err != nil {
		return nil, err
	}
	c := &config.Configuration{}
	incluster, ierr := ctrl.GetConfig()
	if ierr == nil {
		c.KubeConfig = incluster
	} else {
		clikubeConfig, err := o.KubeConfig.ToRESTConfig()
		if err != nil {
			return nil, err
		}
		c.KubeConfig = clikubeConfig
	}
	if err := o.ApplyTo(c); err != nil {
		return nil, err
	}
	if o.Debug != nil {
		debuggingConfig, err := o.Debug.Config()
		if err != nil {
			return nil, err
		}
		c.Debug = debuggingConfig
	}
	if o.Health != nil {
		healthConfig, err := o.Health.Config()
		if err != nil {
			return nil, err
		}
		c.Health = healthConfig
	}
	if o.Metrics != nil {
		metricsConfig, err := o.Metrics.Config()
		if err != nil {
			return nil, err
		}
		c.Metrics = metricsConfig
	}
	return c, nil
}

// ApplyTo fills up controller manager config with options.
func (o *Options) ApplyTo(c *config.Configuration) error {
	var err error

	c.LeaderElection = o.LeaderElection
	c.WebhookEnabled = o.WebhookEnabled
	c.WebhookHost = o.webhookHost
	c.WebhookPort = o.webhookPort

	return err
}

// Validate is used to validate the options and config before launching the controller
func (o *Options) Validate() error {
	var errs []error

	webhookErr := fmt.Errorf("invalid flag --webhook-bind-address: %s", o.WebhookBindAddress)
	parts := strings.Split(o.WebhookBindAddress, ":")
	if len(parts) > 2 {
		errs = append(errs, webhookErr)
	} else {
		_, err := strconv.Atoi(parts[0])
		if err == nil {
			errs = append(errs, webhookErr)
		}
		o.webhookHost = parts[0]
		if len(parts) == 2 {
			port, err := strconv.Atoi(parts[1])
			if err != nil {
				errs = append(errs, webhookErr)
			}
			o.webhookPort = port
		}
	}
	if o.Debug != nil {
		errs = append(errs, o.Debug.Validate())
	}
	if o.Health != nil {
		errs = append(errs, o.Health.Validate())
	}
	if o.Metrics != nil {
		errs = append(errs, o.Metrics.Validate())
	}

	return utilerrors.NewAggregate(errs)
}
