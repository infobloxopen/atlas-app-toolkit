package uuid

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

// NewCodec returns new resource.Codec that encodes and decodes Protocol Buffer
// representation of infoblox.rpc.Identifier by  by encoding/decoding it to be
// stored in SQL DB as an internal resource with Resource ID converted to
// a string in UUID format.
// Internal means a resource that belongs to current application.
// If the ResourceId of infoblox.rpc.Identifier is empty the value of
// the resource.Default will be returned.
// Does not support NULL values
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

func (c codec) String() string {
	return "codec: " + c.applicationName + "/" + c.resourceType
}

func (c codec) Decode(pb *resourcepb.Identifier) (*resource.Identifier, error) {
	if v := pb.GetApplicationName(); v != "" && v != c.applicationName {
		return nil, fmt.Errorf("uuid: invalid application name %s of %s", pb.GetApplicationName(), c)
	}
	if v := pb.GetResourceType(); v != "" && v != c.resourceType {
		return nil, fmt.Errorf("uuid: invalid resource type %s of %s", pb.GetResourceType(), c)
	}

	if pb.GetResourceId() == "" {
		return &resource.Default, nil
	}
	v, err := uuid.Parse(pb.GetResourceId())
	if err != nil {
		return nil, fmt.Errorf("uuid: unable to convert resource id %v of %s - %s", pb.ResourceId, c, err)
	}
	return &resource.Identifier{
		Valid:      true,
		ResourceID: v.String(),
	}, nil
}

func (c codec) Encode(id *resource.Identifier) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if id == nil || id.ResourceID == nil || !id.Valid {
		return nil, fmt.Errorf("uuid: the resource id of %s cannot be NULL", c)
	}

	pb.ApplicationName = c.applicationName
	pb.ResourceType = c.resourceType

	v, ok := id.ResourceID.(string)
	if !ok {
		return nil, fmt.Errorf("uuid: invalid resource id type %T of %s", id.ResourceID, c)
	}

	return &resourcepb.Identifier{
		ApplicationName: c.applicationName,
		ResourceType:    c.resourceType,
		ResourceId:      v,
	}, nil
}
