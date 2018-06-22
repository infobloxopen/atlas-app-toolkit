package fqstring

import (
	"database/sql/driver"
	"testing"

	resourcepb "github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

func TestCodec_Decode(t *testing.T) {
	codec := NewCodec()

	tcases := []struct {
		Identifier    *resourcepb.Identifier
		Value         driver.Value
		ExpectedError string
	}{
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "ext",
			},
			Value: "app/res/ext",
		},
		{
			Identifier: nil,
			Value:      nil,
		},
		{
			Identifier: &resourcepb.Identifier{},
			Value:      nil,
		},
		{
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
			},
			Value:         nil,
			ExpectedError: "fqstring: identifier is not fully qualified - application_name:\"app\" ",
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
	codec := NewCodec()

	tcases := []struct {
		Value         driver.Value
		Identifier    *resourcepb.Identifier
		ExpectedError string
	}{
		{
			Value:      nil,
			Identifier: &resourcepb.Identifier{},
		},
		{
			Value:         true,
			Identifier:    nil,
			ExpectedError: "fqstring: invalid resource id type bool",
		},
		{
			Value:         "app",
			Identifier:    &resourcepb.Identifier{},
			ExpectedError: "fqstring: resolved identifier is not fully qualified - app",
		},
		{
			Value: "app/res/ext",
			Identifier: &resourcepb.Identifier{
				ApplicationName: "app",
				ResourceType:    "res",
				ResourceId:      "ext",
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
