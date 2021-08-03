# CMode

Package implements tools to get/set provided options (opts) via HTTP requests.

## Example

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
curl SERVER_ADDR/cmode
```

### Get options values
```bash
curl SERVER_ADDR/cmode/values
```

### Set option value
```bash
curl -X POST SERVER_ADDR/cmode/values?OPT_NAME=OPT_VALUE
```
