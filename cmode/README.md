# CMode

Package implements tools to get/set provided options (opts) via HTTP requests.

## Example

### HTTP Handler

```go
    appLogger := logger.New() // Implements CModeLogger
    someOption := NewOption() // CModeOpt

    cm := cmode.New(appLogger, someOption)

    http.Handle("/", Handler(cm))

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