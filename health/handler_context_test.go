package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewChecksContextHandler(t *testing.T) {
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
			expectedCode: http.StatusOK,
			failHealth:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			h := NewChecksContextHandler("/healthz", "/ready")

			if test.failHealth {
				h.AddLiveness("Liveness check test", func(_ context.Context) error {
					return errors.New("Liveness check failed")
				})
			}

			if test.failReady {
				h.AddReadiness("Readiness check test", func(_ context.Context) error {
					return errors.New("Readiness check failed")
				})
			}

			req, err := http.NewRequestWithContext(context.TODO(), test.method, test.path, nil)
			assert.NoError(t, err)

			reqStr := test.method + " " + test.path
			httpRecorder := httptest.NewRecorder()
			h.Handler().ServeHTTP(httpRecorder, req)
			assert.Equal(t, test.expectedCode, httpRecorder.Code,
				"Result codes don't match %q. [%s]", reqStr, test.name)
		})
	}
}

func addNiceLivenessContext(h CheckerContext, number int, counterCall *int) {
	h.AddLiveness("Liveness"+strconv.Itoa(number), func(context.Context) error {
		*counterCall++
		return nil
	})
}

func addFailedLivenessContext(h CheckerContext, number int, counterCall *int) {
	h.AddLiveness("Liveness"+strconv.Itoa(number), func(context.Context) error {
		*counterCall++
		return errors.New("Liveness" + strconv.Itoa(number) + " check failed")
	})
}

func TestNoFailFastHandlerContext(t *testing.T) {
	h := NewChecksContextHandler("/healthz", "/ready")

	counterCall := 0
	expectedCalls := 3

	addNiceLivenessContext(h, 1, &counterCall)
	addFailedLivenessContext(h, 2, &counterCall)
	addFailedLivenessContext(h, 3, &counterCall)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/healthz", nil)
	assert.NoError(t, err)

	httpRecorder := httptest.NewRecorder()
	h.Handler().ServeHTTP(httpRecorder, req)

	assert.Equal(t, expectedCalls, counterCall, "Excepted %d calls of check.", expectedCalls)
	assert.Equal(t, http.StatusServiceUnavailable, httpRecorder.Code,
		"Result codes don't match, current is '%s'.", http.StatusText(httpRecorder.Code))
}

func TestFailFastHandlerContext(t *testing.T) {
	h := NewChecksContextHandler("/healthz", "/ready")
	h.SetFailFast(true)

	counterCall := 0
	expectedCalls := 2

	addNiceLivenessContext(h, 1, &counterCall)
	addFailedLivenessContext(h, 2, &counterCall)
	addFailedLivenessContext(h, 3, &counterCall)

	req, err := http.NewRequestWithContext(context.TODO(), http.MethodGet, "/healthz", nil)
	assert.NoError(t, err)

	httpRecorder := httptest.NewRecorder()
	h.Handler().ServeHTTP(httpRecorder, req)

	// we cannot determine the order of elements while iterationg over `checks` map in `handle` function
	// so we just check that amount is between the range
	assert.GreaterOrEqual(t, expectedCalls, counterCall, "Excepted less or equal %d calls of check.", expectedCalls)
	assert.NotEqual(t, 0, counterCall, "Cannot be zero calls of check.")
	assert.Equal(t, http.StatusServiceUnavailable, httpRecorder.Code,
		"Result codes don't match, current is '%s'.", http.StatusText(httpRecorder.Code))
}
