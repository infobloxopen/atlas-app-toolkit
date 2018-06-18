package external

import (
	"database/sql/driver"
	"fmt"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/internal"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

// NewCodec returns new resource.Codec that encodes and decodes RPC representation
// of Identifier by treating them as references to an external resource.
// External Identifier implements sql.Scanner and driver.Valuer interfaces by
// storing itself in a format specified for Atlas References.
func NewCodec() resource.Codec {
	return codec{}
}

type codec struct{}

func (codec) Decode(pb *resourcepb.Identifier) (resource.Identifier, error) {
	var id identifier

	if pb == nil || (pb.ApplicationName == "" && pb.ResourceType == "" && pb.ResourceId == "") {
		return &id, nil
	}

	if pb.ApplicationName == "" || pb.ResourceType == "" || pb.ResourceId == "" {
		return nil, fmt.Errorf("external: identifier is not fully qualified - %s", pb)
	}

	id.applicationName, id.resourceType, id.resourceID = pb.ApplicationName, pb.ResourceType, pb.ResourceId
	id.valid = true

	return &id, nil
}

func (codec) Encode(id resource.Identifier) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if id == nil {
		return &pb, nil
	}

	eid, ok := id.(*identifier)
	if !ok {
		return nil, fmt.Errorf("external: invalid type of identifier - %T", id)
	}

	if eid == nil || !eid.valid {
		return &pb, nil
	}

	if eid.applicationName == "" || eid.resourceType == "" || eid.resourceID == "" {
		return nil, fmt.Errorf("external: resolved identifier is not fully qualified - %s", eid)
	}
	pb.ApplicationName, pb.ResourceType, pb.ResourceId = eid.applicationName, eid.resourceType, eid.resourceID

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
	return i.String(), nil
}

func (i *identifier) Scan(v interface{}) error {
	if v == nil {
		return nil
	}
	s, ok := v.(string)
	if !ok {
		return fmt.Errorf("external: invalid sql type of resource id %T", v)
	}
	i.applicationName, i.resourceType, i.resourceID = internal.ParseString(s)
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
	return true
}

func (i identifier) Valid() bool {
	return i.valid
}
