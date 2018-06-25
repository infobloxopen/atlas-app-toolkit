package resource

import (
	"errors"
	"testing"

	"database/sql/driver"

	"github.com/golang/protobuf/proto"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

type TestCodec struct{}

func (TestCodec) Decode(id *resourcepb.Identifier) (driver.Value, error) {
	return id.ResourceId, nil
}
func (TestCodec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	return &resourcepb.Identifier{ResourceId: value.(string)}, nil

}

type TestProtoMessage struct{}

func (TestProtoMessage) XXX_MessageName() string { return "TestProtoMessage" }
func (TestProtoMessage) Reset()                  {}
func (TestProtoMessage) String() string          { return "TestProtoMessage" }
func (TestProtoMessage) ProtoMessage()           {}

func HandlePanic(t *testing.T, fn func()) (err error) {
	t.Helper()

	defer func() {
		v, ok := recover().(string)
		if !ok {
			return
		}
		err = errors.New(v)
	}()
	fn()
	return
}

func TestRegisterCodec(t *testing.T) {
	RegisterCodec(TestCodec{}, &TestProtoMessage{})

	tcases := []struct {
		Fn            func()
		ExpectedError string
	}{
		{
			Fn:            func() { RegisterCodec(nil, nil) },
			ExpectedError: "resource: register nil codec for resource <default>",
		},
		{
			Fn: func() { RegisterCodec(TestCodec{}, nil) },
		},
		{
			Fn:            func() { RegisterCodec(TestCodec{}, &TestProtoMessage{}) },
			ExpectedError: "resource: register codec called twice for resource " + TestProtoMessage{}.XXX_MessageName(),
		},
	}

	for n, tc := range tcases {
		err := HandlePanic(t, tc.Fn)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
	}
	// cleanup
	registry = make(map[string]Codec)
}

func TestRegisterApplication(t *testing.T) {
	tcases := []struct {
		Fn            func()
		AppName       string
		ExpectedError string
	}{
		{
			Fn:      func() { RegisterApplication("app") },
			AppName: "app",
		},
		{
			Fn:            func() { RegisterApplication("app1") },
			ExpectedError: "resource: application name already registered",
			AppName:       "app",
		},
	}

	for n, tc := range tcases {
		err := HandlePanic(t, tc.Fn)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v := ApplicationName(); v != tc.AppName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.AppName)
		}
	}
	// cleanup
	appname = ""
}

func TestDecode(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")

	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Message       proto.Message
		Value         driver.Value
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "1",
			},
			Message: &TestProtoMessage{},
			Value:   "1",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "1",
			},
			Message: nil,
			Value:   "app/res/1",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "1",
			},
			Message: &resourcepb.Identifier{}, // any proto not registered
			Value:   "1",
		},
	}

	for n, tc := range tcases {
		v, err := Decode(tc.Message, tc.Identifier)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v != tc.Value {
			t.Errorf("tc %d: invalid value %v, expected %v", n, v, tc.Value)
		}

	}

	// cleanup
	registry = make(map[string]Codec)
	appname = ""
}

func TestDecodeInt64(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")

	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Message       proto.Message
		Value         int64
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "1",
			},
			Message: &TestProtoMessage{},
			Value:   1,
		},
		{
			Identifier: &resourcepb.Identifier{
				ResourceId: "s",
			},
			Message:       &resourcepb.Identifier{}, // any proto not registered
			ExpectedError: "resource: invalid value type, expected int64",
		},
	}

	for n, tc := range tcases {
		v, err := DecodeInt64(tc.Message, tc.Identifier)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v != tc.Value {
			t.Errorf("tc %d: invalid value %v, expected %v", n, v, tc.Value)
		}

	}

	// cleanup
	registry = make(map[string]Codec)
	appname = ""
}

func TestEncode(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")

	tcases := []struct {
		Value         driver.Value
		Message       proto.Message
		Identifier    *resourcepb.Identifier
		ExpectedError string
	}{
		{
			Value:   "1",
			Message: &TestProtoMessage{},
			Identifier: &resourcepb.Identifier{
				ResourceId: "1",
			},
		},
		{
			Value:   "app/res/1",
			Message: nil,
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "1",
			},
		},
		{
			Value:   "1",
			Message: &resourcepb.Identifier{},
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "identifiers",
				ResourceId:      "1",
			},
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			Message:       nil,
			Value:         12,
			ExpectedError: "resource: invalid value type int, expected string",
		},
	}

	for n, tc := range tcases {
		id, err := Encode(tc.Message, tc.Value)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v := id.GetApplicationName(); v != tc.Identifier.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.Identifier.ApplicationName)
		}
		if v := id.GetResourceType(); v != tc.Identifier.ResourceType {
			t.Errorf("tc %d: nvalid resource type %s, expected %s", n, v, tc.Identifier.ResourceType)
		}
		if v := id.GetResourceId(); v != tc.Identifier.ResourceId {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.Identifier.ResourceId)
		}
	}

	// cleanup
	registry = make(map[string]Codec)
	appname = ""
}

func TestEncodeInt64(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")

	tcases := []struct {
		Value         int64
		Message       proto.Message
		Identifier    *resourcepb.Identifier
		ExpectedError string
	}{
		{
			Value:   1,
			Message: nil,
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "",
				ResourceId:      "1",
			},
		},
	}

	for n, tc := range tcases {
		id, err := EncodeInt64(tc.Message, tc.Value)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v := id.GetApplicationName(); v != tc.Identifier.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.Identifier.ApplicationName)
		}
		if v := id.GetResourceType(); v != tc.Identifier.ResourceType {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, v, tc.Identifier.ResourceType)
		}
		if v := id.GetResourceId(); v != tc.Identifier.ResourceId {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.Identifier.ResourceId)
		}
	}

	// cleanup
	registry = make(map[string]Codec)
	appname = ""
}
