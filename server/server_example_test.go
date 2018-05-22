package server_test

import (
	"fmt"
	"io"
	"math"
	"net/http"

	"log"

	"io/ioutil"

	"github.com/golang/protobuf/proto"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/grpc-ecosystem/grpc-gateway/utilities"
	"github.com/infobloxopen/atlas-app-toolkit/gateway"
	"github.com/infobloxopen/atlas-app-toolkit/health"
	"github.com/infobloxopen/atlas-app-toolkit/server"
	"github.com/infobloxopen/atlas-app-toolkit/servertest"
	"golang.org/x/net/context"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/status"
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
	RegisterHelloServer(grpcServer, &ExampleServerImpl{})

	healthChecks := health.NewChecksHandler("healthz", "ready")
	healthChecks.AddLiveness("grpc", func() error {
		_, err := grpc.Dial(grpcL.Addr().String(), grpc.WithInsecure())
		return err
	})

	s, err := server.NewServer(
		server.WithGrpcServer(grpcServer),
		server.WithHealthChecks(healthChecks),
		server.WithGateway(
			gateway.WithEndpointRegistration("/v1/", RegisterHelloHandlerFromEndpoint),
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
	client := NewHelloClient(conn)
	gResp, err := client.SayHello(context.Background(), &HelloRequest{Name: "exampleClient"})
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

// The remaining code is copied from the following rendered .proto file:
//	syntax = "proto3";
//	package server;
//	option go_package = "github.com/infobloxopen/atlas-app-toolkit/server/testdata;server_test";
//
//	import "google/api/annotations.proto";
//
//	message HelloRequest {
//	string name = 1;
//	}
//
//	message HelloResponse {
//	string greeting = 1;
//	}
//
//	service Hello {
//	rpc SayHello (HelloRequest) returns (HelloResponse) {
//	option (google.api.http) = {
//	get: "/hello"
//	};
//	}
//	}

type ExampleServerImpl struct{}

func (ExampleServerImpl) SayHello(ctx context.Context, req *HelloRequest) (*HelloResponse, error) {
	return &HelloResponse{Greeting: fmt.Sprintf("hello, %s!", req.Name)}, nil
}

var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type HelloRequest struct {
	Name string `protobuf:"bytes,1,opt,name=name" json:"name,omitempty"`
}

func (m *HelloRequest) Reset()                    { *m = HelloRequest{} }
func (m *HelloRequest) String() string            { return proto.CompactTextString(m) }
func (*HelloRequest) ProtoMessage()               {}
func (*HelloRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *HelloRequest) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

type HelloResponse struct {
	Greeting string `protobuf:"bytes,1,opt,name=greeting" json:"greeting,omitempty"`
}

func (m *HelloResponse) Reset()                    { *m = HelloResponse{} }
func (m *HelloResponse) String() string            { return proto.CompactTextString(m) }
func (*HelloResponse) ProtoMessage()               {}
func (*HelloResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *HelloResponse) GetGreeting() string {
	if m != nil {
		return m.Greeting
	}
	return ""
}

func init() {
	proto.RegisterType((*HelloRequest)(nil), "server.HelloRequest")
	proto.RegisterType((*HelloResponse)(nil), "server.HelloResponse")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Hello service

type HelloClient interface {
	SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error)
}

type helloClient struct {
	cc *grpc.ClientConn
}

func NewHelloClient(cc *grpc.ClientConn) HelloClient {
	return &helloClient{cc}
}

func (c *helloClient) SayHello(ctx context.Context, in *HelloRequest, opts ...grpc.CallOption) (*HelloResponse, error) {
	out := new(HelloResponse)
	err := grpc.Invoke(ctx, "/server.Hello/SayHello", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Hello service

type HelloServer interface {
	SayHello(context.Context, *HelloRequest) (*HelloResponse, error)
}

func RegisterHelloServer(s *grpc.Server, srv HelloServer) {
	s.RegisterService(&_Hello_serviceDesc, srv)
}

func _Hello_SayHello_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(HelloRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(HelloServer).SayHello(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/server.Hello/SayHello",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(HelloServer).SayHello(ctx, req.(*HelloRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _Hello_serviceDesc = grpc.ServiceDesc{
	ServiceName: "server.Hello",
	HandlerType: (*HelloServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SayHello",
			Handler:    _Hello_SayHello_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "github.com/infobloxopen/atlas-app-toolkit/server/testdata/test.proto",
}

func init() {
	proto.RegisterFile("github.com/infobloxopen/atlas-app-toolkit/server/testdata/test.proto", fileDescriptor0)
}

var fileDescriptor0 = []byte{
	// 238 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x9c, 0x90, 0xbf, 0x4a, 0x04, 0x31,
	0x10, 0xc6, 0x39, 0xd1, 0x65, 0x0d, 0x6a, 0x11, 0x14, 0x64, 0xb1, 0x90, 0xad, 0x04, 0xb9, 0x04,
	0xb4, 0xb4, 0x13, 0xe5, 0x2c, 0xe5, 0xec, 0x6c, 0x64, 0x56, 0xc7, 0x5c, 0x30, 0x37, 0x13, 0x93,
	0x39, 0xd1, 0xd6, 0x57, 0xf0, 0xd1, 0x7c, 0x05, 0x1f, 0x44, 0x2e, 0x39, 0x44, 0x2c, 0xed, 0xbe,
	0xef, 0x9b, 0x1f, 0xf3, 0x4f, 0x5d, 0x38, 0x2f, 0xb3, 0xc5, 0x60, 0xee, 0x79, 0x6e, 0x3d, 0x3d,
	0xf2, 0x10, 0xf8, 0x95, 0x23, 0x92, 0x05, 0x09, 0x90, 0xc7, 0x10, 0xe3, 0x58, 0x98, 0xc3, 0x93,
	0x17, 0x9b, 0x31, 0xbd, 0x60, 0xb2, 0x82, 0x59, 0x1e, 0x40, 0xa0, 0x08, 0x13, 0x13, 0x0b, 0xeb,
	0xa6, 0xd6, 0xba, 0x03, 0xc7, 0xec, 0x02, 0x5a, 0x88, 0xde, 0x02, 0x11, 0x0b, 0x88, 0x67, 0xca,
	0x95, 0xea, 0x7b, 0xb5, 0x75, 0x85, 0x21, 0xf0, 0x14, 0x9f, 0x17, 0x98, 0x45, 0x6b, 0xb5, 0x4e,
	0x30, 0xc7, 0xfd, 0xd1, 0xe1, 0xe8, 0x68, 0x73, 0x5a, 0x74, 0x7f, 0xac, 0xb6, 0x57, 0x4c, 0x8e,
	0x4c, 0x19, 0x75, 0xa7, 0x5a, 0x97, 0x10, 0xc5, 0x93, 0x5b, 0x81, 0x3f, 0xfe, 0xe4, 0x5a, 0x6d,
	0x14, 0x58, 0x4f, 0x54, 0x7b, 0x03, 0x6f, 0x55, 0xef, 0x9a, 0xba, 0x8c, 0xf9, 0x3d, 0xab, 0xdb,
	0xfb, 0x93, 0xd6, 0xee, 0xfd, 0xce, 0xfb, 0xe7, 0xd7, 0xc7, 0x5a, 0xab, 0x1b, 0x3b, 0x5b, 0xe6,
	0xe7, 0x93, 0xdb, 0xcb, 0x7f, 0x3f, 0xe4, 0xac, 0xfa, 0xbb, 0xa5, 0x1f, 0x9a, 0x72, 0xf2, 0xe9,
	0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0x0d, 0xac, 0x7d, 0x0c, 0x60, 0x01, 0x00, 0x00,
}

//
// the following is the grpc-gateway rendering of the .proto file
//

var _ codes.Code
var _ io.Reader
var _ status.Status
var _ = runtime.String
var _ = utilities.NewDoubleArray

var (
	filter_Hello_SayHello_0 = &utilities.DoubleArray{Encoding: map[string]int{}, Base: []int(nil), Check: []int(nil)}
)

func request_Hello_SayHello_0(ctx context.Context, marshaler runtime.Marshaler, client HelloClient, req *http.Request, pathParams map[string]string) (proto.Message, runtime.ServerMetadata, error) {
	var protoReq HelloRequest
	var metadata runtime.ServerMetadata

	if err := runtime.PopulateQueryParameters(&protoReq, req.URL.Query(), filter_Hello_SayHello_0); err != nil {
		return nil, metadata, status.Errorf(codes.InvalidArgument, "%v", err)
	}

	msg, err := client.SayHello(ctx, &protoReq, grpc.Header(&metadata.HeaderMD), grpc.Trailer(&metadata.TrailerMD))
	return msg, metadata, err

}

// RegisterHelloHandlerFromEndpoint is same as RegisterHelloHandler but
// automatically dials to "endpoint" and closes the connection when "ctx" gets done.
func RegisterHelloHandlerFromEndpoint(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error) {
	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
			return
		}
		go func() {
			<-ctx.Done()
			if cerr := conn.Close(); cerr != nil {
				grpclog.Printf("Failed to close conn to %s: %v", endpoint, cerr)
			}
		}()
	}()

	return RegisterHelloHandler(ctx, mux, conn)
}

// RegisterHelloHandler registers the http handlers for service Hello to "mux".
// The handlers forward requests to the grpc endpoint over "conn".
func RegisterHelloHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
	return RegisterHelloHandlerClient(ctx, mux, NewHelloClient(conn))
}

// RegisterHelloHandler registers the http handlers for service Hello to "mux".
// The handlers forward requests to the grpc endpoint over the given implementation of "HelloClient".
// Note: the gRPC framework executes interceptors within the gRPC handler. If the passed in "HelloClient"
// doesn't go through the normal gRPC flow (creating a gRPC client etc.) then it will be up to the passed in
// "HelloClient" to call the correct interceptors.
func RegisterHelloHandlerClient(ctx context.Context, mux *runtime.ServeMux, client HelloClient) error {

	mux.Handle("GET", pattern_Hello_SayHello_0, func(w http.ResponseWriter, req *http.Request, pathParams map[string]string) {
		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()
		if cn, ok := w.(http.CloseNotifier); ok {
			go func(done <-chan struct{}, closed <-chan bool) {
				select {
				case <-done:
				case <-closed:
					cancel()
				}
			}(ctx.Done(), cn.CloseNotify())
		}
		inboundMarshaler, outboundMarshaler := runtime.MarshalerForRequest(mux, req)
		rctx, err := runtime.AnnotateContext(ctx, mux, req)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}
		resp, md, err := request_Hello_SayHello_0(rctx, inboundMarshaler, client, req, pathParams)
		ctx = runtime.NewServerMetadataContext(ctx, md)
		if err != nil {
			runtime.HTTPError(ctx, mux, outboundMarshaler, w, req, err)
			return
		}

		forward_Hello_SayHello_0(ctx, mux, outboundMarshaler, w, req, resp, mux.GetForwardResponseOptions()...)

	})

	return nil
}

var (
	pattern_Hello_SayHello_0 = runtime.MustPattern(runtime.NewPattern(1, []int{2, 0}, []string{"hello"}, ""))
)

var (
	forward_Hello_SayHello_0 = runtime.ForwardResponseMessage
)
