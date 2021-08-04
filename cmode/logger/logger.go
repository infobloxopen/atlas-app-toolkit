package logger

import (
	"github.com/sirupsen/logrus"
)

type CModeLogger struct {
	*logrus.Logger
}

func New(logger *logrus.Logger) *CModeLogger {
	return &CModeLogger{logger}
}

func (l *CModeLogger) Name() string {
	return "loglevel"
}

func (l *CModeLogger) Get() string {
	return l.Level.String()
}

func (l *CModeLogger) ParseAndSet(val string) error {
	level, err := logrus.ParseLevel(val)
	if err != nil {
		return err
	}
	l.SetLevel(level)
	return nil
}

func (l *CModeLogger) Description() string {
	return "set logging level"
}

func (l *CModeLogger) ValidValues() []string {
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
