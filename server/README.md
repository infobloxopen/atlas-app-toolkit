# Server

## Server Wrapper

You can package your gRPC server along with your REST gateway, health checks and any other http endpoints using [`server.NewServer`](server/server.go):
```go
s, err := server.NewServer(
    server.WithGrpcServer(grpcServer),
    server.WithHealthChecks(healthChecks),
    server.WithGateway(
        gateway.WithEndpointRegistration("/v1/", server_test.RegisterHelloHandlerFromEndpoint),
        gateway.WithServerAddress(grpcL.Addr().String()),
    ),
)
if err != nil {
    log.Fatal(err)
}
// serve it by passing in net.Listeners for the respective servers
if err := s.Serve(grpcListener, httpListener); err != nil {
    log.Fatal(err)
}
```
You can see a full example [here](server/server_example_test.go).