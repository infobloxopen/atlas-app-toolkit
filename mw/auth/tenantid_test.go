package auth

import (
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

func TestGetTenantID(t *testing.T) {
	var tenantIDTests = []struct {
		claims   jwt.Claims
		expected string
		err      error
	}{
		{
			jwt.MapClaims{
				"TenantID": "tenantid-abc-123",
			},
			"tenantid-abc-123",
			nil,
		},
		{
			jwt.MapClaims{},
			"",
			errMissingTenantID,
		},
	}
	for _, test := range tenantIDTests {
		token := makeToken(test.claims, t)
		ctx, err := contextWithToken(token)
		tenantID, err := GetTenantID(ctx, nil)
		if err != test.err {
			t.Errorf("Invalid error value: %v - expected %v", err, test.err)
		}
		if tenantID != test.expected {
			t.Errorf("Invalid tenant ID: %v - expected %v", tenantID, test.expected)
		}
	}
}
