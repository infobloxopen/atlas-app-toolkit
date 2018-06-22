package fqstring

import (
	"fmt"

	"database/sql/driver"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

// NewCodec returns new resource.Codec that encodes and decodes Protocol Buffer
// representation of infoblox.rpc.Identifier by encoding/decoding it to be
// stored in SQL DB as a fully qualified string value in a format specified for Atlas References.
// If the infoblox.rpc.Identifier is empty the resource.Nil is returned, it it has
// missing one of the part the error is returned.
func NewCodec() resource.Codec {
	return codec{}
}

type codec struct{}

func (codec) Decode(pb *resourcepb.Identifier) (driver.Value, error) {
	if pb == nil || (pb.ApplicationName == "" && pb.ResourceType == "" && pb.ResourceId == "") {
		return nil, nil
	}

	if pb.ApplicationName == "" || pb.ResourceType == "" || pb.ResourceId == "" {
		return nil, fmt.Errorf("fqstring: identifier is not fully qualified - %s", pb)
	}

	value := resourcepb.BuildString(pb.ApplicationName, pb.ResourceType, pb.ResourceId)

	return value, nil
}

func (codec) Encode(value driver.Value) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if value == nil {
		return &pb, nil
	}

	v, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("fqstring: invalid resource id type %T", value)
	}

	pb.ApplicationName, pb.ResourceType, pb.ResourceId = resourcepb.ParseString(v)

	if pb.ApplicationName == "" || pb.ResourceType == "" || pb.ResourceId == "" {
		return nil, fmt.Errorf("fqstring: resolved identifier is not fully qualified - %v", v)
	}

	return &pb, nil
}
