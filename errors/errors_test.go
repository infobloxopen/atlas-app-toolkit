package errors

import (
	"testing"

	"google.golang.org/grpc/codes"
)

// checkErr function checks err fields.
func checkErr(t *testing.T, c error, code codes.Code, message string) {
	v, ok := c.(*Container)

	if !ok {
		t.Errorf("Invalid error")
	}

	if v.errMessage != message ||
		v.errCode != code {
		t.Errorf("Expected code=%d, message=%q, got code=%d, message=%q",
			code, message, v.errCode, v.errMessage)
	}
}
