package resource

import (
	"encoding/json"
	"strings"

	"github.com/golang/protobuf/jsonpb"
)

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface by marshal
// Identifier from a JSON string in accordance with Atlas Reference format
// 		<application_name>/<resource_type>/<resource_id>
// Support "null" value.
func (m Identifier) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	v := BuildString(m.GetApplicationName(), m.GetResourceType(), m.GetResourceId())
	if v == "" {
		v = "null"
	}
	return []byte(`"` + v + `"`), nil
}

// MarshalJSON implements json.Marshaler interface
func (m *Identifier) MarshalJSON() ([]byte, error) {
	return m.MarshalJSONPB(nil)
}

var _ json.Marshaler = &Identifier{}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface by unmarshal
// Identifier to a JSON string in accordance with Atlas Reference format
// 		<application_name>/<resource_type>/<resource_id>
// Support "null" value.
func (m *Identifier) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	v := strings.Trim(string(data), "\"")
	if v == "null" {
		v = ""
	}
	m.ApplicationName, m.ResourceType, m.ResourceId = ParseString(v)
	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *Identifier) UnmarshalJSON(data []byte) error {
	return m.UnmarshalJSONPB(nil, data)
}

var _ json.Unmarshaler = &Identifier{}
