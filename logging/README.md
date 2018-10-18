## Per Request Logger Settings

The functions here support request scoped log settings.
Specifically, the log level and a custom log field can be set with the ctxlogrus context logger based on http Header values with the grpc-gateway or via grpc metadata.

The metadata annotator accepts input headers `-H "log-trace-key: <value>"` and `-H "log-level: <level>"`, but with or without it the input headers `-H "grpc-gateway-log-trace-key: <value>"` and `-H "grpc-gateway-log-level: <level>"` will also work.

The `log-trace-key` value is intended to be used for simplifying the process of isolating logs for a single request (so that the request-id doesn't have to be overridden or determined) or a class or requests.

The `LogLevelInterceptor` needs to be placed after the `ctxlogrus.UnaryServerInterceptor`, and accepts its own default logging level, so that the ctxlogrus interceptor (and the interceptors between it and this one) can be allowed to log at a different level than the proceeding ones even without setting it in the request (for example, to always/never log the Info message in the ctxlogrus interceptor, despite having a higher/lower log level).
Note that the `LogLevelInterceptor` cannot effect whether or not the Info level message in the `ctxlogrus.UnaryServerInterceptor` is printed or not.

The helper function `CopyLoggerWithLevel` can be used to make a deep copy of a logger at a new level, or using `CopyLoggerWithLevel(entry.Logger, level).WithFields(entry.Data)` can copy a logrus.Entry.
