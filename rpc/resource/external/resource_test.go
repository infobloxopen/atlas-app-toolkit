package external

import (
	"database/sql/driver"
	"testing"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/resourcepb"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec()

	tcases := []struct {
		PB              *resourcepb.Identifier
		ApplicationName string
		ResourceType    string
		ResourceID      string
		Valid           bool
		External        bool
		ExpectedError   string
	}{
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "ext",
			},
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "ext",
			Valid:           true,
			External:        true,
			ExpectedError:   "",
		},
		{
			PB:       nil,
			External: true,
		},
		{
			PB:       &resourcepb.Identifier{},
			External: true,
		},
		{
			PB: &resourcepb.Identifier{
				ApplicationName: "app",
			},
			External:      true,
			ExpectedError: "external: identifier is not fully qualified - application_name:\"app\" ",
		},
	}

	for n, tc := range tcases {
		id, err := codec.Decode(tc.PB)
		if err != nil && tc.ExpectedError != err.Error() {
			t.Errorf("tc %d:invalid error message %s, expected %s", n, err, tc.ExpectedError)
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
		if id.ResourceID() != tc.ResourceID {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, id.ResourceID(), tc.ResourceID)
		}
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := NewCodec()

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
				applicationName: "app",
				valid:           true,
			},
			ExpectedError: "external: resolved identifier is not fully qualified - app",
		},
		{
			ID: &identifier{
				applicationName: "app",
				resourceType:    "res",
				resourceID:      "ext",
			},
		},
		{
			ID: &identifier{
				applicationName: "app",
				resourceType:    "res",
				resourceID:      "ext",
				valid:           true,
			},
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "ext",
		},
	}

	for n, tc := range tcases {
		pb, err := codec.Encode(tc.ID)
		if err != nil && tc.ExpectedError != err.Error() {
			t.Errorf("tc %d:invalid error message %s, expected %s", n, err, tc.ExpectedError)
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
			Value:           "app/res/ext",
			Valid:           true,
			External:        true,
			ApplicationName: "app",
			ResourceType:    "res",
			ResourceID:      "ext",
			ExpectedError:   "",
		},
		{
			Value:    nil,
			External: true,
		},
		{
			Value:         1986,
			External:      true,
			ExpectedError: "external: invalid sql type of resource id int",
		},
	}

	for n, tc := range tcases {
		var id identifier
		if err := id.Scan(tc.Value); err != nil && tc.ExpectedError != err.Error() {
			t.Errorf("tc %d: invalid error message %s, expected %s", n, err, tc.ExpectedError)
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
				resourceID:      "ext",
				valid:           true,
			},
			Value:         driver.Value("app/res/ext"),
			ExpectedError: "",
		},
		{
			Identifier:    &identifier{},
			Value:         nil,
			ExpectedError: "",
		},
	}

	for n, tc := range tcases {
		v, err := tc.Identifier.Value()
		if err != nil && err.Error() != tc.ExpectedError {
			t.Errorf("tc %d: invalid error message %s, expected %s", n, err, tc.ExpectedError)
		}
		if v != tc.Value {
			t.Errorf("tc %d: invalid value %s, expected %s", n, v, tc.Value)
		}
	}
}
