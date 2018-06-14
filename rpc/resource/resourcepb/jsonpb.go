package resourcepb

import (
	"github.com/golang/protobuf/jsonpb"

	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource/internal"
)

func (m Identifier) MarshalJSONPB(*jsonpb.Marshaler) ([]byte, error) {
	return []byte(internal.BuildString(m.ApplicationName, m.ResourceType, m.ResourceId)), nil
}

func (m *Identifier) UnmarhsalJSONPB(_ *jsonpb.Unmarshaler, data []byte) error {
	m.ApplicationName, m.ResourceType, m.ResourceId = internal.ParseString(string(data))
	return nil
}
