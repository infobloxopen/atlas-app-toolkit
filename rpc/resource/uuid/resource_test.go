package uuid

import (
	"database/sql/driver"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec("app", "res")

	tcases := []struct {
		PB              *resourcepb.Identifier
		ApplicationName string
		ResourceType    string
		ResourceID      string
		Valid           bool
		External        bool
		ExpectedError   string
		NotEmpty        bool
	}{
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "uuid",
			},
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "uuid",
			Valid:           true,
		},
		{
			PB: nil,
		},
		{
			PB:              &resourcepb.Identifier{},
			ApplicationName: "app",
			ResourceType:    "res",
			Valid:           true,
			NotEmpty:        true,
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "myapp",
			},
			ExpectedError: "uuid: invalid application name myapp of uuid codec of app/res",
		},
		{
			PB: &resourcepb.Identifier{
				ResourceType: "myres",
			},
			ExpectedError: "uuid: invalid resource type myres of uuid codec of app/res",
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "",
			},
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "uuid",
			Valid:           true,
			NotEmpty:        true,
		},
	}

	for n, tc := range tcases {
		id, err := codec.Decode(tc.PB)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d:invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if id == nil {
			continue
		}
		if id.Valid() != tc.Valid {
			t.Errorf("tc %d: resource valid is %t, expected %t", n, id.Valid(), tc.Valid)
		}
		if id.External() != tc.External {
			t.Errorf("tc %d: resource external is %t, expected %t", n, id.External(), tc.External)
		}
		if id.ApplicationName() != tc.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, id.ApplicationName(), tc.ApplicationName)
		}
		if id.ResourceType() != tc.ResourceType {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, id.ResourceType(), tc.ResourceType)
		}

		if tc.NotEmpty {
			if id.ResourceID() == "" {
				t.Errorf("invalid resource id %s, expected non empty", id.ResourceID())
			}
			continue
		}

		if id.ResourceID() != tc.ResourceID {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, id.ResourceID(), tc.ResourceID)
		}
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := NewCodec("app", "res")

	tcases := []struct {
		ID              resource.Identifier
		ApplicationName string
		ResourceType    string
		ResourceID      string
		ExpectedError   string
	}{
		{
			ID: nil,
		},
		{
			ID: (*identifier)(nil),
		},
		{
			ID: &identifier{},
		},
		{
			ID: &identifier{
				resourceID: "uuid",
				valid:      true,
			},
			ExpectedError: "uuid: invalid resource id uuid in uuid codec of app/res - invalid UUID length: 4",
		},
		{
			ID: &identifier{
				resourceID: "cd89d33a-7310-11e8-8297-787b8ac3ce32",
				valid:      true,
			},
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "cd89d33a-7310-11e8-8297-787b8ac3ce32",
		},
	}

	for n, tc := range tcases {
		pb, err := codec.Encode(tc.ID)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d:invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if pb == nil {
			continue
		}

		if v := pb.GetApplicationName(); v != tc.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.ApplicationName)
		}
		if v := pb.GetResourceType(); v != tc.ResourceType {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, v, tc.ResourceType)
		}
		if v := pb.GetResourceId(); v != tc.ResourceID {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.ResourceID)
		}
	}
}

func TestIdentifier_Scan(t *testing.T) {
	tcases := []struct {
		Value           interface{}
		Valid           bool
		External        bool
		ApplicationName string
		ResourceType    string
		ResourceID      string
		ExpectedError   string
	}{
		{
			Value:      "uuid",
			Valid:      true,
			ResourceID: "uuid",
		},
		{
			Value: nil,
		},
		{
			Value:         2018,
			ExpectedError: "uuid: invalid sql type of resource id int",
		},
	}

	for n, tc := range tcases {
		var id identifier
		if err := id.Scan(tc.Value); (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if id.Valid() != tc.Valid {
			t.Errorf("tc %d: resource valid is %t, expected %t", n, id.Valid(), tc.Valid)
		}
		if id.External() != tc.External {
			t.Errorf("tc %d: resource external is %t, expected %t", n, id.External(), tc.External)
		}
		if id.ApplicationName() != tc.ApplicationName {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, id.ApplicationName(), tc.ApplicationName)
		}
		if id.ResourceType() != tc.ResourceType {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, id.ResourceType(), tc.ResourceType)
		}
		if id.ResourceID() != tc.ResourceID {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, id.ResourceID(), tc.ResourceID)
		}
	}
}

func TestIdentifier_Value(t *testing.T) {
	tcases := []struct {
		Identifier    *identifier
		Value         driver.Value
		ExpectedError string
	}{
		{
			Identifier: &identifier{
				applicationName: "app",
				resourceType:    "res",
				resourceID:      "uuid",
				valid:           true,
			},
			Value: driver.Value("uuid"),
		},
		{
			Identifier: &identifier{},
			Value:      nil,
		},
	}

	for n, tc := range tcases {
		v, err := tc.Identifier.Value()
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if v != tc.Value {
			t.Errorf("tc %d: invalid value %s, expected %s", n, v, tc.Value)
		}
	}
}
