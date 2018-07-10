package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type Entity struct {
	Field1       int
	Field2       int
	Field3       int
	NestedEntity NestedEntity
}

type NestedEntity struct {
	NestedField1 int
	NestedField2 int
}

func TestGormFiltering(t *testing.T) {

	tests := []struct {
		rest  string
		gorm  string
		args  []interface{}
		assoc map[string]struct{}
		err   error
	}{
		{
			"not(field1 == 'value1' or field2 == 'value2' and field3 != 'value3')",
			"NOT((entities.field1 = ?) OR ((entities.field2 = ?) AND NOT(entities.field3 = ?)))",
			[]interface{}{"value1", "value2", "value3"},
			nil,
			nil,
		},
		{
			"field1 ~ 'regex'",
			"(entities.field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
			nil,
		},
		{
			"field1 !~ 'regex'",
			"NOT(entities.field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
			nil,
		},
		{
			"field1 == 22",
			"(entities.field1 = ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"not field1 == 22",
			"NOT(entities.field1 = ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"field1 > 22",
			"(entities.field1 > ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"not field1 > 22",
			"NOT(entities.field1 > ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"field1 >= 22",
			"(entities.field1 >= ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"not field1 >= 22",
			"NOT(entities.field1 >= ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"field1 < 22",
			"(entities.field1 < ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"not field1 < 22",
			"NOT(entities.field1 < ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"field1 <= 22",
			"(entities.field1 <= ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"not field1 <= 22",
			"NOT(entities.field1 <= ?)",
			[]interface{}{22.0},
			nil,
			nil,
		},
		{
			"field1 == null",
			"(entities.field1 IS NULL)",
			nil,
			nil,
			nil,
		},
		{
			"field1 != null",
			"NOT(entities.field1 IS NULL)",
			nil,
			nil,
			nil,
		},
		{
			"field1 != null",
			"NOT(entities.field1 IS NULL)",
			nil,
			nil,
			nil,
		},
		{
			"nested_entity.nested_field1 == 11 and nested_entity.nested_field2 == 22",
			"((nested_entities.nested_field1 = ?) AND (nested_entities.nested_field2 = ?))",
			[]interface{}{11.0, 22.0},
			map[string]struct{}{"NestedEntity": struct{}{}},
			nil,
		},
		{
			"field1 === null",
			"",
			nil,
			nil,
			&query.UnexpectedSymbolError{},
		},
	}

	for _, test := range tests {
		gorm, args, assoc, err := FilterStringToGorm(test.rest, &Entity{})
		assert.Equal(t, test.gorm, gorm)
		assert.Equal(t, test.args, args)
		assert.Equal(t, test.assoc, assoc)
		assert.IsType(t, test.err, err)
	}
}
