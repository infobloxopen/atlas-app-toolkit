package server_test

import (
	"fmt"
	"net/http"

	"log"

	"io/ioutil"

	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	server_test "github.com/infobloxopen/atlas-app-toolkit/server/testdata"
	"github.com/infobloxopen/atlas-app-toolkit/servertest"
	"golang.org/x/net/context"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
)

func Example() {
	// for real-world apps, these net.Listeners will be created from addresses passed in from args (flag or env or whatever)
	grpcL, err := servertest.NewLocalListener()
	if err != nil {
		log.Fatal(err)
	}
	httpL, err := servertest.NewLocalListener()
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	server_test.RegisterHelloServer(grpcServer, &server_test.HelloServerImpl{})

	healthChecks := health.NewChecksHandler("healthz", "ready")
	healthChecks.AddLiveness("grpc", func() error {
		_, err := grpc.Dial(grpcL.Addr().String(), grpc.WithInsecure())
		return err
	})

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

	// normally, this would be the end of your main.go implementation. For the sake of this exampleClient, we'll make a
	// simple request for demonstration
	go s.Serve(grpcL, httpL)
	defer s.Stop()

	// demonstrate making a gRPC request through the grpc server url
	conn, err := grpc.Dial(grpcL.Addr().String(), grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := server_test.NewHelloClient(conn)
	gResp, err := client.SayHello(context.Background(), &server_test.HelloRequest{Name: "exampleClient"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(gResp.Greeting)

	// demonstrate making a health check against the http url
	hResp, err := http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/healthz"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(hResp.StatusCode)

	// demonstrate making a REST request against the http url
	gwResp, err := http.Get(fmt.Sprint("http://", httpL.Addr().String(), "/v1/hello?name=exampleREST"))
	if err != nil {
		log.Fatal(err)
	}
	respBytes, err := ioutil.ReadAll(gwResp.Body)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(respBytes))

	// Output:
	// hello, exampleClient!
	// 200
	// {"greeting":"hello, exampleREST!"}
}
