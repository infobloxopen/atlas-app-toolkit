package gorm

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

func TestGormFiltering(t *testing.T) {

	tests := []struct {
		rest string
		gorm string
		args []interface{}
		err  error
	}{
		{
			"not(field1 == 'value1' or field2 == 'value2' and field3 != 'value3')",
			"NOT((field1 = ?) OR ((field2 = ?) AND NOT(field3 = ?)))",
			[]interface{}{"value1", "value2", "value3"},
			nil,
		},
		{
			"field1 ~ 'regex'",
			"(field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
		},
		{
			"field1 !~ 'regex'",
			"NOT(field1 ~ ?)",
			[]interface{}{"regex"},
			nil,
		},
		{
			"field1 == 22",
			"(field1 = ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not field1 == 22",
			"NOT(field1 = ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"field1 > 22",
			"(field1 > ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not field1 > 22",
			"NOT(field1 > ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"field1 >= 22",
			"(field1 >= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not field1 >= 22",
			"NOT(field1 >= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"field1 < 22",
			"(field1 < ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not field1 < 22",
			"NOT(field1 < ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"field1 <= 22",
			"(field1 <= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"not field1 <= 22",
			"NOT(field1 <= ?)",
			[]interface{}{22.0},
			nil,
		},
		{
			"field1 == null",
			"(field1 IS NULL)",
			nil,
			nil,
		},
		{
			"field1 != null",
			"NOT(field1 IS NULL)",
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
			"field1 === null",
			"",
			nil,
			&query.UnexpectedSymbolError{},
		},
	}

	for _, test := range tests {
		gorm, args, err := FilterStringToGorm(test.rest)
		assert.Equal(t, test.gorm, gorm)
		assert.Equal(t, test.args, args)
		assert.IsType(t, test.err, err)
	}
}
