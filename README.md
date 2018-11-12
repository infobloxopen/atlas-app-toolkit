# Atlas Application Toolkit
[![Build Status](https://img.shields.io/travis/infobloxopen/atlas-app-toolkit/master.svg?label=build)](https://travis-ci.org/infobloxopen/atlas-app-toolkit)

This repository provides common Go utilities and helpers that are reusable from project-to-project. The goal is to prevent code duplication by encouraging teams to use and contribute to toolkit libraries.

The toolkit is not a framework. Rather, it is a set of (mostly gRPC-related) plugins and helpers.

## Background

The toolkit's approach is based on the following assumptions.

- An application is composed of one or more independent services (micro-service architecture)
- Each independent service uses gRPC
- The REST API is presented by a separate service (gRPC Gateway) that serves as a reverse-proxy and forwards incoming HTTP requests to gRPC services

To get started with the toolkit, check out the [Atlas CLI](https://github.com/infobloxopen/atlas-cli) repository. The Atlas CLI's "bootstrap command" can generate new applications that make use of the toolkit. For Atlas newcomers, this is a great way to get up-to-speed with toolkit best practices.

## Toolkit Libraries

The following libraries have how-to guides included at the package level.

#### Request Handling Tools

[`requestid`](requestid) - gets the request ID from incoming requests (or creates a unique ID if it doesn't exist) 

[`query`](query) - provides query parameter-specific helpers, like sorting, paging, and filtering resources

[`auth`](auth) - includes authorization and authentication-related helpers

[`rpc`](rpc) - provides additional messages that can be utilized in your `proto` definitions

[`errors`](errors) - helps developers create detailed error messages or return multiple error messages

#### Server Utilities

[`server`](server) - provides a wrapper utility that manages a gRPC server and its REST gateway as a single unit

[`gateway`](gateway) - creates a gRPC gateway with built-in REST syntax compliancy 

[`health`](health) -  helps developers add health and readiness checks to their gRPC services

#### Database Utilities

[`gorm`](gorm) - offers a set of utilities for [GORM](http://gorm.io/) library

#### Testing

[`integration`](integration) - provides a set of utilities that help manage integration testing

## Core Concepts

If you are new to the gRPC world, the following resources will be helpful to you.

#### REST API Syntax Specification

To ensure that public REST API endpoints are consistent, the toolkit enforces API syntax requirements that are common for all applications at Infoblox. For more information about making your syntax-compliant (e.g. for error handling), see the [`errors`](errors), [`gateway`](gateway), and [`query`](query) packages.

#### gRPC Protobuf

See official documentation for [Protocol Buffer](https://developers.google.com/protocol-buffers/) and
for [gRPC](https://grpc.io/docs)

As an alternative you may use [this plugin](https://github.com/gogo/protobuf) to generate Golang code. That is the same
as official plugin but with [gadgets](https://github.com/gogo/protobuf/blob/master/extensions.md).

#### gRPC Gateway

See official [documentation](https://github.com/grpc-ecosystem/grpc-gateway)

#### gRPC Interceptors

One of the requirements to the API Toolkit is to support a Pipeline model.
We recommend to use gRPC server interceptor as middleware. See [examples](https://github.com/grpc-ecosystem/go-grpc-middleware)

## Example Toolkit Application

An example app that is based on api-toolkit can be found [here](https://github.com/infobloxopen/atlas-contacts-app)

## Related Tools

The following are toolkit-recommended utilities that are maintained in separate repositories.

#### Validation
We recommend to use [this validation plugin](https://github.com/lyft/protoc-gen-validate) to generate
`Validate` method for your gRPC requests.

As an alternative you may use [this plugin](https://github.com/mwitkow/go-proto-validators) too.

Validation can be invoked "automatically" if you add [this](https://github.com/grpc-ecosystem/go-grpc-middleware/tree/master/validator) middleware as a gRPC server interceptor.

#### Database Migrations
The toolkit does not require any specific method for database provisioning and setup. However, if [golang-migrate](https://github.com/golang-migrate/migrate) or the [infobloxopen fork](https://github.com/infobloxopen/migrate) of it is used, a couple helper functions are provided [here](gorm/version.go) for verifying that the database version matches a required version without having to import the entire migration package.

#### Documentation

We recommend to use [this plugin](https://github.com/pseudomuto/protoc-gen-doc) to generate documentation.

Documentation can be generated in different formats.

Here are several most used instructions used in documentation generation:

##### Leading comments

Leading comments can be used everywhere.

```proto
/**
 * This is a leading comment for a message
*/

message SomeMessage {
  // this is another leading comment
  string value = 1;
}
```

##### Trailing comments

Fields, Service Methods, Enum Values and Extensions support trailing comments.

```proto
enum MyEnum {
  DEFAULT = 0; // the default value
  OTHER   = 1; // the other value
}
```

##### Excluding comments

If you want to have some comment in your proto files, but don't want them to be part of the docs, you can simply prefix the comment with @exclude.

Example: include only the comment for the id field

```proto
/**
 * @exclude
 * This comment won't be rendered
 */
message ExcludedMessage {
  string id   = 1; // the id of this message.
  string name = 2; // @exclude the name of this message

  /* @exclude the value of this message. */
  int32 value = 3;
}
```

#### Swagger

Optionally you may generate [Swagger](https://swagger.io/) schema from your proto file.
To do so install [this plugin](https://github.com/grpc-ecosystem/grpc-gateway/tree/master/protoc-gen-swagger).

```sh
go get -u github.com/golang/protobuf/protoc-gen-go
```

Then invoke it as a plugin for Proto Compiler

```sh
protoc -I/usr/local/include -I. \
  -I$GOPATH/src \
  -I$GOPATH/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
  --swagger_out=logtostderr=true:. \
  path/to/your_service.proto
```

##### How to add Swagger definitions in my proto scheme?

```proto
import "protoc-gen-swagger/options/annotations.proto";

option (grpc.gateway.protoc_gen_swagger.options.openapiv2_swagger) = {
  info: {
    title: "My Service";
    version: "1.0";
  };
  schemes: HTTP;
  schemes: HTTPS;
  consumes: "application/json";
  produces: "application/json";
};

message MyMessage {
  option (grpc.gateway.protoc_gen_swagger.options.openapiv2_schema) = {
    external_docs: {
      url: "https://infoblox.com/docs/mymessage";
      description: "MyMessage description";
    }
};
```

For more Swagger options see [this scheme](https://github.com/grpc-ecosystem/grpc-gateway/blob/master/protoc-gen-swagger/options/openapiv2.proto)

See example [contacts app](https://github.com/infobloxopen/atlas-contacts-app/blob/master/pkg/pb/contacts.proto).
Here is a [generated Swagger schema](https://github.com/infobloxopen/atlas-contacts-app/blob/master/pkg/pb/contacts.swagger.json).

**NOTE** [Well Known Types](https://developers.google.com/protocol-buffers/docs/reference/google.protobuf) are
generated in a bit unusual way:

```json
    "protobufEmpty": {
      "type": "object",
      "description": "service Foo {\n      rpc Bar(google.protobuf.Empty) returns (google.protobuf.Empty);\n    }\n\nThe JSON representation for `Empty` is empty JSON object `{}`.",
      "title": "A generic empty message that you can re-use to avoid defining duplicated\nempty messages in your APIs. A typical example is to use it as the request\nor the response type of an API method. For instance:"
    },
```

#### Atlas Protoc Gentool

For convenience purposes there is an atlas-gentool image available which contains a pre-installed set of often used plugins.
For more details see [infobloxopen/atlas-gentool](https://github.com/infobloxopen/atlas-gentool) repository.

