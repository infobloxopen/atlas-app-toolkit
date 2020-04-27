package logging

import (
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
)

func New(level string) *logrus.Logger {
	log := logrus.New()

	// Default timeFormat
	log.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339Nano,
	})

	parsedLevel, err := logrus.ParseLevel(level)
	if err != nil {
		log.Errorf("Invalid %q level provided for log", level)
		log.SetLevel(logrus.InfoLevel)
		return log
	}

	log.SetLevel(parsedLevel)
	return log
}

type options struct {
	codeToLevel CodeToLevel
	fields      []string
	headers     []string
}

type Option func(*options)

func initOptions(opts []Option) *options {
	o := &options{
		codeToLevel: grpc_logrus.DefaultCodeToLevel,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}

// CodeToLevel function defines the mapping between gRPC return codes and interceptor log level.
// From https://github.com/grpc-ecosystem/go-grpc-middleware/blob/06f64829ca1f521d41cd6235a7a204a6566fb0dc/logging/logrus/options.go#L57
type CodeToLevel func(code codes.Code) logrus.Level

// WithLevels customizes the function for mapping gRPC return codes and interceptor log level statements.
// From https://github.com/grpc-ecosystem/go-grpc-middleware/blob/06f64829ca1f521d41cd6235a7a204a6566fb0dc/logging/logrus/options.go#L70
func WithLevels(f CodeToLevel) Option {
	return func(o *options) {
		o.codeToLevel = f
	}
}

// Allows to provide custom fields for logging which are expected to be in JWT token
func WithCustomFields(fields []string) Option {
	return func(o *options) {
		o.fields = fields
	}
}

// Allows to provide custom fields for logging from request headers
func WithCustomHeaders(headers []string) Option {
	return func(o *options) {
		o.headers = headers
	}
}

// From https://github.com/grpc-ecosystem/go-grpc-middleware/blob/cfaf5686ec79ff8344257723b6f5ba1ae0ffeb4d/logging/logrus/server_interceptors.go#L91
func levelLogf(entry *logrus.Entry, level logrus.Level, format string, args ...interface{}) {
	switch level {
	case logrus.DebugLevel:
		entry.Debugf(format, args...)
	case logrus.InfoLevel:
		entry.Infof(format, args...)
	case logrus.WarnLevel:
		entry.Warningf(format, args...)
	case logrus.ErrorLevel:
		entry.Errorf(format, args...)
	case logrus.FatalLevel:
		entry.Fatalf(format, args...)
	case logrus.PanicLevel:
		entry.Panicf(format, args...)
	}
}
