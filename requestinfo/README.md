# Request Info

This package provide annotator witch extract information from request and return it as metadata  
## Adding support for Request Info

You can enable support for Request Info in your gRPC-Server by adding the annotator to the middleware chain.
However, the **ordering is important**.
If you also use the grpc_logrus interceptor, the request-id middleware should be later in the middleware chain, but should also be before any other service middlewares to ensure it is present in the context to be included in those requests.

```golang
import (
  ...
  ...
  "github.com/infobloxopen/atlas-app-toolkit/requestinfo"
)
func main() {
    server := grpc.NewServer(
      grpc.UnaryInterceptor(
        runtime.WithMetadata(requestinfo.MetadataAnnotator)
      )
    ...
}
```

## Extracting the Request Info

Once the middleware is included, the following function
```golang
requestInfo, err := requestinfo.FromContext(ctx)
```
can extract the request info anywhere it is needed.
The `err` field represent error if something went wrong.
