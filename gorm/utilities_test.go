package gorm

import (
	"context"
	"reflect"
	"testing"

	"time"

	"github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
)

type Human struct {
	Name  string
	Age   uint32 `gorm:"column:years"`
	Child *Child
}

func (p Human) TableName() string {
	return "db_humans"
}

type Child struct {
	Name string
}

func TestHandleFieldPath(t *testing.T) {
	tests := []struct {
		fieldPath []string
		dbName    string
		assoc     string
		err       bool
	}{
		{[]string{"name"}, "db_humans.name", "", false},
		{[]string{"age"}, "db_humans.years", "", false},
		{[]string{"child", "name"}, "child.name", "Child", false},
		{[]string{}, "", "", true},
	}
	for _, test := range tests {
		dbName, assoc, err := HandleFieldPath(context.Background(), test.fieldPath, &Human{})
		if test.err {
			assert.Equal(t, "", dbName)
			assert.Equal(t, "", assoc)
			assert.NotNil(t, err)
		} else {
			assert.Equal(t, test.dbName, dbName)
			assert.Equal(t, test.assoc, assoc)
			assert.Nil(t, err)
		}
	}
}

type JSONEntity struct {
	Tags *postgres.Jsonb
}

func TestHandleJSONFieldPath(t *testing.T) {
	tests := []struct {
		name      string
		fieldPath []string
		wantErr   bool
		wantDB    string
	}{
		{
			name:      "valid single segment",
			fieldPath: []string{"tags"},
			wantErr:   false,
			wantDB:    "json_entities.tags",
		},
		{
			name:      "valid nested json path",
			fieldPath: []string{"tags", "location"},
			wantErr:   false,
			wantDB:    "json_entities.tags #>> '{location}'",
		},
		{
			name:      "valid nested json path with hyphen",
			fieldPath: []string{"tags", "my-key"},
			wantErr:   false,
			wantDB:    "json_entities.tags #>> '{my-key}'",
		},
		{
			name:      "sql injection via quote in segment",
			fieldPath: []string{"tags", "loc'; DROP TABLE users; --"},
			wantErr:   true,
		},
		{
			name:      "sql injection via quote in first segment",
			fieldPath: []string{"tags' OR 1=1; --", "key"},
			wantErr:   true,
		},
		{
			name:      "sql injection via curly brace",
			fieldPath: []string{"tags", "a}','b'),('"},
			wantErr:   true,
		},
		{
			name:      "empty segment rejected",
			fieldPath: []string{"tags", ""},
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbName, _, err := HandleJSONFieldPath(context.Background(), tt.fieldPath, &JSONEntity{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantDB, dbName)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello_world", "HelloWorld"},
		{"single", "Single"},
		{"a_b_c", "ABC"},
		{"hello__world", "HelloWorld"}, // double underscore
		{"_leading", "Leading"},        // leading underscore
		{"trailing_", "Trailing"},      // trailing underscore
		{"__double__", "Double"},       // multiple empty segments
		{"already", "Already"},         // no underscores
		{"", ""},                       // empty string
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			assert.NotPanics(t, func() {
				got := camelCase(tt.input)
				assert.Equal(t, tt.want, got)
			})
		})
	}
}

func TestIsModel(t *testing.T) {
	for tName, tCase := range map[string]struct {
		input reflect.Type
		want  bool
	}{
		"time.Time":  {reflect.TypeOf(time.Time{}), false},
		"*time.Time": {reflect.TypeOf(&time.Time{}), false},
	} {
		t.Run(tName, func(t *testing.T) {
			if got, want := isModel(tCase.input), tCase.want; got != want {
				t.Errorf("got %v; want %v", got, want)
			}
		})
	}
}
