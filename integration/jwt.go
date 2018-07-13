package integration

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/infobloxopen/atlas-app-toolkit/auth"
)

const (
	// TestSecret dummy secret used for signing test JWTs
	TestSecret = "some-secret-123"
)

var (
	// DefaultClaims is the standard payload inside a test JWT
	StandardClaims = jwt.MapClaims{
		auth.MultiTenancyField: "TestAccount",
	}
)

// MakeToken generates a token string based on the given jwt claims
func MakeTestJWT(method jwt.SigningMethod, claims jwt.Claims) (string, error) {
	token, err := jwt.NewWithClaims(
		method, claims,
	).SignedString([]byte(TestSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

func StandardTestJWT() (string, error) {
	return MakeTestJWT(jwt.SigningMethodHS256, StandardClaims)
}
