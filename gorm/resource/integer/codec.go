package integer

import (
	"fmt"
	"strconv"

	"database/sql/driver"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

// NewCodec returns new resource.Codec that encodes and decodes Protocol Buffer
// representation of Identifier to/from int64 type.
//
// The only resource_id part is encoded/decoded.
//
// Does not support NULL values. Could be used for "serial" primary keys.
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

func (c codec) Decode(pb *resourcepb.Identifier) (driver.Value, error) {
	if v := pb.GetApplicationName(); v != "" && v != c.applicationName {
		return nil, fmt.Errorf("integer: invalid application name %s of %s", pb.GetApplicationName(), c)
	}
	if v := pb.GetResourceType(); v != "" && v != c.resourceType {
		return nil, fmt.Errorf("integer: invalid resource type %s of %s", pb.GetResourceType(), c)
	}

	if pb.GetResourceId() == "" {
		return int64(0), nil
	}
	i, err := strconv.ParseInt(pb.ResourceId, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("integer: unable to convert resource id %v of %s - %s", pb.ResourceId, c, err.(*strconv.NumError).Err)
	}
	return i, nil
}

func (c codec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if value == nil {
		return nil, fmt.Errorf("integer: the resource id of %s cannot be NULL", c)
	}

	pb.ApplicationName = c.applicationName
	pb.ResourceType = c.resourceType

	v, ok := value.(int64)
	if !ok {
		return nil, fmt.Errorf("integer: invalid resource id type %T of %s", value, c)
	}

	return &resourcepb.Identifier{
		ApplicationName: c.applicationName,
		ResourceType:    c.resourceType,
		ResourceId:      strconv.FormatInt(v, 10),
	}, nil
}
