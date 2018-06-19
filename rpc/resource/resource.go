package resource //import "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"

import (
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

const defaultResource = "<default>"

var (
	mu       sync.RWMutex
	registry = make(map[string]Codec)
)

// Identifier defines an interface to represent RPC Identifier.
type Identifier interface {
	// ApplicationName is an application identifier that will be used among
	// other infrastructure services to identify the application
	ApplicationName() string

	// ResourceType is an application specific type name of a resource
	ResourceType() string

	// ResourceID is an application specific resource identity of a resource
	ResourceID() string

	// External is a flag indicates whether the identifier represents
	// an fq resource or holds an identity of a resource that is belongs
	// to an application itself.
	External() bool

	// Valid is a flag indicates whether the identifier is a valid or not.
	// The criteria is up to an implementation.
	Valid() bool
}

// Codec defines the interface package uses to encode and decode RPC Identifier
// to an Identifier interface implementation.
// Note that implementation must be thread safe.
type Codec interface {
	// Encode returns RPC representation of Identifier interface
	Encode(Identifier) (*resourcepb.Identifier, error)
	// Decode returns implementation of Identifier interface based on RPC representation
	Decode(*resourcepb.Identifier) (Identifier, error)
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
func Decode(pb proto.Message, identifier *resourcepb.Identifier) (Identifier, error) {
	codec, err := lookupCodec(pb)
	if err != nil {
		return nil, err
	}

	return codec.Decode(identifier)
}

// Encode encodes identifier using a codec registered for pb.
// If codec has not been registered the error is returned.
func Encode(pb proto.Message, identifier Identifier) (*resourcepb.Identifier, error) {
	codec, err := lookupCodec(pb)
	if err != nil {
		return nil, err
	}

	return codec.Encode(identifier)
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
