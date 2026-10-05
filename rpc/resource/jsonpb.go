package resource

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/golang/protobuf/jsonpb"
)

// MarshalJSONPB implements jsonpb.JSONPBMarshaler interface by marshal
// Identifier from a JSON string in accordance with Atlas Reference format
//
//	<application_name>/<resource_type>/<resource_id>
//
// Support "null" value.
func (m Identifier) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	v := BuildString(m.GetApplicationName(), m.GetResourceType(), m.GetResourceId())
	if v == "" {
		v = "null"
	}
	return marshalString(v)
}

// marshalString encodes v as a JSON string. HTML escaping is disabled so that
// ids containing &, < or > keep the bytes they were written with before.
func marshalString(v string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), nil
}

// MarshalJSON implements json.Marshaler interface
func (m *Identifier) MarshalJSON() ([]byte, error) {
	return m.MarshalJSONPB(nil)
}

var _ json.Marshaler = &Identifier{}

// UnmarshalJSONPB implements jsonpb.JSONPBUnmarshaler interface by unmarshal
// Identifier to a JSON string in accordance with Atlas Reference format
//
//	<application_name>/<resource_type>/<resource_id>
//
// Support "null" value.
func (m *Identifier) UnmarshalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	if len(data) == 0 {
		return nil
	}
	// Decoding into *string rejects arrays, objects, numbers and booleans,
	// accepts a JSON null (leaving s nil) and resolves JSON escapes.
	var s *string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("invalid value for resource identifier: expected a string, got: %s: %w", truncateBytes(data, 64), err)
	}
	v := ""
	if s != nil {
		v = *s
	}
	if v == "null" {
		v = ""
	}
	m.ApplicationName, m.ResourceType, m.ResourceId = ParseString(v)
	return nil
}

// truncateBytes returns data as a valid UTF-8 string cut at a rune boundary near
// maxLen bytes, with an ellipsis when cut. The error text ends up in gRPC status messages,
// which protojson refuses to marshal when they contain invalid UTF-8.
func truncateBytes(data []byte, maxLen int) string {
	if len(data) <= maxLen {
		return strings.ToValidUTF8(string(data), "\uFFFD")
	}
	for maxLen > 0 && !utf8.RuneStart(data[maxLen]) {
		maxLen--
	}
	return strings.ToValidUTF8(string(data[:maxLen]), "\uFFFD") + "..."
}

// UnmarshalJSON implements json.Unmarshaler interface
func (m *Identifier) UnmarshalJSON(data []byte) error {
	return m.UnmarshalJSONPB(nil, data)
}

var _ json.Unmarshaler = &Identifier{}
