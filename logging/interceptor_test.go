package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	grpc_logrus "github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/logrus/ctxlogrus"
	"github.com/grpc-ecosystem/go-grpc-middleware/util/metautils"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/infobloxopen/atlas-app-toolkit/logging/mocks"
	"github.com/infobloxopen/atlas-app-toolkit/requestid"
)

const (
	testJWT                 = `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWJqZWN0Ijp7ImlkIjoidGVzdElEIiwic3ViamVjdF90eXBlIjoidGVzdFVzZXIiLCJhdXRoZW50aWNhdGlvbl90eXBlIjoidGVzdCJ9LCJhY2NvdW50X2lkIjoidGVzdC1hY2MtaWQiLCJjdXN0b21fZmllbGQiOiJ0ZXN0LWN1c3RvbS1maWVsZCJ9.pEuJadBkY_twamJid9GKHGZWtIHsZ3cXv84sRqPG-vw`
	testAuthorizationHeader = "authorization"
	testCustomHeaderKey     = "custom_header"
	testCustomHeaderVal     = "test-custom-header"
	testCustomJWTFieldKey   = "custom_field"
	testCustomJWTFieldVal   = "test-custom-field"
	testAccID               = "test-acc-id"
	testRequestID           = "test-request-id"
	testMethod              = "TestMethod"
	testFullMethod          = "/app.Object/TestMethod"
)

var (
	buf         bytes.Buffer
	reader      io.Reader
	testLogger  = New("Info")
	testMD      = metautils.NiceMD{}.Set(testAuthorizationHeader, testJWT).Set(requestid.DefaultRequestIDKey, testRequestID).Set(testCustomHeaderKey, testCustomHeaderVal)
	testSubject = map[string]interface{}{"id": "testID", "subject_type": "testUser", "authentication_type": "test"}
)

func TestNewLoggerFields(t *testing.T) {
	startTime := time.Now()
	expected := logrus.Fields{
		grpc_logrus.SystemField: "grpc",
		grpc_logrus.KindField:   "server",
		DefaultGRPCServiceKey:   "app.Object",
		DefaultGRPCMethodKey:    "TestMethod",
		DefaultGRPCStartTimeKey: startTime.Format(time.RFC3339Nano),
	}

	result := newLoggerFields(testFullMethod, startTime, "server")
	assert.Equal(t, expected, result)
}

func TestUnaryClientInterceptor(t *testing.T) {
	testLogger.Out = &buf
	interceptor := UnaryClientInterceptor(logrus.NewEntry(testLogger))

	// FIXME: clients have OutgoingContext, so the mock should too
	// https://github.com/infobloxopen/atlas-app-toolkit/issues/191
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	invokerMock := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		newMD, ok := metadata.FromIncomingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testJWT, newMD.Get(testAuthorizationHeader)[0])

		reqID, ok := requestid.FromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testRequestID, reqID)

		assert.Equal(t, testMethod, method)

		return nil
	}

	err := interceptor(ctx, testMethod, nil, nil, nil, invokerMock)
	assert.NoError(t, err)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
	assert.Equal(t, testMethod, result[DefaultGRPCMethodKey])
	assert.Equal(t, "finished unary call with code OK", result["msg"])
}

func TestUnaryClientInterceptor_Failed(t *testing.T) {
	testLogger.Out = &buf
	interceptor := UnaryClientInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	invokerMock := func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, opts ...grpc.CallOption) error {
		return fmt.Errorf("Error completing RPC")
	}

	err := interceptor(ctx, testMethod, nil, nil, nil, invokerMock)
	assert.Error(t, err)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Error completing RPC", result[logrus.ErrorKey])
}

func TestStreamClientInterceptor(t *testing.T) {
	testLogger.Out = &buf
	interceptor := StreamClientInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	streamerMock := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		newMD, ok := metadata.FromIncomingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testJWT, newMD.Get(testAuthorizationHeader)[0])

		reqID, ok := requestid.FromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testRequestID, reqID)

		assert.Equal(t, testMethod, method)

		return nil, nil
	}

	cs, err := interceptor(ctx, nil, nil, testMethod, streamerMock)
	assert.NoError(t, err)
	assert.Nil(t, cs)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
	assert.Equal(t, testMethod, result[DefaultGRPCMethodKey])
	assert.Equal(t, "finished client streaming call with code OK", result["msg"])
}

func TestStreamClientInterceptor_Failed(t *testing.T) {
	testLogger.Out = &buf
	interceptor := StreamClientInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	streamerMock := func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		return nil, fmt.Errorf("Stream rpc error")
	}

	cs, err := interceptor(ctx, nil, nil, testMethod, streamerMock)
	assert.Error(t, err)
	assert.Nil(t, cs)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Stream rpc error", result[logrus.ErrorKey])
}

