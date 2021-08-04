# CMode

Package implements tools to get/set provided options (opts) via HTTP requests.

## Examples

Package can be used to change log level dynamically in the running service

### HTTP Handler

```go
    // Create logrus logger for application
    appLogger := logrus.StandartLogger()

    // Create CMode option to change log level dynamically
    cmLogger := logger.New(appLogger)
    
    // Create CMode object, pass logger for logging (mandatory) and a CMode option (cmLogger).
    // You can add more self-made options (see /pkg/cmode/logger as a reference)
    cm := cmode.New(appLogger, cmLogger)

    http.Handle("/", cm.Handler())

    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        appLogger.Fatalf("Server fatal error - %s", err)
    }
```

### List of available opts and endpoints
```bash
# Curl
curl localhost:8080/cmode

# Kubectl
kubectl -n <ns> exec <pod> -- curl localhost:8080/cmode
```

### Get options values
```bash
# Curl
curl localhost:8080/cmode/values

# Kubectl
kubectl -n <ns> exec <pod> -- curl localhost:8080/cmode/values
```

### Set option value
```bash
# Curl
curl -X POST localhost:8080/cmode/values?loglevel=debug

# Kubectl
kubectl -n <ns> exec <pod> -- curl -X POST localhost:8080/cmode/values?loglevel=debug
```
