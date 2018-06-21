package fqstring

import (
	"fmt"

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

func (codec) Decode(pb *resourcepb.Identifier) (*resource.Identifier, error) {
	if pb == nil || (pb.ApplicationName == "" && pb.ResourceType == "" && pb.ResourceId == "") {
		return &resource.Nil, nil
	}

	if pb.ApplicationName == "" || pb.ResourceType == "" || pb.ResourceId == "" {
		return nil, fmt.Errorf("fqstring: identifier is not fully qualified - %s", pb)
	}

	value := resourcepb.BuildString(pb.ApplicationName, pb.ResourceType, pb.ResourceId)
	id := resource.Identifier{
		Valid:      true,
		ResourceID: value,
	}

	return &id, nil
}

func (codec) Encode(id *resource.Identifier) (*resourcepb.Identifier, error) {
	var pb resourcepb.Identifier

	if id == nil || id.ResourceID == nil || !id.Valid {
		return &pb, nil
	}

	value := fmt.Sprintf("%v", id.ResourceID)
	pb.ApplicationName, pb.ResourceType, pb.ResourceId = resourcepb.ParseString(value)

	if pb.ApplicationName == "" || pb.ResourceType == "" || pb.ResourceId == "" {
		return nil, fmt.Errorf("fqstring: resolved identifier is not fully qualified - %v", id.ResourceID)
	}

	return &pb, nil
}