func TestUnaryServerInterceptor(t *testing.T) {
	testLogger.Out = &buf
	interceptor := UnaryServerInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	handlerMock := func(ctx context.Context, req interface{}) (interface{}, error) {
		newMD, ok := metadata.FromIncomingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testJWT, newMD.Get(testAuthorizationHeader)[0])

		reqID, ok := requestid.FromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testRequestID, reqID)

		entry := ctxlogrus.Extract(ctx)
		assert.Equal(t, testRequestID, entry.Data[DefaultRequestIDKey])
		assert.Equal(t, testAccID, entry.Data[DefaultAccountIDKey])
		assert.Equal(t, testSubject, entry.Data[DefaultSubjectKey])

		return nil, nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: testFullMethod}, handlerMock)
	assert.NoError(t, err)
	assert.Nil(t, resp)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
	assert.Equal(t, "app.Object", result[DefaultGRPCServiceKey])
	assert.Equal(t, testMethod, result[DefaultGRPCMethodKey])
	assert.Equal(t, "finished unary call with code OK", result["msg"])
}

func TestUnaryServerInterceptorDeprecatedHeader(t *testing.T) {
	testLogger.Out = &buf
	interceptor := UnaryServerInterceptor(logrus.NewEntry(testLogger))

	md := metautils.NiceMD{}.Set(testAuthorizationHeader, testJWT).Set(requestid.DeprecatedRequestIDKey, testRequestID).Set(testCustomHeaderKey, testCustomHeaderVal)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(md))

	handlerMock := func(ctx context.Context, req interface{}) (interface{}, error) {
		newMD, ok := metadata.FromIncomingContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testJWT, newMD.Get(testAuthorizationHeader)[0])

		reqID, ok := requestid.FromContext(ctx)
		assert.True(t, ok)
		assert.Equal(t, testRequestID, reqID)

		entry := ctxlogrus.Extract(ctx)
		assert.Equal(t, testRequestID, entry.Data[DefaultRequestIDKey])
		assert.Equal(t, testAccID, entry.Data[DefaultAccountIDKey])
		assert.Equal(t, testSubject, entry.Data[DefaultSubjectKey])

		return nil, nil
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: testFullMethod}, handlerMock)
	assert.NoError(t, err)
	assert.Nil(t, resp)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
	assert.Equal(t, "app.Object", result[DefaultGRPCServiceKey])
	assert.Equal(t, testMethod, result[DefaultGRPCMethodKey])
	assert.Equal(t, "finished unary call with code OK", result["msg"])
}

func TestUnaryServerInterceptor_Failed(t *testing.T) {
	testLogger.Out = &buf
	interceptor := UnaryServerInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	handlerMock := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, fmt.Errorf("Server handler error")
	}

	resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{FullMethod: testFullMethod}, handlerMock)
	assert.Error(t, err)
	assert.Nil(t, resp)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Server handler error", result[logrus.ErrorKey])
}

func TestStreamServerInterceptor(t *testing.T) {
	testLogger.Out = &buf
	interceptor := StreamServerInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	handlerMock := func(srv interface{}, stream grpc.ServerStream) error {
		newMD, ok := metadata.FromIncomingContext(stream.Context())
		assert.True(t, ok)
		assert.Equal(t, testJWT, newMD.Get(testAuthorizationHeader)[0])

		reqID, ok := requestid.FromContext(stream.Context())
		assert.True(t, ok)
		assert.Equal(t, testRequestID, reqID)

		entry := ctxlogrus.Extract(stream.Context())
		assert.Equal(t, testRequestID, entry.Data[DefaultRequestIDKey])
		assert.Equal(t, testAccID, entry.Data[DefaultAccountIDKey])
		assert.Equal(t, testSubject, entry.Data[DefaultSubjectKey])

		return nil
	}

	stream := &mocks.ServerStreamMock{}
	stream.ContextReturns(ctx)
	err := interceptor(ctx, stream, &grpc.StreamServerInfo{FullMethod: testFullMethod}, handlerMock)
	assert.NoError(t, err)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
	assert.Equal(t, "app.Object", result[DefaultGRPCServiceKey])
	assert.Equal(t, testMethod, result[DefaultGRPCMethodKey])
	assert.Equal(t, "finished streaming call with code OK", result["msg"])
}

func TestStreamServerInterceptor_Failed(t *testing.T) {
	testLogger.Out = &buf
	interceptor := StreamServerInterceptor(logrus.NewEntry(testLogger))

	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	handlerMock := func(srv interface{}, stream grpc.ServerStream) error {
		return fmt.Errorf("Stream handler error")
	}

	stream := &mocks.ServerStreamMock{}
	stream.ContextReturns(ctx)
	err := interceptor(ctx, stream, &grpc.StreamServerInfo{FullMethod: testFullMethod}, handlerMock)
	assert.Error(t, err)

	reader = &buf
	bts, err := ioutil.ReadAll(reader)
	assert.NoError(t, err)

	result := map[string]interface{}{}

	err = json.Unmarshal(bts, &result)
	assert.NoError(t, err)
	assert.Equal(t, "Stream handler error", result[logrus.ErrorKey])
}

