package errfields

import (
	"encoding/json"
)

func (m *FieldInfo) AddField(target string, msg string) {
	if m.Fields == nil {
		m.Fields = map[string]*StringListValue{}
	}

	if m.Fields[target] == nil {
		m.Fields[target] = &StringListValue{Values: []string{}}
	}

	m.Fields[target].Values = append(m.Fields[target].Values, msg)
}

// MarshalJSON implements json.Marshaler.
func (fi *FieldInfo) MarshalJSON() ([]byte, error) {
	fiMap := map[string][]string{}

	for k, v := range fi.Fields {
		var descArr []string
		if v != nil {
			descArr = make([]string, len(v.Values))
			for i, desc := range v.Values {
				descArr[i] = desc
			}
		}

		fiMap[k] = descArr
	}

	return json.Marshal(&fiMap)
}

// UnmarshalJSON implements json.Unmarshaler.
func (fi *FieldInfo) UnmarshalJSON(data []byte) error {
	dst := map[string][]string{}
	if err := json.Unmarshal(data, &dst); err != nil {
		return err
	}

	for k, v := range dst {
		for _, vVal := range v {
			fi.AddField(k, vVal)
		}
	}

	return nil
}
