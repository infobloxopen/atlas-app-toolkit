package gorm

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/atlas-app-toolkit/query"
	"github.com/infobloxopen/atlas-app-toolkit/rpc/resource"
)

type Entity struct {
	Field1       int
	Field2       int
	Field3       int
	FieldString  string
	NestedEntity NestedEntity
	Id           string
	Ref          *string
}

type EntityProto struct {
	Id  *resource.Identifier
	Ref *resource.Identifier
}

func (*EntityProto) Reset() {
}

func (*EntityProto) ProtoMessage() {
}

func (*EntityProto) String() string {
	return "Entity"
}

func (*EntityProto) ToORM(ctx context.Context) (Entity, error) {
	id := "convertedid"
	ref := "convertedref"
	return Entity{Id: id, Ref: &ref}, nil
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
			"field_string > 'str'",
			"(entities.field_string > ?)",
			[]interface{}{"str"},
			nil,
			nil,
		},
		{
			"field_string >= 'str'",
			"(entities.field_string >= ?)",
			[]interface{}{"str"},
			nil,
			nil,
		},
		{
			"field_string < 'str'",
			"(entities.field_string < ?)",
			[]interface{}{"str"},
			nil,
			nil,
		},
		{
			"field_string <= 'str'",
			"(entities.field_string <= ?)",
			[]interface{}{"str"},
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
			"((nested_entity.nested_field1 = ?) AND (nested_entity.nested_field2 = ?))",
			[]interface{}{11.0, 22.0},
			map[string]struct{}{"NestedEntity": {}},
			nil,
		},
		{
			"field1 === null",
			"",
			nil,
			nil,
			&query.UnexpectedSymbolError{},
		},
		{
			"id == 'id' and ref == 'ref'",
			"((entities.id = ?) AND (entities.ref = ?))",
			[]interface{}{"convertedid", "convertedref"},
			nil,
			nil,
		},
		{
			"id := 'ID'",
			"(lower(entities.id) = lower(?))",
			[]interface{}{"convertedid"},
			nil,
			nil,
		},
		{
			"not(id := 'sOmeId')",
			"NOT(lower(entities.id) = lower(?))",
			[]interface{}{"convertedid"},
			nil,
			nil,
		},
		{
			"id in ['sOmeId', 'egegeg']",
			"(entities.id  IN (?, ?))",
			[]interface{}{"convertedid", "convertedid"},
			nil,
			nil,
		},
		{
			"not(id in ['sOmeId', 'egegeg'])",
			"(entities.id NOT IN (?, ?))",
			[]interface{}{"convertedid", "convertedid"},
			nil,
			nil,
		},
		{
			"id in [1, 2]",
			"(entities.id  IN (?, ?))",
			[]interface{}{1.0, 2.0},
			nil,
			nil,
		},
		{
			"not(id in [1, 2])",
			"(entities.id NOT IN (?, ?))",
			[]interface{}{1.0, 2.0},
			nil,
			nil,
		},
	}

	for _, test := range tests {
		gorm, args, assoc, err := FilterStringToGorm(context.Background(), test.rest, &Entity{}, &EntityProto{})
		assert.Equal(t, test.gorm, gorm)
		assert.Equal(t, test.args, args)
		assert.Equal(t, test.assoc, assoc)
		assert.IsType(t, test.err, err)
	}
}
