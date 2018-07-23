package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMakeStandardRequest(t *testing.T) {
	client := http.Client{}
	var tests = []struct {
		name    string
		method  string
		payload interface{}
		handler http.Handler
		err     error
	}{
		{
			name:    "check JWT presence in GET request",
			method:  http.MethodGet,
			payload: nil,
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					authHeader := r.Header.Get("Authorization")
					if expected := fmt.Sprintf(
						"%s %s", "token", standardToken,
					); expected != authHeader {
						t.Error("token missing in standard request")
					}
				},
			),
			err: nil,
		},
		{
			name:    "check JSON payload in POST request",
			method:  http.MethodPost,
			payload: map[string]string{"test-key": "test-value"},
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					body := make(map[string]string)
					if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
						t.Fatalf("unable to decode json payload: %v", err)
					}
					if value, ok := body["test-key"]; !ok || value != "test-value" {
						t.Errorf("unexpected json payload in request body: %v", body)
					}
				},
			),
			err: nil,
		},
		{
			name:    "check invalid JSON payload with unsupported type",
			method:  "not-an-http-verb",
			payload: nil,
			handler: http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {},
			),
			err: errors.New("net/http: invalid method"),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			testServer := httptest.NewServer(test.handler)
			defer testServer.Close()
			req, err := MakeStandardRequest(
				test.method, testServer.URL, test.payload,
			)
			if err != nil && test.err == nil {
				t.Errorf("unexpected error: %v", err)
			}
			client.Do(req)
		})
	}
}
