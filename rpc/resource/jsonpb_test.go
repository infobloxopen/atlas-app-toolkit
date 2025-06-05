package resource

import (
	"testing"

	"bytes"

	jsonpb "google.golang.org/protobuf/encoding/protojson"
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
