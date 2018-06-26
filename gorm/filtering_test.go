package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

type Entity struct {
}

func TestGormFiltering(t *testing.T) {

	tests := []struct {
		rest string
		gorm string
		args []interface{}
		err  error
	}{
		{
			"not(entities.field1 == 'value1' or entities.field2 == 'value2' and entities.field3 != 'value3')",
			"NOT((entities.field1 = ?) OR ((entities.field2 = ?) AND NOT(entities.field3 = ?)))",
			[]interface{}{"value1", "value2", "value3"},
			nil,
		},
		{
			"entities.field1 ~ 'regex'",
			"(entities.field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
		},
		{
			"entities.field1 !~ 'regex'",
			"NOT(entities.field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
		},
		{
			"entities.field1 == 22",
			"(entities.field1 = ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not entities.field1 == 22",
			"NOT(entities.field1 = ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"entities.field1 > 22",
			"(entities.field1 > ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not entities.field1 > 22",
			"NOT(entities.field1 > ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"entities.field1 >= 22",
			"(entities.field1 >= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not entities.field1 >= 22",
			"NOT(entities.field1 >= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"entities.field1 < 22",
			"(entities.field1 < ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not entities.field1 < 22",
			"NOT(entities.field1 < ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"entities.field1 <= 22",
			"(entities.field1 <= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not entities.field1 <= 22",
			"NOT(entities.field1 <= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"entities.field1 == null",
			"(entities.field1 IS NULL)",
			nil,
			nil,
		},
		{
			"entities.field1 != null",
			"NOT(entities.field1 IS NULL)",
			nil,
			nil,
		},
		{
			"",
			"",
			nil,
			nil,
		},
		{
			"entities.field1 === null",
			"",
			nil,
			&query.UnexpectedSymbolError{},
		},
	}

	for _, test := range tests {
		gorm, args, err := FilterStringToGorm(test.rest, &Entity{})
		assert.Equal(t, test.gorm, gorm)
		assert.Equal(t, test.args, args)
		assert.IsType(t, test.err, err)
	}
}
