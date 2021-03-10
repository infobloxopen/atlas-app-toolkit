package resource

import (
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
