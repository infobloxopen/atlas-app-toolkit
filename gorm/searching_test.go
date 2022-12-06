package gorm

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/infobloxopen/protoc-gen-gorm/types"
)

func NewSearchEntity() SearchEntity {
	return SearchEntity{}
}

type SearchEntity struct {
	FieldInt32  int32
	FieldString string
	FieldTime   *time.Time
	FieldBool   bool
	DDITagsOld  *types.JSONValue
}

func TestGetFullTextSearchDBMask(t *testing.T) {
	entity1 := NewSearchEntity()
	entity1.FieldInt32 = 1
	entity1.FieldString = "Testing"
	time1 := time.Date(2021, time.Month(2), 21, 1, 10, 30, 0, time.UTC)
	entity1.FieldTime = &time1
	entity1.FieldBool = true
	entityTagsOld := types.JSONValue{Value: `{"tag1":"value1", "tag2":"value2"}`}
	entity1.DDITagsOld = &entityTagsOld

	entity2 := NewSearchEntity()
	entity2.FieldInt32 = 2
	entity2.FieldString = "Ned Flanders"
	time2 := time.Date(2022, time.Month(4), 4, 9, 0, 0, 0, time.UTC)
	entity2.FieldTime = &time2
	entity2.FieldBool = false
	entityTagsOld2 := types.JSONValue{Value: `{"ned":"flanders", "lisa":"simpson"}`}
	entity2.DDITagsOld = &entityTagsOld2

	object := []*SearchEntity{
		&entity1,
		&entity2,
	}

	separator := ","

	val := reflect.ValueOf(entity1)
	fields := make([]string, val.Type().NumField())
	for i := 0; i < val.Type().NumField(); i++ {
		fields[i] = val.Type().Field(i).Name
	}

	assert.Equal(t, fields, []string{"FieldInt32", "FieldString", "FieldTime", "FieldBool", "DDITagsOld"})

	// Expect mask to contain the fields FieldInt32, FieldString, FieldTime, and FieldBool
	mask := GetFullTextSearchDBMask(entity1, fields, separator)
	expectedMask := "FieldInt32 || ',' || FieldString || ',' || replace(FieldString, '@', ' ') || ',' ||" +
		" replace(FieldString, '.', ' ') || ',' || coalesce(to_char(FieldTime, 'MM/DD/YY HH:MI pm'), '') || ',' ||" +
		" FieldBool || ',' || "
	assert.Equal(t, mask, expectedMask)

	// Expect to panic if function used with array
	defer func() {
		if r := recover(); r == nil {
			fmt.Println("Recovered panic:", r)
		}
	}()
	GetFullTextSearchDBMask(object, fields, separator)
	t.Errorf("The code did not panic")
}
