package errfields

import ()

func (m *FieldInfo) AddField(target string, msg string) {
	if m.Fields == nil {
		m.Fields = map[string]*StringListValue{}
	}

	if m.Fields[target] == nil {
		m.Fields[target] = &StringListValue{Values: []string{}}
	}

	m.Fields[target].Values = append(m.Fields[target].Values, msg)
}
