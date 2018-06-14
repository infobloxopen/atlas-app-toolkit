package resource

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

const externalResource = "<external>"

var (
	mu       sync.RWMutex
	registry = make(map[string]*struct {
		Manager
	})
)

type Identifier interface {
	fmt.Stringer
	sql.Scanner
	driver.Valuer

	ApplicationName() string
	ResourceType() string
	ResourceID() string

	IsExternal() bool
	IsNil() bool
}

type Manager interface {
	Build(*resourcepb.Identifier) (Identifier, error)
	Resolve(Identifier) (*resourcepb.Identifier, error)
}

func RegisterManager(manager Manager, message proto.Message) {
	mu.Lock()
	defer mu.Unlock()

	var name string
	if message == nil {
		name = externalResource
	} else {
		name = proto.MessageName(message)
	}

	if manager == nil {
		panic("resource: register nil manager for resource " + name)
	}

	v, ok := registry[name]
	if !ok {
		registry[name] = &struct {
			Manager
		}{Manager: manager}
	}
	if v.Manager != nil {
		panic("resource: register manager called twice for resource " + name)
	}
	v.Manager = manager
}

func Build(message proto.Message, identifier *resourcepb.Identifier) (Identifier, error) {
	mu.RLock()
	defer mu.RUnlock()

	var name string
	if message == nil {
		name = externalResource
	} else {
		name = proto.MessageName(message)
	}
	v, ok := registry[name]
	if !ok || v == nil || v.Manager == nil {
		return nil, fmt.Errorf("resource: manager is not registered for resource %s", name)
	}

	return v.Manager.Build(identifier)
}

func Resolve(message proto.Message, identifier Identifier) (*resourcepb.Identifier, error) {
	mu.RLock()
	defer mu.RUnlock()

	var name string
	if message == nil {
		name = externalResource
	} else {
		name = proto.MessageName(message)
	}
	v, ok := registry[name]
	if !ok || v == nil || v.Manager == nil {
		return nil, fmt.Errorf("resource: manager is not registered for resource %s", name)
	}

	return v.Manager.Resolve(identifier)
}
