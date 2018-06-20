# Description

The common interfaces, types and helper functions to work with identifiers are defined in [resource](resource.go) file.

There are also some common implementations. 

# Codecs

The [fqstring](fqstring/codec.go) package provides a codec to encode and decode [infoblox.rpc.Identifier](../../rpc/resource/resource.proto) 
to and from [gorm.Identifier](resource.go).
The implementation converts `infoblox.rpc.Identifier` to a single string delimited by `/` (in **f**ully **q**ualified form).

The [uuid](uuid/codec.go) package provides a codec to encode and decode [infoblox.rpc.Identifier](../../rpc/resource/resource.proto) 
to and from [gorm.Identifier](resource.go).
The implementation validates that `resource_id` is a string in `UUID` format. If `resource_id` is not provided the `resource.Default` is returned,
thereby asking SQL DB engine to generate new `UUID` value.

The [integer](integer/codec.go) package provides a codec to encode and decode [infoblox.rpc.Identifier](../../rpc/resource/resource.proto) 
to and from [gorm.Identifier](resource.go).
The implementation validates that `resource_id` is an ineger value. If `resource_id` is not provided the `resource.Default` is returned,
thereby asking SQL DB engine to generate new `serial` value.

Please see [example](example_test.go) to see how to register codecs and encode/decode identifiers based on resource types of proto messages.

# protoc-gen-gorm

The plugin has support of `infoblox.rpc.Identifier` for **all** association types. All you need is to define your Primary/Foreign keys as
a fields of `infoblox.rpc.Identifier` type.
In the `XxxORM` generated models `infoblox.rpc.Identifier` will be replaced to `atlas-app-toolkit/gorm/resource.Identifier` type.

Let's define some PB resources
```proto
syntax = "proto3";

import "github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resource.proto";
import "github.com/infobloxopen/protoc-gen-gorm/options/gorm.proto";

option go_package = "github.com/yourapp/pb;pb";

message A {
    option (gorm.opts).ormable = true;
    
    infoblox.rpc.Identifier id = 1;
    string value = 2;
    repeated B b_list = 3; // has many
    infoblox.rpc.Identifier external = 4;
}

message B {
    option (gorm.opts).ormable = true;
    
    infoblox.rpc.Identifier id = 1;
    string value = 2;
    infoblox.rpc.Identifier a_id = 3; // foreign key to A  parent
}
```

In JSON it could look like
```json
{
  "id": "someapp/resource:a/1",
  "value": "someAvalue",
  "b_list": [
    {
      "id": "someapp/resource:b/1",
      "value": "someBvalue",
      "a_id": "someapp/resource:a/1"
    }
  ]
}
```

The generated code could be:
```go
import "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"

type AORM struct {
	Id *resource.Identifier
	Value string
	BList []*BORM
	External *resource.Identifier
}

type BORM struct {
	Id *resource.Identifier
	Value string
	AId *resource.Identifier
}
```

And then you need to register codecs in your `main.go` file, for example
```go
package main

import "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
import "github.com/infobloxopen/atlas-app-toolkit/gorm/resource/integer"
import "github.com/infobloxopen/atlas-app-toolkit/gorm/resource/fqstring"
import "github.com/yourapp/pb"

func main() {
	resource.RegisterCodec(integer.NewCodec("someapp", "resource:a"), &pb.A{})
	resource.RegisterCodec(integer.NewCodec("someapp", "resource:b"), &pb.B{})
	// register codec for external identifiers
	resource.RegisterCodec(fqstring.NewCodec(), nil) // nil means for all references that do not have internal PB type
}
```
