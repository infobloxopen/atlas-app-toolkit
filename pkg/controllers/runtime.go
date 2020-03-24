package controllers

import (
	"errors"
	"time"

	"github.com/go-logr/logr"
	"github.com/infobloxopen/atlas-app-toolkit/pkg/signals"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
)

const (
	// Interval used when starting controllers
	DefaultStartInterval = 1 * time.Second
	// Jitter used when starting controller managers
	DefaultStartJitterMax = 1.0
)

// RuntimeContext provides a common way to register  controllers/sub-processes, provide ordered startup,
// allow enabling / disabling controllers via cmdline argument, and standardizing signal handling / exit notification
type RuntimeContext interface {
	signals.ExitSignals

	RegisterControllerInit(controllerName string, initFunc InitFunc, defaultEnabled bool) (registered bool)
	RegisterControllerOrderedInit(controllerName string, initFunc InitFunc, defaultEnabled bool) (registered bool)
	Logger() logr.Logger
	KnownControllers() []string
	KnownControllersDisabledByDefault() []string
	StartControllers(customCtx interface{}, controllerFlags []string) error
}

type runtimeContext struct {
	logger        logr.Logger
	runtimeLogger logr.Logger

	//controllers               []string
	controllersOrderedStart   []string
	controllersInitMap        map[string]InitFunc
	controllersDefaultEnabled sets.String
	controllersStarted        sets.String

	StartInterval  time.Duration // minimum time between controller start
	StartJitterMax float64       // maximum additional jitter delay between controller starts

	// ExitSignals contains various signaling methods that apps can use for receiving/signaling shutdown
	signals.ExitSignals
}

// NewRuntimeContext creates a default runtime context
// if a nil value is provided for the signals.ExitSignals interface, a new handler will be
// registered using pkg/signals (note that creating two handlers with this package will panic)
func NewRuntimeContext(logger logr.Logger, exitSignals signals.ExitSignals) RuntimeContext {
	if exitSignals == nil {
		exitSignals = signals.SetupExitHandlers(logger)
	}
	return &runtimeContext{
		ExitSignals:               exitSignals,
		logger:                    logger,
		runtimeLogger:             logger.WithName("runtime-context"),
		controllersInitMap:        make(map[string]InitFunc),
		controllersDefaultEnabled: sets.NewString(),
		controllersStarted:        sets.NewString(),
		controllersOrderedStart:   make([]string, 0, 5),
		StartInterval:             DefaultStartInterval,
		StartJitterMax:            DefaultStartJitterMax,
	}
}

// Logger returns the embedded log-r logger
func (c *runtimeContext) Logger() logr.Logger {
	return c.logger
}

// RegisterControllerInit adds a controller (by name and InitFunc) to the RuntimeContext controllers
// controllers registered with this function are started after those registered with Ordered Init
// ordering of these controllers is not deterministic
func (c *runtimeContext) RegisterControllerInit(controllerName string, initFunc InitFunc, defaultEnabled bool) (registered bool) {
	if _, exists := c.controllersInitMap[controllerName]; exists {
		c.runtimeLogger.V(1).Info("%s has already been registered, this one will be ignored", controllerName)
		return false
	}
	c.controllersInitMap[controllerName] = initFunc
	if defaultEnabled {
		c.controllersDefaultEnabled.Insert(controllerName)
	}
	c.runtimeLogger.V(1).Info("controller initFunc registered", "name", controllerName, "defaultEnabled", defaultEnabled)
	return true
}

// RegisterControllerOrderedInit adds a controller (by name and InitFunc) to the RuntimeContext controllers and ordered start list
// startup order is based on the registration order of controllers
// all ordered init controllers are started before regular controllers
func (c *runtimeContext) RegisterControllerOrderedInit(controllerName string, initFunc InitFunc, defaultEnabled bool) (registered bool) {
	registered = c.RegisterControllerInit(controllerName, initFunc, defaultEnabled)
	if registered {
		c.controllersOrderedStart = append(c.controllersOrderedStart, controllerName)
		c.runtimeLogger.V(1).Info("controller added to ordered start list", "name", controllerName)
	}
	return registered
}

