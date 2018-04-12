package health

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDNSProbeCheck(t *testing.T) {
	assert.NoError(t, DNSProbeCheck("google.com", 5*time.Second)())
	assert.Error(t, DNSProbeCheck("never.ever.where.com", 5*time.Second)())
}
