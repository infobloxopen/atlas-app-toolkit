package resource

import (
	"errors"
	"testing"

	"database/sql/driver"

	"strconv"

	"bytes"

	"github.com/golang/protobuf/proto"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

type TestCodec struct{}

func (TestCodec) Decode(id *resourcepb.Identifier) (driver.Value, error) {
	if id.GetResourceId() == "err" {
		return nil, errors.New("test error")
	}
	if id.GetResourceId() == "invalid" {
		return true, nil
	}
	if id.GetResourceId() == "str" {
		return "", nil
	}
	if id.GetResourceId() == "strempty" {
		return "", nil
	}
	if id.GetResourceId() == "12" {
		return strconv.ParseInt(id.GetResourceId(), 10, 64)
	}
	return id.ResourceId, nil
}
func (TestCodec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	switch value.(type) {
	case string:
		return &resourcepb.Identifier{ResourceId: value.(string)}, nil
	case int64:
		return &resourcepb.Identifier{ResourceId: strconv.FormatInt(value.(int64), 10)}, nil
	case []byte:
		return &resourcepb.Identifier{ResourceId: string(value.([]byte))}, nil
	}
	return nil, nil
}

type TestInt64Codec struct{}

func (TestInt64Codec) Decode(id *resourcepb.Identifier) (driver.Value, error) {
	if id.GetResourceId() == "err" {
		return nil, errors.New("test error")
	}
	if id.GetResourceId() == "invalid" {
		return true, nil
	}
	if id.GetResourceId() == "str" {
		return "", nil
	}
	return strconv.ParseInt(id.GetResourceId(), 10, 64)
}
func (TestInt64Codec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	return &resourcepb.Identifier{ResourceId: strconv.FormatInt(value.(int64), 10)}, nil
}

type TestBytesCodec struct{}

func (TestBytesCodec) Decode(id *resourcepb.Identifier) (driver.Value, error) {
	if id.GetResourceId() == "err" {
		return nil, errors.New("test error")
	}
	if id.GetResourceId() == "invalid" {
		return true, nil
	}
	if id.GetResourceId() == "strempty" {
		return "", nil
	}
	if id.GetResourceId() == "str" {
		return id.GetResourceId(), nil
	}
	return []byte(id.GetResourceId()), nil
}
func (TestBytesCodec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	return &resourcepb.Identifier{ResourceId: string(value.([]byte))}, nil
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

func Cleanup(t *testing.T) {
	t.Helper()
	// cleanup
	registry = make(map[string]Codec)
	appname = ""
	asEmpty = false
	asPlural = false
}

func TestRegisterCodec(t *testing.T) {
	RegisterCodec(TestCodec{}, &TestProtoMessage{})
	defer Cleanup(t)

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
}

func TestRegisterApplication(t *testing.T) {
	defer Cleanup(t)
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
}

func TestDecode(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")
	defer Cleanup(t)

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
		{
			Identifier: nil,
			Message:    nil,
			Value:      nil,
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
}

func TestDecodeInt64(t *testing.T) {
	RegisterCodec(&TestInt64Codec{}, &TestProtoMessage{})
	RegisterApplication("app")
	defer Cleanup(t)

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
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "err",
			},
			Message:       &TestProtoMessage{},
			Value:         0,
			ExpectedError: "test error",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "invalid",
			},
			Message:       &TestProtoMessage{},
			Value:         0,
			ExpectedError: "resource: invalid value type, expected int64",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "str",
			},
			Message: &TestProtoMessage{},
			Value:   0,
		},
		{
			Identifier: &resourcepb.Identifier{
				ResourceId: "s",
			},
			Message:       &resourcepb.Identifier{}, // any proto not registered
			ExpectedError: "resource: invalid value type, expected int64",
		},
		{
			Identifier: nil,
			Message:    nil,
			Value:      0,
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
}

