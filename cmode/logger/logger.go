package logger

import (
	"github.com/sirupsen/logrus"
)

type CmodeLogger struct {
	*logrus.Logger
}

func NewCmodeLogger(logger *logrus.Logger) CmodeLogger {
	return CmodeLogger{logger}
}

func (l *CmodeLogger) Name() string {
	return "loglevel"
}

func (l *CmodeLogger) Get() string {
	return l.Level.String()
}

func (l *CmodeLogger) ParseAndSet(val string) error {
	level, err := logrus.ParseLevel(val)
	if err != nil {
		return err
	}
	l.SetLevel(level)
	l.Infof("Logging level set to %v", level)
	return nil
}

func (l *CmodeLogger) Description() string {
	return "set logging level"
}

func (l *CmodeLogger) ValidValues() []string {
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
