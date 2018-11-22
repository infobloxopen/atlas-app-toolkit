# Request Info

Request Info contains following information about the request:
```
   ApplicationName string
   ResourceType    string
   ResourceId      string
   OperationType   int //Constant 
```

This package provide annotator witch extracts information from request and return it as metadata.

## Adding support for Request Info

You can enable support for Request Info in your gRPC-Server by adding the annotator to the middleware chain.

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

Once the middleware is included, the function below can be used to extract the Request Info anywhere it is needed.
```golang
requestInfo, err := requestinfo.FromContext(ctx)
```
The `err` represent error if something went wrong.
