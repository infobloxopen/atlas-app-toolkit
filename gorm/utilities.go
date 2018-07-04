package gorm

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	jgorm "github.com/jinzhu/gorm"
	"github.com/jinzhu/inflection"
)

// HandleFieldPath converts fieldPath to appropriate db string for use in where/order by clauses
// according to obj GORM model. If fieldPath cannot be found in obj then original fieldPath is returned
// to allow tables joined by a third party.
func HandleFieldPath(fieldPath []string, obj interface{}) (string, error) {
	if len(fieldPath) > 2 {
		return "", fmt.Errorf("Field path longer than 2 is not supported")
	}
	dbPath, err := FieldPathToDBName(fieldPath, obj)
	if err != nil {
		switch err.(type) {
		case *EmptyFieldPathError, *InvalidGormTagError:
			return "", err
		default:
			return strings.Join(fieldPath, "."), nil
		}
	}
	return dbPath, nil
}

// ToDBName converts fieldPath to appropriate db string for use in where/order by clauses
// according to obj GORM model.
func FieldPathToDBName(fieldPath []string, obj interface{}) (string, error) {
	objType := indirectType(reflect.ValueOf(obj).Type())
	pathLength := len(fieldPath)
	for i, part := range fieldPath {
		if !isModel(objType) {
			return "", fmt.Errorf("%s: non-last field of a field path should be a model", objType)
		}
		sf, ok := objType.FieldByName(generator.CamelCase(part))
		if !ok {
			return "", fmt.Errorf("Cannot find field %s in %s", part, objType)
		}
		if i < pathLength-1 {
			objType = indirectType(sf.Type)
		} else {
			table := reflect.Zero(objType).Interface()
			var tableName string
			if tn, ok := table.(tableNamer); ok {
				tableName = tn.TableName()
			} else {
				tableName = inflection.Plural(jgorm.ToDBName(objType.Name()))
			}
			gormTags := strings.Split(sf.Tag.Get("gorm"), ";")
			for _, tag := range gormTags {
				var key, value string
				keyValue := strings.Split(tag, ":")
				l := len(keyValue)
				switch {
				case l > 2:
					return "", &InvalidGormTagError{tag}
				case l == 2:
					value = keyValue[1]
					fallthrough
				case l == 1:
					key = keyValue[0]
				}

				if strings.ToLower(key) == "column" {
					return tableName + "." + value, nil
				}
			}
			return tableName + "." + part, nil
		}
	}
	return "", &EmptyFieldPathError{}
}

type tableNamer interface {
	TableName() string
}

func indirectType(t reflect.Type) reflect.Type {
	kind := t.Kind()
	switch kind {
	case reflect.Ptr, reflect.Slice, reflect.Array:
		return t.Elem()
	default:
		return t
	}
}

func isModel(t reflect.Type) bool {
	kind := t.Kind()
	_, isValuer := reflect.Zero(t).Interface().(driver.Valuer)
	if (kind == reflect.Struct || kind == reflect.Slice) && !isValuer {
		return true
	}
	return false
}

type EmptyFieldPathError struct {
}

func (e *EmptyFieldPathError) Error() string {
	return fmt.Sprintf("Empty field path is not allowed")
}

type InvalidGormTagError struct {
	Tag string
}

func (e *InvalidGormTagError) Error() string {
	return fmt.Sprintf("%s: gorm tag is invalid", e.Tag)
}
