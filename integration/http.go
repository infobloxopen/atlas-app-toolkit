package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/infobloxopen/atlas-app-toolkit/auth"
)

// MakeStandardRequest issues an HTTP request a specific endpoint with Atlas-specific
// request data (e.g. the authorization token)
func MakeStandardRequest(method, url string, payload interface{}) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	token, err := StandardTestJWT()
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", auth.DefaultTokenType, token))
	return req, nil
}
