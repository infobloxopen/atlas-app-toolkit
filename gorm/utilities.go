package gorm

import (
	"context"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"
	"time"

	"gorm.io/datatypes"

	"github.com/golang/protobuf/proto"
	"github.com/jinzhu/inflection"
	"gorm.io/gorm/schema"

	"github.com/infobloxopen/atlas-app-toolkit/v2/rpc/resource"
	"github.com/infobloxopen/atlas-app-toolkit/v2/util"
)

// HandleFieldPath converts fieldPath to appropriate db string for use in where/order by clauses
// according to obj GORM model. If fieldPath cannot be found in obj then original fieldPath is returned
// to allow tables joined by a third party.
// If association join is required to resolve the field path then it's name is returned as a second return value.
func HandleFieldPath(ctx context.Context, fieldPath []string, obj interface{}) (string, string, error) {
	if len(fieldPath) > 2 {
		return "", "", fmt.Errorf("Field path longer than 2 is not supported")
	}
	dbPath, err := fieldPathToDBName(fieldPath, obj)
	if err != nil {
		switch err.(type) {
		case *EmptyFieldPathError:
			return "", "", err
		default:
			return strings.Join(fieldPath, "."), "", nil
		}
	}
	if len(fieldPath) == 2 {
		return dbPath, util.Camel(fieldPath[0]), nil
	}
	return dbPath, "", nil
}

// HandleJSONFiledPath translate field path to JSONB path for postgres jsonb
func HandleJSONFieldPath(ctx context.Context, fieldPath []string, obj interface{}, values ...string) (string, string, error) {
	operator := "#>>"
	if isRawJSON(values...) {
		operator = "#>"
	}

	dbPath, err := fieldPathToDBName(fieldPath[:1], obj)
	if err != nil {
		switch err.(type) {
		case *EmptyFieldPathError:
			return "", "", err
		default:
			dbPath = fieldPath[0]
		}
	}

	if len(fieldPath) == 1 {
		return dbPath, "", nil
	}

	return fmt.Sprintf("%s %s '{%s}'", dbPath, operator, strings.Join(fieldPath[1:], ",")), "", nil
}

func isRawJSON(values ...string) bool {
	if len(values) == 0 {
		return false
	}

	for _, v := range values {
		//TODO: this is a very poor check to prevent unexpected errors from Database engine consider to make full validation
		//TODO: also we need return an error if json invalid to prevent database error for json parsing
		v = strings.TrimSpace(v)
		if !strings.HasPrefix(v, "{") || !strings.HasSuffix(v, "}") {
			return false
		}
	}

	return true
}

// TODO: add supprt for embeded objects
func IsJSONCondition(ctx context.Context, fieldPath []string, obj interface{}) bool {
	fieldName := util.Camel(fieldPath[0])
	objType := indirectType(reflect.TypeOf(obj))
	field, ok := objType.FieldByName(fieldName)
	if !ok {
		return false
	}

	fType := field.Type

	// we need to check each of indirect types as well
	//
	// this is required as if type alias is used it will be considered
	// as an underlying type (which might be a slice/other indirect type)
	// like for gorm.io/datatypes.JSON
	for isIndirectType(fType) {
		fInterface := reflect.Zero(fType).Interface()
		switch fInterface.(type) {
		case *datatypes.JSON:
			return true
		}

		fType = fType.Elem()
	}

	return false
}

func fieldPathToDBName(fieldPath []string, obj interface{}) (string, error) {
	objType := indirectType(reflect.TypeOf(obj))
	pathLength := len(fieldPath)
	assocAlias := ""
	for i, part := range fieldPath {
		if !isModel(objType) {
			return "", fmt.Errorf("%s: non-last field of %s field path should be a model", objType, fieldPath)
		}
		sf, ok := objType.FieldByName(util.Camel(part))
		if !ok {
			return "", fmt.Errorf("Cannot find field %s in %s", part, objType)
		}
		if i < pathLength-1 {
			objType = indirectType(sf.Type)
			assocAlias = part
		} else {
			if isModel(indirectType(sf.Type)) {
				return "", fmt.Errorf("%s: last field of %s field path should be a model", objType, fieldPath)
			}
			var dbPrefix string
			if assocAlias != "" {
				dbPrefix = assocAlias
			} else {
				dbPrefix = tableName(objType)
			}
			return dbPrefix + "." + columnName(&sf), nil
		}
	}
	return "", &EmptyFieldPathError{}
}

func tableName(t reflect.Type) string {
	table := reflect.Zero(t).Interface()
	if tn, ok := table.(tableNamer); ok {
		return tn.TableName()
	}
	return inflection.Plural(ToDBName(t.Name()))
}

func columnName(sf *reflect.StructField) string {
	ex, tagCol := gormTag(sf, "column")
	if ex {
		return tagCol
	}
	return ToDBName(sf.Name)
}

func gormTag(sf *reflect.StructField, tag string) (bool, string) {
	return extractTag(sf, "gorm", tag)
}

func atlasTag(sf *reflect.StructField, tag string) (bool, string) {
	return extractTag(sf, "atlas", tag)
}

func extractTag(sf *reflect.StructField, tag string, subTag string) (bool, string) {
	gormTags := strings.Split(sf.Tag.Get(tag), ";")
	for _, t := range gormTags {
		var key, value string
		keyValue := strings.Split(t, ":")
		switch len(keyValue) {
		case 2:
			value = keyValue[1]
			fallthrough
		case 1:
			key = keyValue[0]
		}
		if strings.ToLower(key) == strings.ToLower(subTag) {
			return true, value
		}
	}
	return false, ""
}

type tableNamer interface {
	TableName() string
}

func indirectType(t reflect.Type) reflect.Type {
	for isIndirectType(t) {
		t = t.Elem()
	}

	return t
}

func isIndirectType(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return true
	default:
		return false
	}
}

func indirectValue(val reflect.Value) reflect.Value {
	for {
		switch val.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Array:
			val = val.Elem()
		default:
			return val
		}
	}
}

func isModel(t reflect.Type) bool {
	kind := t.Kind()
	_, isValuer := reflect.Zero(t).Interface().(driver.Valuer)
	if (kind == reflect.Struct || kind == reflect.Slice) && !isValuer && t != reflect.TypeOf(time.Time{}) {
		return true
	}
	return false
}

func isProtoMessage(t reflect.Type) bool {
	_, isProtoMessage := reflect.Zero(t).Interface().(proto.Message)
	return isProtoMessage
}

func isIdentifier(t reflect.Type) bool {
	_, isIdentifier := reflect.Zero(t).Interface().(resource.Identifier)
	return isIdentifier
}

type EmptyFieldPathError struct {
}

func (e *EmptyFieldPathError) Error() string {
	return fmt.Sprintf("Empty field path is not allowed")
}

var namingStrategy = schema.NamingStrategy{}

func ToDBName(s string) string {
	return namingStrategy.ColumnName("", s)
}

func camelCase(v string) string {
	sp := strings.Split(v, "_")
	r := make([]string, len(sp))
	for i, v := range sp {
		r[i] = strings.ToUpper(v[:1]) + v[1:]
	}

	return strings.Join(r, "")
}
