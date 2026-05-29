package resource

import (
	"testing"

	"bytes"

	"github.com/golang/protobuf/jsonpb"
)

func TestIdentifier_MarshalJSONPB(t *testing.T) {
	tcases := []struct {
		Identifier         *Identifier
		ExpectedJSONString string
	}{
		{
			&Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res1",
			},
			`"app/resource/res1"`,
		},
		{
			&Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
			`"null"`,
		},
	}

	var (
		marshaler = &jsonpb.Marshaler{}
		buffer    = &bytes.Buffer{}
	)

	for _, tc := range tcases {
		buffer.Reset()

		if err := marshaler.Marshal(buffer, tc.Identifier); err != nil {
			t.Errorf("failed to marshal identifier %s - %s", tc.Identifier, err)
		}

		if s := buffer.String(); s != tc.ExpectedJSONString {
			t.Errorf("ivalid identifier %s, expected %s", s, tc.ExpectedJSONString)
		}
	}
}

func TestIdentifier_UnmarhsalJSONPB(t *testing.T) {
	tcases := []struct {
		JSONData           string
		ExpectedIdentifier *Identifier
	}{
		{
			`"app/resource/res2"`,
			&Identifier{
				ApplicationName: "app",
				ResourceType:    "resource",
				ResourceId:      "res2",
			},
		},
		{
			`"null"`,
			&Identifier{
				ApplicationName: "",
				ResourceType:    "",
				ResourceId:      "",
			},
		},
	}

	var (
		unmarshaler = &jsonpb.Unmarshaler{}
		buffer      = &bytes.Buffer{}
	)

	for _, tc := range tcases {
		buffer.Reset()
		buffer.WriteString(tc.JSONData)
		id := &Identifier{}

		if err := unmarshaler.Unmarshal(buffer, id); err != nil {
			t.Errorf("failed to unmarshal identifier %s", err)
		}

		if id.String() != tc.ExpectedIdentifier.String() {
			t.Errorf("invalid identifier %s, expected %s", id, tc.ExpectedIdentifier)
		}
	}
}

func TestIdentifier_UnmarshalJSONPB_InvalidTypes(t *testing.T) {
	tcases := []struct {
		Name     string
		JSONData string
	}{
		{"array", `["app/resource/id1","app/resource/id2"]`},
		{"empty_array", `[]`},
		{"number", `12345`},
		{"negative_number", `-1`},
		{"float_number", `1.5`},
		{"object", `{"application_name":"app","resource_type":"res","resource_id":"id1"}`},
		{"empty_object", `{}`},
		{"boolean_true", `true`},
		{"boolean_false", `false`},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			id := &Identifier{}
			err := id.UnmarshalJSONPB(nil, []byte(tc.JSONData))
			if err == nil {
				t.Errorf("expected error for input %s, got nil", tc.JSONData)
			}
		})
	}
}

func TestIdentifier_UnmarshalJSONPB_ValidInputs(t *testing.T) {
	tcases := []struct {
		Name               string
		JSONData           string
		ExpectedIdentifier *Identifier
	}{
		{
			"valid resource identifier",
			`"app/resource/res1"`,
			&Identifier{ApplicationName: "app", ResourceType: "resource", ResourceId: "res1"},
		},
		{
			"partial identifier with only resource id",
			`"res1"`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: "res1"},
		},
		{
			"partial identifier with type and id",
			`"resource/res1"`,
			&Identifier{ApplicationName: "", ResourceType: "resource", ResourceId: "res1"},
		},
		{
			"null literal",
			`null`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"quoted null",
			`"null"`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"empty string",
			`""`,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
		{
			"empty data",
			``,
			&Identifier{ApplicationName: "", ResourceType: "", ResourceId: ""},
		},
	}

	for _, tc := range tcases {
		t.Run(tc.Name, func(t *testing.T) {
			id := &Identifier{}
			err := id.UnmarshalJSONPB(nil, []byte(tc.JSONData))
			if err != nil {
				t.Errorf("unexpected error for input %s: %s", tc.JSONData, err)
			}
			if id.String() != tc.ExpectedIdentifier.String() {
				t.Errorf("got %s, expected %s", id, tc.ExpectedIdentifier)
			}
		})
	}
}
