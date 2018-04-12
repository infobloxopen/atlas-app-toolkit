package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestHTTPGetCheck(t *testing.T) {
	assert.NoError(t, HTTPGetCheck("http://httpbin.org/get", 5*time.Second)(), "Simple HTTP GET request shouldn't fail")
	assert.Error(t, HTTPGetCheck("http://httpbin.org/relative-redirect/:1", 5*time.Second)(), "Redirrect is not HTTP 200: OK")
	assert.Error(t, HTTPGetCheck("http://httpbin.org/nonexistent", 5*time.Second)(), "Non-existing site exists")
}
