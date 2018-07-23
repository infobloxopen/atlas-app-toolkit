package integration

import (
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// break apart the token so it isn't one huge line
	standardToken = fmt.Sprintf("%s%s%s",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.",
		"eyJBY2NvdW50SUQiOiJUZXN0QWNjb3VudCJ9.",
		"Qt7yYMcN2rEYavhgbhVMV762RBmAVUd32mW_VDIAKvM",
	)
)

// DefaultContext returns a context that has a jwt for basic testing purposes
func TestMakeTestJWT(t *testing.T) {
	var tests = []struct {
		name          string
		expected      string
		signingMethod jwt.SigningMethod
		claims        jwt.Claims
		err           error
	}{
		{
			"standard claims and signing method",
			standardToken,
			jwt.SigningMethodHS256,
			StandardClaims,
			nil,
		},
		{
			"force error when signing",
			"",
			mockSigningMethod{},
			StandardClaims,
			jwt.ErrInvalidKey,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			token, err := MakeTestJWT(test.signingMethod, test.claims)
			if token != test.expected {
				t.Errorf("unexpected value when building test token: have %s, expected %s",
					token, test.expected,
				)
			}
			if err != test.err {
				t.Errorf("unexpected error: have %s, expected %s",
					err, test.err,
				)
			}
		})
	}
}

func TestStandardTestJWT(t *testing.T) {
	t.Run("check test token", func(t *testing.T) {
		token, err := StandardTestJWT()
		if err != nil {
			t.Fatalf("unexpected error when building standard test token: %v", err)
		}
		if token != standardToken {
			t.Errorf("unexpected token value: have %s, expected %s",
				token, standardToken,
			)
		}
	})
}

type mockSigningMethod struct{}

func (mockSigningMethod) Verify(string, string, interface{}) error { return nil }

func (mockSigningMethod) Sign(string, interface{}) (string, error) {
	// return an arbitrary signing-related error
	return "", jwt.ErrInvalidKey
}

func (mockSigningMethod) Alg() string { return "" }
