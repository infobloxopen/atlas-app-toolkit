package integer

import (
	"testing"

	"database/sql/driver"

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
				ResourceId:      "123",
			},
			Value: int64(123),
		},
		{
			Identifier: nil,
			Value:      int64(0),
		},
		{
			Identifier: &resourcepb.Identifier{},
			Value:      int64(0),
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "bad",
			},
			Value:         nil,
			ExpectedError: "integer: invalid application name bad of codec: app/res",
		},
		{
			Identifier: &resourcepb.Identifier{
				ResourceType: "bad",
			},
			Value:         nil,
			ExpectedError: "integer: invalid resource type bad of codec: app/res",
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "will_fail",
			},
			Value:         nil,
			ExpectedError: "integer: unable to convert resource id will_fail of codec: app/res - invalid syntax",
		},
		{
			Value: int64(0),
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "",
			},
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
			ExpectedError: "integer: the resource id of codec: app/res cannot be NULL",
		},
		{
			Value:         "val",
			Identifier:    nil,
			ExpectedError: "integer: invalid resource id type string of codec: app/res",
		},
		{
			Value: int64(12),
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "12",
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
