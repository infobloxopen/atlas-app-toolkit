package resource

import (
	"errors"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

type TestCodec struct{}

func (TestCodec) Decode(id *resourcepb.Identifier) (Identifier, error) {
	if id.ResourceType == "test_resource" {
		return nil, errors.New("test_resource error")
	}
	return nil, nil
}
func (TestCodec) Encode(Identifier) (*resourcepb.Identifier, error) {
	return nil, errors.New("test_resource error")

}

type TestMessage struct{}

func (TestMessage) XXX_MessageName() string { return "test_resource" }
func (TestMessage) Reset()                  {}
func (TestMessage) String() string          { return "test_resource" }
func (TestMessage) ProtoMessage()           {}

func init() {
	RegisterCodec(TestCodec{}, &TestMessage{})
}

func HandlePanic(t *testing.T, codec Codec, pb proto.Message) (err error) {
	t.Helper()

	defer func() {
		v, ok := recover().(string)
		if !ok {
			return
		}
		err = errors.New(v)
	}()

	RegisterCodec(codec, pb)
	return
}

func TestRegisterCodec(t *testing.T) {
	tcases := []struct {
		Codec         Codec
		Message       proto.Message
		ExpectedError string
	}{
		{
			Codec:         nil,
			Message:       nil,
			ExpectedError: "resource: register nil codec for resource " + defaultResource,
		},
		{
			Codec:         TestCodec{},
			Message:       &TestMessage{},
			ExpectedError: "resource: register codec called twice for resource " + TestMessage{}.XXX_MessageName(),
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
			Message:       &TestMessage{},
			ExpectedError: "test_resource error",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			Message:       nil,
			ExpectedError: "resource: codec is not registered for resource " + defaultResource,
		},
	}

	for _, tc := range tcases {
		_, err := Decode(tc.Message, tc.Identifier)
		if err != nil && err.Error() != tc.ExpectedError {
			t.Fatalf("invalid error %s, expected %s", err, tc.ExpectedError)
		}

	}
}

func TestEncode(t *testing.T) {
	tcases := []struct {
		Identifier    Identifier
		Message       proto.Message
		ExpectedError string
	}{
		{
			Identifier:    nil,
			Message:       &TestMessage{},
			ExpectedError: "test_resource error",
		},
		{
			Identifier:    nil,
			Message:       nil,
			ExpectedError: "resource: codec is not registered for resource " + defaultResource,
		},
	}

	for _, tc := range tcases {
		_, err := Encode(tc.Message, tc.Identifier)
		if err != nil && err.Error() != tc.ExpectedError {
			t.Fatalf("invalid error %s, expected %s", err, tc.ExpectedError)
		}

	}
}
