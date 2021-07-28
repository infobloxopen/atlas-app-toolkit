package cmode

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

type LogrusWrapper struct {
	*logrus.Logger
}

func NewLogrusWrapper(logger *logrus.Logger) LogrusWrapper {
	return LogrusWrapper{logger}
}

func (l *LogrusWrapper) Name() string {
	return "loglevel"
}

func (l *LogrusWrapper) Get() string {
	return l.Level.String()
}

func (l *LogrusWrapper) ParseAndSet(val string) error {
	level, err := logrus.ParseLevel(val)
	if err != nil {
		return err
	}
	l.SetLevel(level)
	l.Infof("Logging level set to %v", level)
	return nil
}

func (l *LogrusWrapper) Description() string {
	return "set logging level"
}

func (l *LogrusWrapper) ValidValues() []string {
	return []string{
		"panic",
		"fatal",
		"error",
		"warning",
		"info",
		"debug",
		"trace",
	}
}

func ExampleCMode() {
	appLogger := logrus.New()

	cmLogger := NewLogrusWrapper(appLogger)
	cm := New(&cmLogger)

	http.Handle("/", Handler(cm))

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		appLogger.Fatalf("Server fatal error - %s", err)
	}
}
