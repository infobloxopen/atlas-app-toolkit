# Logging

This package supports extended settings for request scoped logging.
Specifically, the log level (severity below which to suppress) and a custom log field can be set with the [context logger](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/logging/logrus), based on http Header values with the grpc-gateway or via grpc metadata.
The context logger should then be used inside grpc method implementations instead of a global logger.

The custom field `log-trace-key` value is intended to be used for simplifying the process of isolating logs for a single request or a set of requests.
This is a similar goal to request-ids, but request-ids are randomly generated for uniqueness, making them less versatile for debugging purposes.

## Enabling request-scoped logger settings

To enable these features, the `LogLevelInterceptor` and the `grpc_logrus.UnaryServerInterceptor` have to be included in the server's middleware chain.

The `LogLevelInterceptor` needs to be placed after the `grpc_logrus.UnaryServerInterceptor` in the chain, and accepts its own default logging level, so that the grpc_logrus interceptor (and the interceptors between it and this one) can be allowed to log at a different level than the proceeding ones even without setting it in the request (for example, to always/never log the Info message in the ctxlogrus interceptor, despite having a higher/lower log level).
Note that the `LogLevelInterceptor` cannot effect whether or not the Info level message in the `grpc_logrus.UnaryServerInterceptor` is printed or not.

The middleware chain code should look something like this:
```golang
import (
	"github.com/infobloxopen/atlas-app-toolkit/logging"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

func main() {
	server := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer( // middleware chain
				grpc_logrus.UnaryServerInterceptor(logrus.NewEntry(logger)),
				logging.LogLevelInterceptor(logger.Level), // Request-scoped logging middleware
				...
			),
		),
	)
	...
}
```

For grpc-gateway support, the `MetadataAnnotator` should also be added to the gateway.
Using the toolkit's server package, that setup looks something like this:
```golang
import (
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/logging"
	"github.com/infobloxopen/atlas-app-toolkit/server"
)

func main() {
	gatewayOptions := []runtime.ServeMuxOption{
		runtime.WithMetadata(logging.MetadataAnnotator),
		...
	}

	server.NewServer(
		server.WithGrpcServer(grpcServer),
		server.WithGateway(
			gateway.WithGatewayOptions(gatewayOptions...),
			...
		),
	)
}
```

## Using the request-scoped logger settings

When using the metadata annotator and the grpc-gateway, http requests using headers `-H "log-trace-key: <value>"` and `-H "log-level: <level>"` will be stored in the grpc metadata for consumption by the interceptor.

Without the grpc-gateway, the [metadata](https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-metadata.md) has to be added to the request context directly.
```golang
ctx := metadata.AppendToOutgoingContext(ctx, "log-level", "debug", "log-trace-key", "foobar")

// make unary RPC
response, err := client.SomeRPC(ctx, someRequest)
```

## Gateway logging

Certain client interceptors may reject incoming queries (e.g. due to non-conformant json fields).
This will cause a logging gap compared to queries that fail in the server. To alleviate this, an enhanced gateway logging interceptor is provided.
The `GatewayLoggingInterceptor` should be in the middleware chain before any that could error out.
The `GatewayLoggingSentinelInterceptor` should be the very last middleware in the chain.

For example:
```golang
...
	grpc_middleware.ChainUnaryClient(
		[]grpc.UnaryClientInterceptor{
			GatewayLoggingInterceptor(logger, EnableDynamicLogLevel, EnableAccountID),
			...
			GatewayLoggingSentinelInterceptor(),
		},
	)
...
```

## Other functions

The helper function `CopyLoggerWithLevel` can be used to make a deep copy of a logger at a new level, or using `CopyLoggerWithLevel(entry.Logger, level).WithFields(entry.Data)` can copy a logrus.Entry.