// IsControllerEnabled evaluates whether the controller is enabled/disabled by cmdline flags relative to default-enabled controllers
func (c *runtimeContext) IsControllerEnabled(name string, controllerFlags []string) bool {
	hasStar := false
	for _, ctrl := range controllerFlags {
		if ctrl == name {
			return true
		}
		if ctrl == "-"+name {
			return false
		}
		if ctrl == "*" {
			hasStar = true
		}
	}
	// if we get here, there was no explicit choice
	if !hasStar {
		// nothing on by default
		return false
	}
	return c.EnabledByDefault(name)
}

// EnabledByDefault returns a boolean indicating whether controller is enabled by default
func (c *runtimeContext) EnabledByDefault(name string) bool {
	return c.controllersDefaultEnabled.Has(name)
}

// InitFunc is used to launch a particular controller.  It may run additional "should I activate checks".
// Any error returned will cause the controller process to `Fatal`
// ctx will need to be cast to the parent type which embeds the RuntimeContext struct
// The bool indicates whether the controller was enabled.
type InitFunc func(ctx interface{}) (bool, error)

// KnownControllers returns a list of controller names registered with the RuntimeContext
func (c *runtimeContext) KnownControllers() []string {
	ret := sets.StringKeySet(c.ControllerInitializers())
	return ret.List()
}

// KnownControllersDisabledByDefault returns a list of controllers which were registered with RuntimeContext as disabled by default
func (c *runtimeContext) KnownControllersDisabledByDefault() []string {
	ret := sets.StringKeySet(c.ControllerInitializers())
	return ret.Difference(c.controllersDefaultEnabled).List()
}

// var ControllersDisabledByDefault = sets.NewString()

// NewControllerInitializers is a public map of named controller groups (you can start more than one in an init func)
// paired to their InitFunc.  This allows for structured downstream composition and subdivision.
func (c *runtimeContext) ControllerInitializers() map[string]InitFunc {
	return c.controllersInitMap
}

func (c *runtimeContext) OrderedControllerStart() []string {
	return c.controllersOrderedStart
}

func (c *runtimeContext) startController(customContext interface{}, controllerName string, controllerFlags []string) error {
	if !c.IsControllerEnabled(controllerName, controllerFlags) {
		c.runtimeLogger.V(1).Info("controller is disabled", "name", controllerName)
		return nil
	}
	if c.controllersStarted.Has(controllerName) {
		// skip already running controllers
		c.runtimeLogger.V(1).Info("controller was already started", "name", controllerName)
		return nil
	}
	time.Sleep(wait.Jitter(c.StartInterval, c.StartJitterMax))

	initFunc, exists := c.controllersInitMap[controllerName]
	if !exists {
		return errors.New("fatal start error: initFunc not registered")
	}
	c.runtimeLogger.Info("starting controller", "name", controllerName)
	started, err := initFunc(customContext)
	if err != nil {
		c.runtimeLogger.Error(err, "error during controller start", "name", controllerName)
		return err
	}
	if !started {
		c.runtimeLogger.Info("controller did not start", "name", controllerName)
		return nil
	}
	c.controllersStarted.Insert(controllerName)
	c.runtimeLogger.Info("controller started", "name", controllerName)
	return nil
}

func (c *runtimeContext) StartControllers(customCtx interface{}, controllerFlags []string) error {
	var err error
	c.runtimeLogger.Info("controllers flag", "controllers", controllerFlags)
	if len(c.KnownControllers()) == 0 {
		return errors.New("no controllers are registered")
	}

	// start controllers which are order-dependent (marking each as started)
	c.runtimeLogger.Info("starting ordered controllers")
	for i, controllerName := range c.OrderedControllerStart() {
		c.runtimeLogger.Info("ordered start", "order", i, "name", controllerName)
		if err = c.startController(customCtx, controllerName, controllerFlags); err != nil {
			c.runtimeLogger.Error(err, "controller start failed", "name", controllerName)
			return err
		}
	}
	// start remaining controllers
	c.runtimeLogger.Info("starting remaining controllers")
	for controllerName, _ := range c.controllersInitMap {
		if err = c.startController(customCtx, controllerName, controllerFlags); err != nil {
			return err
		}
	}

	return nil
}
