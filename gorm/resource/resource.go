package resource

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

const defaultResource = "<default>"

var (
	// Nil is an empty Identifier
	Nil = Identifier{}
	// Default is a Identifier with "DEFAULT" (SQL) value,
	Default = Identifier{
		Valid:      true,
		ResourceID: "DEFAULT",
	}

	mu       sync.RWMutex
	registry = make(map[string]Codec)
)

// Identifier represents SQL value with NULL support.
type Identifier struct {
	Valid      bool
	ResourceID interface{}
}

// Value implements driver.Valuer interface with support of NULL values.
// If i.ID implements driver.Valuer interface than it will be used instead.
func (i Identifier) Value() (driver.Value, error) {
	if i.ResourceID == nil || !i.Valid {
		return nil, nil
	}
	if v, ok := i.ResourceID.(driver.Valuer); ok {
		return v.Value()
	}
	return i.ResourceID, nil
}

// Scan implements sql.Scanner interface with support of NULL values.
// If i.ID implements sql.Scanner interface than it will be used instead.
func (i *Identifier) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	i.Valid = true
	if s, ok := i.ResourceID.(sql.Scanner); ok {
		return s.Scan(v)
	}
	i.ResourceID = v
	return nil
}

// Codec defines the interface package uses to encode and decode Protocol Buffer
// infoblox.rpc.Identifier to an Identifier value.
// Note that implementation must be thread safe.
type Codec interface {
	// Encode returns Protocol Buffer representation of an Identifier value.
	Encode(*Identifier) (*resourcepb.Identifier, error)
	// Decode returns an Identifier value based on Protocol Buffer representation
	Decode(*resourcepb.Identifier) (*Identifier, error)
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
func Decode(pb proto.Message, identifier *resourcepb.Identifier) (*Identifier, error) {
	codec, err := lookupCodec(pb)
	if err != nil {
		return nil, err
	}

	return codec.Decode(identifier)
}

// Encode encodes identifier using a codec registered for pb.
// If codec has not been registered the error is returned.
func Encode(pb proto.Message, identifier *Identifier) (*resourcepb.Identifier, error) {
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
