package signals

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	ctrl "sigs.k8s.io/controller-runtime"
)

var signalsRegistered = false

type ExitSignals interface {
	Context() context.Context
	StopCh() <-chan struct{}
	DoneCh() chan<- error
}

type TestExitSignals interface {
	Context() context.Context
	StopCh() <-chan struct{}
	DoneCh() chan<- error
	CancelFunc() context.CancelFunc
}

type exitSignals struct {
	context context.Context
	stopCh  chan struct{}
	doneCh  chan<- error
	cancel  context.CancelFunc
}

func (e *exitSignals) Context() context.Context { return e.context }

func (e *exitSignals) StopCh() <-chan struct{} { return e.stopCh }

func (e *exitSignals) DoneCh() chan<- error { return e.doneCh }

// TestStopCh returns the real stopCh to allow testing of exit handlers
func (e *exitSignals) CancelFunc() context.CancelFunc { return e.cancel }

// SetupExitHandlers creates several exit signaling methods and propogates between the available types
// this function will panic if run more than once (since signal handlers can only be registered with one function)
func SetupExitHandlers(logger logr.Logger) ExitSignals {
	if signalsRegistered {
		panic(fmt.Errorf("SetupExitHandlers should only be run once within an application"))
	}
	// signalCh catches SIGTERM / SIGABRT
	signalChannel := ctrl.SetupSignalHandler()
	signalsRegistered = true

	return SetupExitHandlersWithExistingStopCh(logger, signalChannel)
}

// SetupExitHandlersWithExistingStopCh creates several exit signaling methods and propogates between the available types
// this function will panic if run more than once (since signal handlers can only be registered with one function)
func SetupExitHandlersWithExistingStopCh(logger logr.Logger, signalChannel <-chan struct{}) ExitSignals {
	logger = logger.WithName("signal-handler")

	ctx, cancel := context.WithCancel(context.Background())

	// doneCh is used for signaling final error
	// create this as a buffered channel in case many goroutines are sending error
	// and this allows them to not block while trying to exit
	doneChannel := make(chan error, 5)

	// stopChannel is used to aggregate the different exit conditions (ctx or signal) to a single channel
	stopChannel := make(chan struct{})

	go func() {
		exiting := false
		var err error
		for {
			logger.V(1).Info("signal handler", "exiting", exiting)
			select {
			case err = <-doneChannel:
				// ignore go-routines which are exiting without an error
				if err == nil {
					continue
				}
				logger.Error(err, "exiting due to error")
				exiting = true

			case <-signalChannel:
				logger.Error(err, "exiting due to signal")
				exiting = true

			case <-ctx.Done():
				logger.Error(err, "exiting due to context canceled")
				exiting = true
				return
			}
			if exiting {
				logger.V(1).Info("signal handler", "exiting", exiting)
				cancel()
				close(stopChannel)
				// close the error channel since this can hang a goroutine
				// close(doneChannel)
				return
			}
		}
	}()
	e := &exitSignals{ctx, stopChannel, doneChannel, cancel}
	return e
}
