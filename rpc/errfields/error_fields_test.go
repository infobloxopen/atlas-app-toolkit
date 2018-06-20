package errfields

import (
	"reflect"
	"testing"
)

func TestAddFields(t *testing.T) {
	m := &FieldInfo{}

	// Check add field.

	m.AddField("target", "msg1")
	reflect.DeepEqual(
		m,
		&FieldInfo{
			Fields: map[string]*StringListValue{
				"target": &StringListValue{
					Values: []string{"msg1"},
				},
			},
		},
	)

	// Check append another message to the same target.

	m.AddField("target", "msg2")
	reflect.DeepEqual(
		m,
		&FieldInfo{
			Fields: map[string]*StringListValue{
				"target": &StringListValue{
					Values: []string{"msg1", "msg2"},
				},
			},
		},
	)

	// Check add another field.

	m.AddField("target-2", "msg1")
	reflect.DeepEqual(
		m,
		&FieldInfo{
			Fields: map[string]*StringListValue{
				"target": &StringListValue{
					Values: []string{"msg1", "msg2"},
				},
				"target-2": &StringListValue{
					Values: []string{"msg1"},
				},
			},
		},
	)
}
