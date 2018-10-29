package health

import (
	"context"
	"fmt"
	"net"
	"time"
)

// DNSProbeCheck returns a Check that determines whether a service
// with specified DNS name is reachable or not using net.Resolver's LookupHost method.
func DNSProbeCheck(host string, timeout time.Duration) Check {
	resolver := net.Resolver{}
	return func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		addrs, err := resolver.LookupHost(ctx, host)
		if err != nil {
			return err
		}
		if len(addrs) < 1 {
			return fmt.Errorf("could not resolve host")
		}
		return nil
	}
}
