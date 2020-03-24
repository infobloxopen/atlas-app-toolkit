package manager

import (
	"github.com/go-logr/logr"
	"github.com/infobloxopen/atlas-app-toolkit/pkg/controllers/manager/config"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
)

func NewControllerRuntimeOptions(scheme *runtime.Scheme, config *config.Configuration) *ctrl.Options {

	o := &ctrl.Options{
		Scheme: scheme,
	}
	if config != nil {
		config.ApplyToManagerOptions(o)
	}
	if config.Health != nil {
		config.Health.ApplyToManagerOptions(o)
	}
	if config.Metrics != nil {
		config.Metrics.ApplyToManagerOptions(o)
	}
	return o
}

func NewManager(o *ctrl.Options, kubeConfig *rest.Config, logger logr.Logger) (ctrl.Manager, error) {
	// Set the logger used by controller-runtime
	ctrl.SetLogger(logger)

	// Initialize the controller-runtime.Manager
	mgr, err := ctrl.NewManager(kubeConfig, *o)
	if err != nil {
		ctrl.Log.WithName("kubecontroller").Error(err, "unable to create manager")
		return nil, err
	}

	return mgr, nil
}
