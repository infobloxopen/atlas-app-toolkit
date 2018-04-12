package health

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewHandler(t *testing.T) {
	var tests = []struct {
		name         string
		method       string
		path         string
		failHealth   bool
		failReady    bool
		expectedCode int
	}{
		{
			name:         "Non-existent URL",
			method:       http.MethodPost,
			path:         "/neverwhere",
			expectedCode: http.StatusNotFound,
		},
		{
			name:         "POST Method not allowed health",
			method:       http.MethodPost,
			path:         "/healthz",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "POST Method not allowed ready",
			method:       http.MethodPost,
			path:         "/ready",
			expectedCode: http.StatusMethodNotAllowed,
		},
		{
			name:         "No checks health",
			method:       http.MethodGet,
			path:         "/healthz",
			expectedCode: http.StatusOK,
		},
		{
			name:         "No checks ready",
			method:       http.MethodGet,
			path:         "/ready",
			expectedCode: http.StatusOK,
		},
		{
			name:         "Health succeed Ready fail health",
			method:       http.MethodGet,
			path:         "/healthz",
			expectedCode: http.StatusOK,
			failReady:    true,
		},
		{
			name:         "Health succeed Ready fail ready",
			method:       http.MethodGet,
			path:         "/ready",
			expectedCode: http.StatusServiceUnavailable,
			failReady:    true,
		},
		{
			name:         "Health fail Ready succeed health",
			method:       http.MethodGet,
			path:         "/healthz",
			expectedCode: http.StatusServiceUnavailable,
			failHealth:   true,
		},
		{
			name:         "Health fail Ready succeed ready",
			method:       http.MethodGet,
			path:         "/ready",
			expectedCode: http.StatusServiceUnavailable,
			failHealth:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := NewChecksHandler("/healthz", "/ready")

			if test.failHealth {
				h.AddLiveness("Liveness check test", func() error {
					return errors.New("Liveness check failed")
				})
			}

			if test.failReady {
				h.AddReadiness("Readiness check test", func() error {
					return errors.New("Readiness check failed")
				})
			}

			req, err := http.NewRequest(test.method, test.path, nil)
			assert.NoError(t, err)

			reqStr := test.method + " " + test.path
			httpRecorder := httptest.NewRecorder()
			h.Handler().ServeHTTP(httpRecorder, req)
			assert.Equal(t, test.expectedCode, httpRecorder.Code,
				"Result codes don't match %q. [%s]", reqStr, test.name)
		})
	}
}
