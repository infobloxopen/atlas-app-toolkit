package logging

import (
	"time"

	grpc_logging "github.com/grpc-ecosystem/go-grpc-middleware/logging"
	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
)

func NewLogger(level string) *logrus.Logger {
	log := logrus.StandardLogger()

	// Default timeFormat
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	// Set the log level on the default log based on command line flag
	logLevels := map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"info":    logrus.InfoLevel,
		"warning": logrus.WarnLevel,
		"error":   logrus.ErrorLevel,
		"fatal":   logrus.FatalLevel,
		"panic":   logrus.PanicLevel,
	}

	if lvl, ok := logLevels[level]; !ok {
		log.Errorf("Invalid %q provided for log level", lvl)
		log.SetLevel(logrus.InfoLevel)
	} else {
		log.SetLevel(lvl)
	}

	return log
}

var (
	defaultOptions = &options{
		levelFunc:    nil,
		shouldLog:    grpc_logging.DefaultDeciderMethod,
		codeFunc:     grpc_logging.DefaultErrorToCode,
		durationFunc: grpc_logrus.DefaultDurationToField,
		ophIDEnabled: false,
	}
)

type options struct {
	levelFunc    grpc_logrus.CodeToLevel
	shouldLog    grpc_logging.Decider
	codeFunc     grpc_logging.ErrorToCode
	durationFunc grpc_logrus.DurationToField
	ophIDEnabled bool
}

type Option func(*options)

// Enables ophid field in interceptors logs
func EnableOphID(opts *options) {
	opts.ophIDEnabled = true
}

func evaluateServerOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = grpc_logrus.DefaultCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func evaluateClientOpt(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	optCopy.levelFunc = grpc_logrus.DefaultClientCodeToLevel
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}
