package resource_test

import (
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

type TestCodec struct{}

func (TestCodec) Decode(id *resourcepb.Identifier) (resource.Identifier, error) {
	if id.ResourceType == "test_resource" {
		return nil, errors.New("test_resource error")
	}
	return nil, nil
}
func (TestCodec) Encode(resource.Identifier) (*resourcepb.Identifier, error) {
	return nil, errors.New("test_resource error")

}

type TestProtoMessage struct{}

func (TestProtoMessage) XXX_MessageName() string { return "TestProtoMessage" }
func (TestProtoMessage) Reset()                  {}
func (TestProtoMessage) String() string          { return "TestProtoMessage" }
func (TestProtoMessage) ProtoMessage()           {}

func HandlePanic(t *testing.T, codec resource.Codec, pb proto.Message) (err error) {
	t.Helper()

	defer func() {
		v, ok := recover().(string)
		if !ok {
			return
		}
		err = errors.New(v)
	}()

	resource.RegisterCodec(codec, pb)
	return
}

func TestRegisterCodec(t *testing.T) {
	resource.RegisterCodec(TestCodec{}, &TestProtoMessage{})

	tcases := []struct {
		Codec         resource.Codec
		Message       proto.Message
		ExpectedError string
	}{
		{
			Codec:         nil,
			Message:       nil,
			ExpectedError: "resource: register nil codec for resource <default>",
		},
		{
			Codec:         TestCodec{},
			Message:       &TestProtoMessage{},
			ExpectedError: "resource: register codec called twice for resource " + TestProtoMessage{}.XXX_MessageName(),
		},
	}

	for _, tc := range tcases {
		if err := HandlePanic(t, tc.Codec, tc.Message); err != nil && err.Error() != tc.ExpectedError {
			t.Errorf("invalid error %s, expected %s", err, tc.ExpectedError)
		}
	}
}

func TestDecode(t *testing.T) {
	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Message       proto.Message
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "test_resource",
				ResourceId:      "",
			},
			Message:       &TestProtoMessage{},
			ExpectedError: "test_resource error",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			Message:       nil,
			ExpectedError: "resource: codec is not registered for resource <default>",
		},
	}

	for _, tc := range tcases {
		_, err := resource.Decode(tc.Message, tc.Identifier)
		if err != nil && err.Error() != tc.ExpectedError {
			t.Fatalf("invalid error %s, expected %s", err, tc.ExpectedError)
		}

	}
}

func TestEncode(t *testing.T) {
	tcases := []struct {
		Identifier    resource.Identifier
		Message       proto.Message
		ExpectedError string
	}{
		{
			Identifier:    nil,
			Message:       &TestProtoMessage{},
			ExpectedError: "test_resource error",
		},
		{
			Identifier:    nil,
			Message:       nil,
			ExpectedError: "resource: codec is not registered for resource <default>",
		},
	}

	for _, tc := range tcases {
		_, err := resource.Encode(tc.Message, tc.Identifier)
		if err != nil && err.Error() != tc.ExpectedError {
			t.Fatalf("invalid error %s, expected %s", err, tc.ExpectedError)
		}

	}
}
