package resource

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

const defaultResource = "<default>"

var (
	mu       sync.RWMutex
	registry = make(map[string]Codec)
	appname  string
)

// Codec defines the interface package uses to encode and decode Protocol Buffer
// Identifier to the driver.Value.
// Note that implementation must be thread safe.
type Codec interface {
	// Encode encodes value to Protocol Buffer representation
	Encode(driver.Value) (*resourcepb.Identifier, error)
	// Decode decodes Protocol Buffer representation to the driver.Value
	Decode(*resourcepb.Identifier) (driver.Value, error)
}

// RegisterApplication registers name of the application.
// Registered name is used by Encode to populate application name of
// Protocol Buffer Identifier.
func RegisterApplication(name string) {
	mu.Lock()
	defer mu.Unlock()
	if appname != "" {
		panic("resource: application name already registered")
	}
	appname = name
}

// RegisterCodec registers codec for a given pb.
// If pb is nil the codec is registered as default.
// If codec is nil or registered twice for the same resource
// the panic is raised.
func RegisterCodec(codec Codec, pb proto.Message) {
	mu.Lock()
	defer mu.Unlock()

	var name string
	if pb == nil {
		name = defaultResource
	} else {
		name = proto.MessageName(pb)
	}

	if codec == nil {
		panic("resource: register nil codec for resource " + name)
	}

	_, ok := registry[name]
	if ok {
		panic("resource: register codec called twice for resource " + name)
	}
	registry[name] = codec
}

// Decode decodes identifier using a codec registered for pb if found.
//
// If codec is not found
// - and id is nil, the (nil, nil) are returned.
// - and pb is nil the id is decoded in a fully qualified string value in format specified for Atlas References.
// - only Resource ID part of the identifier is returned as string value.
func Decode(pb proto.Message, id *resourcepb.Identifier) (driver.Value, error) {
	if c, ok := lookupCodec(pb); ok {
		return c.Decode(id)
	}

	if id == nil {
		return nil, nil
	}

	// fully qualified
	if pb == nil {
		return resourcepb.BuildString(id.GetApplicationName(), id.GetResourceType(), id.GetResourceId()), nil
	}

	// resource id
	return id.GetResourceId(), nil
}

// DecodeInt64 decodes value returned by Decode as int64.
// Returns an error if value is not of int64 type.
func DecodeInt64(pb proto.Message, id *resourcepb.Identifier) (int64, error) {
	v, err := Decode(pb, id)
	if err != nil {
		return 0, err
	}
	if v == nil {
		return 0, nil
	}
	i, ok := v.(int64)
	if ok {
		return i, nil
	}
	s, ok := v.(string)
	if !ok {
		return 0, fmt.Errorf("resource: invalid value type, expected int64")
	}
	if s == "" {
		return 0, nil
	}

	i, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("resource: invalid value type, expected int64")
	}

	return i, nil
}

// DecodeBytes decodes value returned by Decode as []byte.
// Returns an error if value is not of []byte type.
func DecodeBytes(pb proto.Message, id *resourcepb.Identifier) ([]byte, error) {
	v, err := Decode(pb, id)
	if err != nil {
		return nil, err
	}
	if v == nil {
		return nil, nil
	}
	b, ok := v.([]byte)
	if ok {
		return b, nil
	}
	s, ok := v.(string)
	if !ok {
		return nil, fmt.Errorf("resource: invalid value type, expected []byte")
	}
	if s == "" {
		return nil, nil
	}
	return []byte(s), nil
}

// Encode encodes identifier using a codec registered for pb.
//
// If codec is not found
// - and value is not of string type an error is returned.
// - and pb is nil the id is encoded as it would be a string value in fully qualified format
// - and value is nil the (nil, nil) are returned
//
// If Resource ID part is not empty, the Application Name and Resource Type parts
// are populated by ApplicationName and Name functions accordingly, otherwise
// the empty identifier is returned.
func Encode(pb proto.Message, value driver.Value) (*resourcepb.Identifier, error) {
	if c, ok := lookupCodec(pb); ok {
		return c.Encode(value)
	}
	if value == nil {
		return nil, nil
	}

	var id resourcepb.Identifier
	s, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("resource: invalid value type %T, expected string", value)
	}
	if s == "" {
		return &id, nil
	}
	if pb == nil {
		id.ApplicationName, id.ResourceType, id.ResourceId = resourcepb.ParseString(s)
	}

	if id.ApplicationName == "" {
		id.ApplicationName = appname
	}
	if id.ResourceType == "" {
		id.ResourceType = Name(pb)
	}
	if id.ResourceId == "" {
		id.ResourceId = s
	}

	return &id, nil
}

// EncodeInt64 converts value to string and forwards call to Encode.
// Returns an error if value is not of int64 type.
func EncodeInt64(pb proto.Message, value driver.Value) (*resourcepb.Identifier, error) {
	if c, ok := lookupCodec(pb); ok {
		return c.Encode(value)
	}

	v, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("resource: invalid value type %T, expected int64", value)
	}
	return Encode(pb, fmt.Sprintf("%d", v))
}

// EncodeBytes converts value to string and forwards call to Encode.
// Returns an error if value is not of []byte type.
func EncodeBytes(pb proto.Message, value driver.Value) (*resourcepb.Identifier, error) {
	if c, ok := lookupCodec(pb); ok {
		return c.Encode(value)
	}

	v, ok := value.([]byte)
	if !ok {
		return nil, fmt.Errorf("resource: invalid value type %T, expected []byte", value)
	}
	return Encode(pb, string(v))
}

// Name returns name of pb.
// If pb implements XXX_MessageName then it is used to return name, otherwise
// proto.MessageName is used and "s" symbol is added at the end of the message name.
func Name(pb proto.Message) string {
	if pb == nil {
		return ""
	}
	type xname interface {
		XXX_MessageName() string
	}
	var (
		name string
		udef bool
	)
	if m, ok := pb.(xname); ok {
		name, udef = m.XXX_MessageName(), true
	}
	name = proto.MessageName(pb)

	v := strings.Split(name, ".")
	name = strings.ToLower(v[len(v)-1])
	if !udef {
		name += "s"
	}
	return name
}

// ApplicationName returns application name registered by RegisterApplication.
func ApplicationName() string {
	mu.RLock()
	defer mu.RUnlock()
	return appname
}

func lookupCodec(pb proto.Message) (Codec, bool) {
	mu.RLock()
	defer mu.RUnlock()

	var name string
	if pb == nil {
		name = defaultResource
	} else {
		name = proto.MessageName(pb)
	}
	codec, ok := registry[name]
	if !ok || codec == nil {
		return nil, false
	}
	return codec, true
}