func TestAddRequestIDField(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	result := logrus.Fields{}
	err := addRequestIDField(ctx, result)
	assert.NoError(t, err)
	assert.Equal(t, testRequestID, result[DefaultRequestIDKey])
}

func TestAddRequestIDField_Failed(t *testing.T) {
	ctx := context.Background()

	err := addRequestIDField(ctx, logrus.Fields{})
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("Unable to get %q from context", DefaultRequestIDKey), err.Error())
}

func TestAddAccountIDField(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	result := logrus.Fields{}
	err := addAccountIDField(ctx, result)
	assert.NoError(t, err)
	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
}

func TestAddAccountID_Failed(t *testing.T) {
	ctx := context.Background()

	result := logrus.Fields{}
	err := addAccountIDField(ctx, result)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("Unable to get %q from context", DefaultAccountIDKey), err.Error())
	assert.Equal(t, valueUndefined, result[DefaultAccountIDKey])
}

func TestAddCustomField(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	result := logrus.Fields{}
	err := addCustomField(ctx, result, testCustomJWTFieldKey)
	assert.NoError(t, err)
	assert.Equal(t, testCustomJWTFieldVal, result[testCustomJWTFieldKey])
}

func TestAddCustomField_SubjectNotAMap(t *testing.T) {
	withSingleSubject := `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWJqZWN0IjoidGVzdC11c2VyIn0.WwqjPgnri4ArIv4vo5qMFwqTCvxYLlE1AYfD3HBP-v4`
	md := metautils.NiceMD{}.Set(testAuthorizationHeader, withSingleSubject)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(md))

	result := logrus.Fields{}
	err := addCustomField(ctx, result, DefaultSubjectKey)
	assert.NoError(t, err)
	assert.Equal(t, "test-user", result[DefaultSubjectKey])
}

func TestAddCustomField_Failed(t *testing.T) {
	ctx := context.Background()

	err := addCustomField(ctx, logrus.Fields{}, "test")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("Unable to get custom %q field from context", "test"), err.Error())
}

func TestAddCustomField_WithS2SToken(t *testing.T) {
	withSingleSubject := `Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJzZXJ2aWNlIjoiYWxsIiwic3ViamVjdCI6eyJpZCI6InNlcnZpY2UuYWxsLjEwMDAwMDAwIiwic3ViamVjdF90eXBlIjoiIiwiYXV0aGVudGljYXRpb25fdHlwZSI6IiJ9LCJhdWQiOiJpYi1zdGsiLCJleHAiOjE2MzQ3MDAwMDAwLCJqdGkiOiJzZXJ2aWNlLmFsbC4xMDAwMDAwMCIsImlhdCI6MTYwMzEwMDAwMDAsImlzcyI6Inh4eHh4LWF1dGhuLXN2YyIsIm5iZiI6MTYwMzE2NjUwN30.e2BI0x6EFONKT7Kqcsro9BHFtu0QA1BOcWB0SjiDl5D3UDg5WOd--xk6liwubVk-aZGYQLpnqqaM87pVIhS5tQ`

	md := metautils.NiceMD{}.Set(testAuthorizationHeader, withSingleSubject)
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(md))

	result := logrus.Fields{}
	// should ignore svc token if DefaultSubjectKey is requested
	err := addCustomField(ctx, result, DefaultSubjectKey)
	assert.NoError(t, err)

	// should return an error if other key is requested
	err = addCustomField(ctx, result, "other_key")
	assert.Error(t, err)
}

func TestAddHeaderField(t *testing.T) {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))

	result := logrus.Fields{}
	err := addHeaderField(ctx, result, testCustomHeaderKey)
	assert.NoError(t, err)
	assert.Equal(t, testCustomHeaderVal, result[testCustomHeaderKey])
}

func TestAddHeaderField_Failed(t *testing.T) {
	ctx := context.Background()

	err := addHeaderField(ctx, logrus.Fields{}, "test")
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("Unable to get custom header %q from context", "test"), err.Error())
}

func TestSetClientInterceptorFields(t *testing.T) {
	opts := []Option{
		WithCustomFields(testFields),
		WithCustomHeaders(testHeaders),
	}

	result := logrus.Fields{}
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD(testMD))
	setClientInterceptorFields(ctx, result, testLogger, initOptions(opts), time.Now())

	assert.Equal(t, testAccID, result[DefaultAccountIDKey])
	assert.Equal(t, testCustomJWTFieldVal, result[testCustomJWTFieldKey])
	assert.Equal(t, testCustomHeaderVal, result[testCustomHeaderKey])
	assert.Equal(t, testSubject, result[DefaultSubjectKey])
}
