package integration

import (
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/infobloxopen/atlas-app-toolkit/auth"
)

const (
	// testSecret is a dummy secret used for signing test JWTs
	testSecret = "some-secret-123"
)

var (
	// StandardClaims is the standard payload inside a test JWT
	StandardClaims = jwt.MapClaims{
		auth.MultiTenancyField: "TestAccount",
	}
)

// MakeTestJWT generates a token string based on the given JWT claims
func MakeTestJWT(method jwt.SigningMethod, claims jwt.Claims) (string, error) {
	token, err := jwt.NewWithClaims(
		method, claims,
	).SignedString([]byte(testSecret))
	if err != nil {
		return "", err
	}
	return token, nil
}

// StandardTestJWT builds a JWT with the standard test claims in the JWT payload
func StandardTestJWT() (string, error) {
	return MakeTestJWT(jwt.SigningMethodHS256, StandardClaims)
}
