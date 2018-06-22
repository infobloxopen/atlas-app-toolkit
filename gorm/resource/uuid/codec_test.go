package uuid

import (
	"database/sql/driver"
	"testing"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec("app", "res")

	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Value         driver.Value
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "00000000-0000-0000-0000-000000000000",
			},
			Value: "00000000-0000-0000-0000-000000000000",
		},
		{
			Identifier: nil,
			Value:      "",
		},
		{
			Identifier: &resourcepb.Identifier{},
			Value:      "",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "bad",
			},
			Value:         nil,
			ExpectedError: "uuid: invalid application name bad of codec: app/res",
		},
		{
			Identifier: &resourcepb.Identifier{
				ResourceType: "bad",
			},
			Value:         nil,
			ExpectedError: "uuid: invalid resource type bad of codec: app/res",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "will_fail",
			},
			Value:         nil,
			ExpectedError: "uuid: unable to convert resource id will_fail of codec: app/res - invalid UUID length: 9",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "",
			},
			Value: "",
		},
	}

	for n, tc := range tcases {
		value, err := codec.Decode(tc.Identifier)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d: invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}
		if value != nil && tc.Value != nil && value != tc.Value {
			t.Errorf("tc: %d: invalid identifier %v, expected %v", n, value, tc.Value)
		}
	}
}

func TestCodec_Encode(t *testing.T) {
	codec := NewCodec("app", "res")

	tcases := []struct {
		Value         driver.Value
		Identifier    *resourcepb.Identifier
		ExpectedError string
	}{
		{
			Value:         nil,
			Identifier:    nil,
			ExpectedError: "uuid: the resource id of codec: app/res cannot be NULL",
		},
		{
			Value:         12,
			Identifier:    &resourcepb.Identifier{},
			ExpectedError: "uuid: invalid resource id type int of codec: app/res",
		},
		{
			Value: "00000000-0000-0000-0000-000000000000",
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "00000000-0000-0000-0000-000000000000",
			},
		},
	}

	for n, tc := range tcases {
		pb, err := codec.Encode(tc.Value)
		if (err != nil && tc.ExpectedError != err.Error()) || (err == nil && tc.ExpectedError != "") {
			t.Errorf("tc %d:invalid error message %q, expected %q", n, err, tc.ExpectedError)
		}

		if v := pb.GetApplicationName(); v != tc.Identifier.GetApplicationName() {
			t.Errorf("tc %d: invalid application name %s, expected %s", n, v, tc.Identifier.GetApplicationName())
		}
		if v := pb.GetResourceType(); v != tc.Identifier.GetResourceType() {
			t.Errorf("tc %d: invalid resource type %s, expected %s", n, v, tc.Identifier.GetResourceType())
		}
		if v := pb.GetResourceId(); v != tc.Identifier.GetResourceId() {
			t.Errorf("tc %d: invalid resource id %s, expected %s", n, v, tc.Identifier.GetResourceId())
		}
	}
}
