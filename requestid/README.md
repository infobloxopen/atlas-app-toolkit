# RequestID Package

We support cases where each request need to be assigned with its own request_id, so when the request goes from one service to another it can be easily traced in logs.

Request-Id interceptor will extract Request-Id from incoming metadata and update context with the same.

##### How can I add support for Request-Id to track my requests?

You can enable support of Request-Id middleware in your gRPC-Server by adding

```golang
import (
  ...
  ...
  "github.com/infobloxopen/atlas-app-toolkit/requestid"
)
func main() {
    server := grpc.NewServer(
      grpc.UnaryInterceptor(
        grpc_middleware.ChainUnaryServer(  // middleware chain
          ...
          requestid.UnaryServerInterceptor(),  // Request-Id middleware
          ...
          ),
        ),
      )
    ...
}
```