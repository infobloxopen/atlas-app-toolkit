# Resource

A number of application services require a mechanism to encode the identity of a particular resource across applications.
The resource identifiers consist of the application ID, an application-defined resource type, and an application-defined ID for that resource.
The reference format captures the same information as the three-tuple format, in a single string delimited by `/`:

```
<app_id>/<resource_type>/<resource_id>
```

## Defining your proto

The common representation of resource identifiers defined in Protocol Buffer format
you could find in [resource.proto](resource.proto) file.
The `message Identifier` implements `jsonpb.JSONPBMarshaler` and `jsonpb.JSONPBUnmarshaler`
interfaces so that it renders itself in JSON as a string in a single string delimited by `/`.

You could use it to define identifiers in your proto messages, e.g.

```proto
syntax = "proto3";

import "github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resource.proto";

message MyMessage {
    atlas.rpc.Identifier id = 1;
    string some_data = 2;
    atlas.rpc.Identifier reference_on_external_resource = 3;
    atlas.rpc.Identifier foreign_id_of_internal_resource = 4;
}
```

Please give a read to [README](../../gorm/resource/README.md) of `gorm/resource`
package to see how it could be used with gorm and `protoc-gen-gorm` generated code. 

## Utility functions

Package provides several utility functions to operate with resource identifier.

### Check that an Identifier is nil

In order to check that an identifier is nil use `resource.Nil` function.

```go
package main

import (
	"fmt"
	
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func main() {
    var id *resource.Identifier
	
    if resource.Nil(id) {
    	fmt.Println("resource is nil identifier")
    }
}
```

See [unit test](nil_test.go) for more details.