func TestDecodeBytes(t *testing.T) {
	RegisterCodec(&TestBytesCodec{}, &TestProtoMessage{})
	RegisterApplication("app")
	defer Cleanup(t)

	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Message       proto.Message
		Value         []byte
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "1",
			},
			Message: &TestProtoMessage{},
			Value:   []byte("1"),
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "err",
			},
			Message:       &TestProtoMessage{},
			Value:         nil,
			ExpectedError: "test error",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "invalid",
			},
			Message:       &TestProtoMessage{},
			Value:         nil,
			ExpectedError: "resource: invalid value type, expected []byte",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "str",
			},
			Message: &TestProtoMessage{},
			Value:   []byte("str"),
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "strempty",
			},
			Message: &TestProtoMessage{},
			Value:   nil,
		},
		{
			Identifier: nil,
			Message:    nil,
			Value:      nil,
		},
	}

	for n, tc := range tcases {
		v, err := DecodeBytes(tc.Message, tc.Identifier)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if !bytes.Equal(v, tc.Value) {
			t.Errorf("tc %d: invalid value %v, expected %v", n, v, tc.Value)
		}

	}
}

func TestEncode(t *testing.T) {
	RegisterCodec(&TestCodec{}, &TestProtoMessage{})
	RegisterApplication("app")
	defer Cleanup(t)

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
				ResourceType:    "identifier",
				ResourceId:      "1",
			},
		},
		{
			Value:   int64(12),
			Message: &TestProtoMessage{},
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "12",
			},
		},
		{
			Value:   int64(1),
			Message: nil,
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "",
				ResourceId:      "1",
			},
		},
		{
			Value:   []byte("1"),
			Message: &TestProtoMessage{},
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "1",
			},
		},
		{
			Value:   []byte("1"),
			Message: nil,
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "",
				ResourceId:      "1",
			},
		},
		{
			Value:      nil,
			Message:    nil,
			Identifier: nil,
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			Message: nil,
			Value:   "",
		},
	}

	for n, tc := range tcases {
		id, err := Encode(tc.Message, tc.Value)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Fatalf("tc %d: invalid error %s, expected %s", n, err, tc.ExpectedError)
		}
		if v := id.GetApplicationName(); v != tc.Identifier.GetApplicationName() {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.Identifier.ApplicationName)
		}
		if v := id.GetResourceType(); v != tc.Identifier.GetResourceType() {
			t.Errorf("tc %d: nvalid resource type %s, expected %s", n, v, tc.Identifier.ResourceType)
		}
		if v := id.GetResourceId(); v != tc.Identifier.GetResourceId() {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.Identifier.ResourceId)
		}
	}
}

type NamerMessage string

func (m NamerMessage) ResourceName() string { return string(m) }
func (NamerMessage) ProtoMessage()          {}
func (NamerMessage) Reset()                 {}
func (NamerMessage) String() string         { return "" }

func TestName(t *testing.T) {
	tcases := []struct {
		Message      proto.Message
		ExpectedName string
	}{
		{
			Message:      NamerMessage("books"),
			ExpectedName: "books",
		},
		{
			Message:      TestProtoMessage{},
			ExpectedName: "test_proto_message",
		},
	}
	for n, tc := range tcases {
		if name := Name(tc.Message); name != tc.ExpectedName {
			t.Errorf("tc %d: invalid resource name %s, expected %s", n, name, tc.ExpectedName)
		}
	}
}

func TestNamePlural(t *testing.T) {
	SetPlural()
	defer Cleanup(t)

	tcases := []struct {
		Message      proto.Message
		ExpectedName string
	}{
		{
			Message:      NamerMessage("books"),
			ExpectedName: "books",
		},
		{
			Message:      TestProtoMessage{},
			ExpectedName: "test_proto_messages",
		},
	}
	for n, tc := range tcases {
		if name := Name(tc.Message); name != tc.ExpectedName {
			t.Errorf("tc %d: invalid resource name %s, expected %s", n, name, tc.ExpectedName)
		}
	}
}
