# Health

Package contains tools that allow to add REST endpoint for handling `health` and `readiness` requests.
Package also provides implementation for most typical checks that may be done to figure our if service is healthy/ready.

## HTTP Handler

A handler to serve checks using `/health` and `/ready` endpoints can be added to your service in a following way: 

```go
	healthChecker := health.NewChecksHandler("/health", "/ready"),
	)
	healthChecker.AddReadiness("Readiness check name", readinessCheck)
	healthChecker.AddLiveness("Liveness check name", livenessCheck)

	s, err := server.NewServer(
		// register our health checks
		server.WithHealthChecks(healthChecker),
	)
```

In a sample above `readinessCheck` and `livenessCheck` are functions that perform a single check.
It is possible to add multiple checks independently and handler will call them in sequence.

## Checks

Packange provides several predefined checks:
 - `DNSProbeCheck(host string, timeout time.Duration)` returns a `Check` that determines whether a service with specified DNS name is reachable or not using net.Resolver's LookupHost method.
 - `HTTPGetCheck(url string, timeout time.Duration)` returns a `Check` that performs an HTTP GET request to the specified URL. It fails if timeout was exceeded or non-200-OK status code returned.
