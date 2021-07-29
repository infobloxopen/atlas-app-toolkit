package cmode

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var stubLoggerUsage = []string{
	"Usage:",
	fmt.Sprintf("GET  %s                           -- print usage", urlPath),
	fmt.Sprintf("GET  %s                    -- get current values", valuesUrlPath),
	fmt.Sprintf("POST %s?loglevel=$LOGLEVEL -- set logging level", valuesUrlPath),
	"",
	"valid loglevel values: [ error, info ]",
}

type stubLogger struct {
	*logrus.Logger
}

func newStubLogger(logger *logrus.Logger) stubLogger {
	return stubLogger{logger}
}

func (l *stubLogger) Name() string {
	return "loglevel"
}

func (l *stubLogger) Get() string {
	return l.Level.String()
}

func (l *stubLogger) ParseAndSet(val string) error {
	level, err := logrus.ParseLevel(val)
	if err != nil {
		return err
	}
	l.SetLevel(level)
	l.Infof("Logging level set to %v", level)
	return nil
}

func (l *stubLogger) Description() string {
	return "set logging level"
}

func (l *stubLogger) ValidValues() []string {
	return []string{
		"error",
		"info",
	}
}

type dataOpt string

func newDataOpt() dataOpt {
	return "none"
}

func (l *dataOpt) Name() string {
	return "data"
}

func (l *dataOpt) Get() string {
	return string(*l)
}

func (l *dataOpt) ParseAndSet(val string) error {
	switch val {
	case "all":
		*l = "all"
	case "none":
		*l = "none"
	default:
		return fmt.Errorf("not a valid data value: %s", val)
	}

	return nil
}

func (l *dataOpt) Description() string {
	return "set data value"
}

func (l *dataOpt) ValidValues() []string {
	return []string{
		"all",
		"none",
	}
}

func TestCModeUsage(t *testing.T) {
	logrusLogger := logrus.New()
	logger := newStubLogger(logrusLogger)

	var tests = []struct {
		name          string
		logger        *stubLogger
		expectedUsage []string
	}{
		{
			name:   "Logger is nil",
			logger: nil,
			expectedUsage: []string{
				"Usage:",
				fmt.Sprintf("GET  %s        -- print usage", urlPath),
				fmt.Sprintf("GET  %s -- get current values", valuesUrlPath),
			},
		},
		{
			name:          "Logger is not nil",
			logger:        &logger,
			expectedUsage: stubLoggerUsage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := New(test.logger)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			if req == nil {
				t.Fatal("Can not create http request")
			}

			rec := httptest.NewRecorder()
			if rec == nil {
				t.Fatal("Can not create http recorder")
			}

			handler := Handler(cm)
			handler.ServeHTTP(rec, req)

			reqRes := rec.Result()
			if reqRes.StatusCode != http.StatusOK {
				t.Fatalf("Hadnler returned wrong status code: %v", reqRes.StatusCode)
			}

			expectedUsage := strings.Join(test.expectedUsage, "\n")
			res := strings.TrimSpace(rec.Body.String())
			if res != expectedUsage {
				t.Fatalf("Hadnler returned unexpected usage: \n- Got \n%s \n- Want \n%s", res, expectedUsage)
			}
		})
	}
}

func TestCModeOpts(t *testing.T) {
	logrusLogger := logrus.New()
	logger := newStubLogger(logrusLogger)
	dataOpt := newDataOpt()

	logLevelExpectedError := []string{
		"invalid loglevel value: notvalid",
	}
	logLevelExpectedError = append(logLevelExpectedError, stubLoggerUsage...)

	var tests = []struct {
		name        string
		logger      *stubLogger
		opts        []CModeOpt
		value       string
		valueName   string
		expectedErr error
	}{
		{
			name:        "Set Logger to 'info'",
			logger:      &logger,
			opts:        nil,
			value:       "info",
			valueName:   "loglevel",
			expectedErr: nil,
		},
		{
			name:        "Set Logger to 'error'",
			logger:      &logger,
			opts:        nil,
			value:       "error",
			valueName:   "loglevel",
			expectedErr: nil,
		},
		{
			name:        "Set 'data' opt to 'all'",
			logger:      &logger,
			opts:        []CModeOpt{&dataOpt},
			value:       "all",
			valueName:   "data",
			expectedErr: nil,
		},
		{
			name:        "Try to pass not valid 'loglevel' val",
			logger:      &logger,
			opts:        nil,
			value:       "notvalid",
			valueName:   "loglevel",
			expectedErr: fmt.Errorf("%s", strings.Join(logLevelExpectedError, "\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := New(test.logger, test.opts...)
			path := fmt.Sprintf("%s?%s=%s", valuesUrlPath, test.valueName, test.value)

			reqSetValue := httptest.NewRequest(http.MethodPost, path, nil)
			if reqSetValue == nil {
				t.Fatal("Can not create http request")
			}

			rec := httptest.NewRecorder()
			if rec == nil {
				t.Fatal("Can not create http recorder")
			}

			handler := Handler(cm)
			handler.ServeHTTP(rec, reqSetValue)

			reqRes := rec.Result()
			if test.expectedErr == nil {
				if reqRes.StatusCode != http.StatusOK {
					t.Fatalf("Hadnler returned wrong status code: %v", reqRes.StatusCode)
				}

				for _, opt := range cm.opts {
					if opt.Name() == test.valueName {
						assert.Equal(t, test.value, opt.Get())
						return
					}
				}
			} else {
				if reqRes.StatusCode != http.StatusBadRequest {
					t.Fatalf("Hadnler returned wrong status code: %v", reqRes.StatusCode)
				}

				res := strings.TrimSpace(rec.Body.String())
				if res != test.expectedErr.Error() {
					t.Fatalf("Hadnler returned unexpected response: \n- Got \n%s \n- Want \n%s",
						res, test.expectedErr.Error())
				}
				return
			}

			t.Fatalf("There is no opt with name '%s'", test.valueName)
		})
	}
}
