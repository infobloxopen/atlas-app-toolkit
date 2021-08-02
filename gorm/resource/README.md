# Description

The common interfaces, types and helper functions to work with identifiers are defined in [resource](resource.go) file.
Please see [example](example_test.go) to see how to work with identifiers.

# Codecs

You could register `resource.Codec` for you PB type to be used to convert `atlas.rpc.Identifier` that defined for that type.

By default if PB resource is undefined (`nil`) the `atlas.rpc.Identifier` is converted to a string in fully qualified format specified for
Atlas References, otherwise the Resource ID part is returned as string value.

If `driver.Value` is `nil` default codecs returns `nil` `atlas.rpc.Identifier`, if you want to override such behavior in order to return empty
`atlas.rpc.Identifier` that could be rendered to `null` string in JSON - 

If `resource.Codec` is not registered for a PB type the value of identifier is converted from `driver.Value` to a string.
If Resource Type is not found it is populated from the name of PB type,
the Application Name is populated if was registered. (see `RegisterApplication`).

# protoc-gen-gorm

The plugin has support of `atlas.rpc.Identifier` for **all** association types. All you need is to define your Primary/Foreign keys as
a fields of `atlas.rpc.Identifier` type.

In the `XxxORM` generated models `atlas.rpc.Identifier` will be replaced to the type specified in `(gorm.field).tag` option.
The only numeric and text formats are supported. If type is not set it will be generated as `interface{}`.

If you want to expose foreign keys on API just leave them with empty type in `gorm.field.tag` and it will be calculated based on the
parent's primary key type.

By default `Identifier`s are nillable, it means that for primary keys you need to set corresponding tag `primary_key: true` and for foreign keys
and external references `not_null: true`.

The Postgres types from tags are converted as follows:

```go
    switch type_from_tag {
    case "uuid", "text", "char", "array", "cidr", "inet", "macaddr":
        orm_model_field = "string"
    case "smallint", "integer", "bigint", "numeric", "smallserial", "serial", "bigserial":
        orm_model_field = "int64"
    case "jsonb", "bytea":
        orm_model_field = "[]byte"
    case "":
        orm_model_field = "interface{}"
    default:
        return errors.New("unknown tag type of atlas.rpc.Identifier")
    }
```

**NOTE** Be sure to set type properly for association fields.

**Step 1** Let's define PB resources

```proto
syntax = "proto3";

import "github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resource.proto";
import "github.com/infobloxopen/protoc-gen-gorm/options/gorm.proto";

option go_package = "github.com/yourapp/pb;pb";

message A {
    option (gorm.opts).ormable = true;
    
    atlas.rpc.Identifier id = 1 [(gorm.field).tag = {type: "integer" primary_key: true}];
    string value = 2;
    repeated B b_list = 3; // has many
    atlas.rpc.Identifier external = 4 [(gorm.field).tag = {type: "text"}];
}

message B {
    option (gorm.opts).ormable = true;
    
    atlas.rpc.Identifier id = 1 [(gorm.field).tag = {type: "integer" primary_key: true}];
    string value = 2;
     // foreign key to A  parent. !!! Will be set to the type of A.id
    atlas.rpc.Identifier a_id = 3;
    atlas.rpc.Identifier external_not_null = 4 [(gorm.field).tag = {type: "text" not_null: true}];
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
	Id int64
	Value string
	BList []*BORM
	External *string
}

type BORM struct {
	Id int64
	Value string
	AId *int64
	ExternalNotNull string
}
```

**Step 2** The last thing you need to make identifiers work correctly is to register the name of your application,
that name will be used during encoding to populate ApplicationName field of `atlas.rpc.Identifier`.

```go
package main

import "github.com/infobloxopen/atlas-app-toolkit/gorm/resource"

func main() {
    resource.RegisterApplication("MyAppName")
}
```

# FAQ

## How to customize name of my PB type?

Implement `Namer` interface (`ResourceName() string` function) or add `XXX_MessageName() string` method to you PB type. See [Name](resource.go) function.

## I want to validate/generate atlas.rpc.Identifier for PB types in my application

Implement [resource.Codec](resource.go) for your PB types and you will be given a full control on how `Identifier`s 
are converted to/from `driver.Value`. 
