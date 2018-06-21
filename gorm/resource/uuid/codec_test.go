package uuid

import (
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/gorm/resource"
	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec("app", "res")

	tcases := []struct {
		PB            *resourcepb.Identifier
		ID            *resource.Identifier
		ExpectedError string
	}{
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "00000000-0000-0000-0000-000000000000",
			},
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: "00000000-0000-0000-0000-000000000000",
			},
		},
		{
			PB: nil,
			ID: nil,
		},
		{
			PB: &resourcepb.Identifier{},
			ID: nil,
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "bad",
			},
			ID:            nil,
			ExpectedError: "uuid: invalid application name bad of codec: app/res",
		},
		{
			PB: &resourcepb.Identifier{
				ResourceType: "bad",
			},
			ID:            nil,
			ExpectedError: "uuid: invalid resource type bad of codec: app/res",
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "will_fail",
			},
			ID:            nil,
			ExpectedError: "uuid: unable to convert resource id will_fail of codec: app/res - invalid UUID length: 9",
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "",
			},
			ID: &resource.Default,
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
	codec := NewCodec("app", "res")

	tcases := []struct {
		ID            *resource.Identifier
		PB            *resourcepb.Identifier
		ExpectedError string
	}{
		{
			ID:            nil,
			PB:            nil,
			ExpectedError: "uuid: the resource id of codec: app/res cannot be NULL",
		},
		{
			ID:            &resource.Nil,
			PB:            nil,
			ExpectedError: "uuid: the resource id of codec: app/res cannot be NULL",
		},
		{
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: 12,
			},
			PB:            &resourcepb.Identifier{},
			ExpectedError: "uuid: invalid resource id type int of codec: app/res",
		},
		{
			ID: &resource.Identifier{
				Valid:      true,
				ResourceID: "00000000-0000-0000-0000-000000000000",
			},
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "00000000-0000-0000-0000-000000000000",
			},
		},
	}

	for n, tc := range tcases {
		pb, err := codec.Encode(tc.ID)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d:invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}

		if v := pb.GetApplicationName(); v != tc.PB.GetApplicationName() {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.PB.GetApplicationName())
		}
		if v := pb.GetResourceType(); v != tc.PB.GetResourceType() {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, v, tc.PB.GetResourceType())
		}
		if v := pb.GetResourceId(); v != tc.PB.GetResourceId() {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.PB.GetResourceId())
		}
	}
}
