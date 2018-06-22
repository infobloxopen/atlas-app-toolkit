package resource

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"

	"database/sql/driver"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

const defaultResource = "<default>"

var (
	mu       sync.RWMutex
	registry = make(map[string]Codec)
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

// Decode decodes identifier using a codec registered for pb.
// If codec has not been registered the error is returned.
func Decode(pb proto.Message, identifier *resourcepb.Identifier) (driver.Value, error) {
	codec, err := lookupCodec(pb)
	if err != nil {
		return nil, err
	}

	return codec.Decode(identifier)
}

// Encode encodes identifier using a codec registered for pb.
// If codec has not been registered the error is returned.
func Encode(pb proto.Message, value driver.Value) (*resourcepb.Identifier, error) {
	codec, err := lookupCodec(pb)
	if err != nil {
		return nil, err
	}

	return codec.Encode(value)
}

func lookupCodec(pb proto.Message) (Codec, error) {
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
		return nil, fmt.Errorf("resource: codec is not registered for resource %s", name)
	}
	return codec, nil
}
