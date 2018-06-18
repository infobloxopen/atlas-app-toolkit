package uuid

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/internal"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

// NewCodec returns new resource.Codec that encodes and decodes RPC representation
// of Identifier by treating them as references to an internal resource with
// Resource ID in UUID format.
// Internal means a resource that belongs to current application.
// If ResourceID of decoded RPC identifier is empty the codec populates it with
// new id in UUID format.
// Internal Identifier implements sql.Scanner and driver.Valuer interfaces by
// storing only the Resource ID part of itself.
func NewCodec(applicationName, resourceType string) resource.Codec {
	return &codec{
		applicationName: applicationName,
		resourceType:    resourceType,
	}
}

type codec struct {
	applicationName string
	resourceType    string
}

func (m codec) String() string {
	return fmt.Sprintf("uuid codec of %s/%s", m.applicationName, m.resourceType)
}

func (m codec) Decode(pb *resourcepb.Identifier) (resource.Identifier, error) {
	var id identifier

	if pb == nil {
		return &id, nil
	}

	if pb.ApplicationName != "" && pb.ApplicationName != m.applicationName {
		return nil, fmt.Errorf("uuid: invalid application name %s of %s", pb.ApplicationName, m)
	}
	id.applicationName = m.applicationName

	if pb.ResourceType != "" && pb.ResourceType != m.resourceType {
		return nil, fmt.Errorf("uuid: invalid resource type %s of %s", pb.ResourceType, m)
	}
	id.resourceType = m.resourceType

	if pb.ResourceId == "" {
		v, err := uuid.NewUUID()
		if err != nil {
			return nil, err
		}
		id.resourceID = v.String()
	} else {
		id.resourceID = pb.ResourceId
	}
	id.valid = true

	return &id, nil
}

func (m codec) Encode(id resource.Identifier) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if id == nil {
		return &pb, nil
	}

	uid, ok := id.(*identifier)
	if !ok {
		return nil, fmt.Errorf("uuid: invalid type of identifier in %s", m)
	}

	if uid == nil || !uid.valid || uid.resourceID == "" {
		return &pb, nil
	}

	pb.ApplicationName = m.applicationName
	pb.ResourceType = m.resourceType

	v, err := uuid.Parse(uid.resourceID)
	if err != nil {
		return nil, fmt.Errorf("uuid: invalid resource id %s in %s - %s", uid.resourceID, m, err)
	}
	pb.ResourceId = v.String()

	return &pb, nil
}

type identifier struct {
	applicationName string
	resourceType    string
	resourceID      string
	valid           bool
}

func (i identifier) String() string {
	return internal.BuildString(i.applicationName, i.resourceType, i.resourceID)
}

func (i identifier) Value() (driver.Value, error) {
	if !i.valid {
		return nil, nil
	}
	return i.resourceID, nil
}

func (i *identifier) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("uuid: invalid sql type of resource id %T", v)
	}
	i.resourceID = s
	i.valid = true

	return nil
}

func (i identifier) ApplicationName() string {
	return i.applicationName
}

func (i identifier) ResourceType() string {
	return i.resourceType
}

func (i identifier) ResourceID() string {
	return i.resourceID
}

func (i identifier) External() bool {
	return false
}

func (i identifier) Valid() bool {
	return i.valid
}
