package fqstring

import (
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec()

	tcases := []struct {
		PB            *resourcepb.Identifier
		ID            *resource.Identifier
		ExpectedError string
	}{
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "ext",
			},
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: "app/res/ext",
			},
		},
		{
			PB: nil,
			ID: nil,
		},
		{
			PB: &resourcepb.Identifier{},
			ID: resource.Nil,
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
			},
			ID:            nil,
			ExpectedError: "fqstring: identifier is not fully qualified - application_name:\"app\" ",
		},
	}

	for n, tc := range tcases {
		id, err := codec.Decode(tc.PB)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if id != nil && tc.ID != nil && (id.Valid != tc.ID.Valid || id.ResourceID != tc.ID.ResourceID) {
			t.Errorf("tc: %d: invalid identifier %v, expected %v", n, id, tc.ID)
		}
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := NewCodec()

	tcases := []struct {
		ID            *resource.Identifier
		PB            *resourcepb.Identifier
		ExpectedError string
	}{
		{
			ID: nil,
			PB: &resourcepb.Identifier{},
		},
		{
			ID: resource.Nil,
			PB: &resourcepb.Identifier{},
		},
		{
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: "app",
			},
			PB:            &resourcepb.Identifier{},
			ExpectedError: "fqstring: resolved identifier is not fully qualified - app",
		},
		{
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: "app/res/ext",
			},
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "ext",
			},
		},
	}

	for n, tc := range tcases {
		pb, err := codec.Encode(tc.ID)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d:invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}

		if v := pb.GetApplicationName(); v != tc.PB.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.PB.ApplicationName)
		}
		if v := pb.GetResourceType(); v != tc.PB.ResourceType {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, v, tc.PB.ResourceType)
		}
		if v := pb.GetResourceId(); v != tc.PB.ResourceId {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.PB.ResourceId)
		}
	}
}
