package resource

import (
	"database/sql/driver"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/golang/protobuf/proto"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/util"
)

const defaultResource = "<default>"

var (
	mu       sync.RWMutex
	registry = make(map[string]Codec)
	appname  string
	asEmpty  bool
	asPlural bool
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

// Namer is the interface that names resource.
type Namer interface {
	// ResourceName returns the name of a resource.
	ResourceName() string
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

// SetReturnEmpty sets package flag that indicates all nil values of driver.Value
// type in codecs must be converted to empty instance of Identifier.
// Default value is false.
func SetReturnEmpty() {
	mu.Lock()
	defer mu.Unlock()
	asEmpty = true
}

// ReturnEmpty returns flag that indicates all nil values of driver.Value type
// in codecs must be converted to empty instance of Identifier.
func ReturnEmpty() bool {
	mu.RLock()
	defer mu.RUnlock()
	return asEmpty
}

// SetPlural sets package flag that instructs resource.Name to
// return name in plural form by adding 's' at the end of the name.
func SetPlural() {
	mu.Lock()
	defer mu.Unlock()
	asPlural = true
}

// Plural returns true if resource.SetPlural was called,
// otherwise returns false.
func Plural() bool {
	mu.RLock()
	defer mu.RUnlock()
	return asPlural
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

	if resourcepb.Nil(id) {
		return nil, nil
	}

	// fully qualified
	if pb == nil {
		return resourcepb.BuildString(id.GetApplicationName(), id.GetResourceType(), id.GetResourceId()), nil
	}

	appName := ApplicationName()
	resourceName := Name(pb)

	if id.ApplicationName != appName && id.ApplicationName != "" {
		return 0, fmt.Errorf("resource: invalid application name - %s, expected %s", id.ApplicationName, appName)
	}

	if id.ResourceType != resourceName && id.ResourceType != "" {
		return 0, fmt.Errorf("resource: invalid resource name - %s, expected %s", id.ResourceType, resourceName)
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
	var id resourcepb.Identifier

	if c, ok := lookupCodec(pb); ok {
		return c.Encode(value)
	}
	if value == nil {
		if ReturnEmpty() {
			return &id, nil
		}
		return nil, nil
	}
	var sval string
	switch v := value.(type) {
	case []byte:
		sval = string(v)
	case int64:
		sval = fmt.Sprintf("%d", v)
	case string:
		sval = v
	default:
		return nil, fmt.Errorf("resource: unsupported value type %T", value)
	}

	if sval == "" {
		return &id, nil
	}
	if pb == nil {
		id.ApplicationName, id.ResourceType, id.ResourceId = resourcepb.ParseString(sval)
	}

	if id.ApplicationName == "" {
		id.ApplicationName = appname
	}
	if id.ResourceType == "" {
		id.ResourceType = Name(pb)
	}
	if id.ResourceId == "" {
		id.ResourceId = sval
	}

	return &id, nil
}

// Name returns name of pb.
// If pb implements Namer interface it is used to return name,
// otherwise the the proto.MessageName is used to obtain fully qualified resource name.
//
// The only last part of fully qualified resource name is used and converted to lower case.
// E.g. infoblox.rpc.Identifier -> identifier
//
// If SetPlural is called the 's' symbol is added at the end of resource name.
func Name(pb proto.Message) string {
	if pb == nil {
		return ""
	}

	if v, ok := pb.(Namer); ok {
		return v.ResourceName()
	}

	name := proto.MessageName(pb)

	v := strings.Split(name, ".")
	name = util.CamelToSnake(v[len(v)-1])
	if Plural() {
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
