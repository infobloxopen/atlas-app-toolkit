package cmode

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/cmode/logger"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var logrusWrapperUsage = []string{
	"Usage:",
	fmt.Sprintf("GET  %s                           -- print usage", urlPath),
	fmt.Sprintf("GET  %s                    -- get current values", valuesUrlPath),
	fmt.Sprintf("POST %s?loglevel=$LOGLEVEL -- set logging level", valuesUrlPath),
	"",
	"valid loglevel values: [ panic, fatal, error, warning, info, debug, trace ]",
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
	logger := logger.New(logrusLogger)

	var tests = []struct {
		name          string
		opts          []CModeOpt
		expectedUsage []string
	}{
		{
			name: "No opts",
			opts: nil,
			expectedUsage: []string{
				"Usage:",
				fmt.Sprintf("GET  %s        -- print usage", urlPath),
				fmt.Sprintf("GET  %s -- get current values", valuesUrlPath),
			},
		},
		{
			name:          "Logger is in opts",
			opts:          []CModeOpt{logger},
			expectedUsage: logrusWrapperUsage,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := New(nil, test.opts...)

			req := httptest.NewRequest(http.MethodGet, urlPath, nil)
			rec := httptest.NewRecorder()

			handler := cm.Handler()
			handler.ServeHTTP(rec, req)

			reqRes := rec.Result()
			if reqRes.StatusCode != http.StatusOK {
				t.Fatalf("Handler returned wrong status code: %v", reqRes.StatusCode)
			}

			expectedUsage := strings.Join(test.expectedUsage, "\n")
			res := strings.TrimSpace(rec.Body.String())
			if res != expectedUsage {
				t.Fatalf("Handler returned unexpected usage: \n- Got \n%s \n- Want \n%s", res, expectedUsage)
			}
		})
	}
}

func TestCModeOpts(t *testing.T) {
	logrusLogger := logrus.New()
	logger := logger.New(logrusLogger)
	dataOpt := newDataOpt()

	logLevelExpectedError := []string{
		"invalid loglevel value: notvalid",
	}
	logLevelExpectedError = append(logLevelExpectedError, logrusWrapperUsage...)

	var tests = []struct {
		name        string
		opts        []CModeOpt
		value       string
		valueName   string
		expectedErr error
	}{
		{
			name:        "Set Logger to 'info'",
			opts:        []CModeOpt{logger},
			value:       "info",
			valueName:   "loglevel",
			expectedErr: nil,
		},
		{
			name:        "Set Logger to 'error'",
			opts:        []CModeOpt{logger},
			value:       "error",
			valueName:   "loglevel",
			expectedErr: nil,
		},
		{
			name:        "Set 'data' opt to 'all'",
			opts:        []CModeOpt{&dataOpt},
			value:       "all",
			valueName:   "data",
			expectedErr: nil,
		},
		{
			name:        "Try to pass not valid 'loglevel' val",
			opts:        []CModeOpt{logger},
			value:       "notvalid",
			valueName:   "loglevel",
			expectedErr: fmt.Errorf("%s", strings.Join(logLevelExpectedError, "\n")),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			cm := New(logrusLogger, test.opts...)
			path := fmt.Sprintf("%s?%s=%s", valuesUrlPath, test.valueName, test.value)

			reqSetValue := httptest.NewRequest(http.MethodPost, path, nil)
			rec := httptest.NewRecorder()

			handler := cm.Handler()
			handler.ServeHTTP(rec, reqSetValue)

			reqRes := rec.Result()
			if test.expectedErr == nil {
				if reqRes.StatusCode != http.StatusOK {
					t.Fatalf("Handler returned wrong status code: %v", reqRes.StatusCode)
				}

				expectedReply := fmt.Sprintf("%s is set to %s", test.valueName, test.value)

				res := strings.TrimSpace(rec.Body.String())
				if res != expectedReply {
					t.Fatalf("Handler returned unexpected reply msg: \n- Got \n%s \n- Want \n%s",
						res, expectedReply)
				}

				for _, opt := range cm.opts {
					if opt.Name() == test.valueName {
						assert.Equal(t, test.value, opt.Get())
						return
					}
				}

				t.Fatalf("There is no opt with name '%s'", test.valueName)
			} else { // Error is expected
				if reqRes.StatusCode != http.StatusBadRequest {
					t.Fatalf("Handler returned wrong status code: %v", reqRes.StatusCode)
				}

				res := strings.TrimSpace(rec.Body.String())
				if res != test.expectedErr.Error() {
					t.Fatalf("Handler returned unexpected response: \n- Got \n%s \n- Want \n%s",
						res, test.expectedErr.Error())
				}
			}
		})
	}
}
