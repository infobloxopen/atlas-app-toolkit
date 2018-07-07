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
// If association join is required to resolve the field path then it's name is returned as a second return value.
func HandleFieldPath(fieldPath []string, obj interface{}) (string, string, error) {
	if len(fieldPath) > 2 {
		return "", "", fmt.Errorf("Field path longer than 2 is not supported")
	}
	dbPath, err := fieldPathToDBName(fieldPath, obj)
	if err != nil {
		switch err.(type) {
		case *EmptyFieldPathError, *InvalidGormTagError:
			return "", "", err
		default:
			return strings.Join(fieldPath, "."), "", nil
		}
	}
	if len(fieldPath) == 2 {
		return dbPath, generator.CamelCase(fieldPath[0]), nil
	}
	return dbPath, "", nil
}

// ToDBName converts fieldPath to appropriate db string for use in where/order by clauses
// according to obj GORM model.
func fieldPathToDBName(fieldPath []string, obj interface{}) (string, error) {
	objType := indirectType(reflect.ValueOf(obj).Type())
	pathLength := len(fieldPath)
	for i, part := range fieldPath {
		if !isModel(objType) {
			return "", fmt.Errorf("%s: non-last field of %s field path should be a model", objType, fieldPath)
		}
		sf, ok := objType.FieldByName(generator.CamelCase(part))
		if !ok {
			return "", fmt.Errorf("Cannot find field %s in %s", part, objType)
		}
		if i < pathLength-1 {
			objType = indirectType(sf.Type)
		} else {
			if isModel(indirectType(sf.Type)) {
				return "", fmt.Errorf("%s: last field of %s field path should be a model", objType, fieldPath)
			}
			return tableName(objType) + "." + columnName(&sf), nil
		}
	}
	return "", &EmptyFieldPathError{}
}

func tableName(t reflect.Type) string {
	table := reflect.Zero(t).Interface()
	if tn, ok := table.(tableNamer); ok {
		return tn.TableName()
	}
	return inflection.Plural(jgorm.ToDBName(t.Name()))
}

func columnName(sf *reflect.StructField) string {
	ex, tagCol := gormTag(sf, "column")
	if ex {
		return tagCol
	}
	return jgorm.ToDBName(sf.Name)
}

func gormTag(sf *reflect.StructField, tag string) (bool, string) {
	gormTags := strings.Split(sf.Tag.Get("gorm"), ";")
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
		if strings.ToLower(key) == strings.ToLower(tag) {
			return true, value
		}
	}
	return false, ""
}

type tableNamer interface {
	TableName() string
}

func indirectType(t reflect.Type) reflect.Type {
	for {
		switch t.Kind() {
		case reflect.Ptr, reflect.Slice, reflect.Array:
			t = t.Elem()
		default:
			return t
		}
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

func JoinInfo(obj interface{}, assoc string) (string, []string, []string, error) {
	objType := indirectType(reflect.ValueOf(obj).Type())
	sf, ok := objType.FieldByName(assoc)
	if !ok {
		return "", nil, nil, fmt.Errorf("Cannot find field %s in %s", assoc, objType)
	}
	ok, assocKey := gormTag(&sf, "association_foreignkey")
	if !ok {
		return "", nil, nil, fmt.Errorf("association_foreignkey tag is absent in %s", objType)
	}
	assocKeys := strings.Split(assocKey, ",")
	ok, fKey := gormTag(&sf, "foreignkey")
	if !ok {
		return "", nil, nil, fmt.Errorf("foreignkey tag is absent in %s", objType)
	}
	fKeys := strings.Split(fKey, ",")

	if len(assocKeys) != len(fKeys) {
		return "", nil, nil, fmt.Errorf(`%s: the number of association keys is not equal to the number
of foreign keys in %s association`, objType, assoc)
	}

	assocType := indirectType(sf.Type)

	_, childTableName, dbAssocKeys, dbFKeys, err := parseParentChildAssoc(objType, assocType, assocKeys, fKeys)
	if err != nil {
		parentTableName, _, dbAssocKeys, dbFKeys, err := parseParentChildAssoc(assocType, objType, assocKeys, fKeys)
		if err != nil {
			return "", nil, nil, err
		}
		return parentTableName, dbFKeys, dbAssocKeys, nil
	}
	return childTableName, dbAssocKeys, dbFKeys, nil
}

func parseParentChildAssoc(parent reflect.Type, child reflect.Type, assocKeys []string, fKeys []string) (string, string, []string, []string, error) {
	parentTableName := tableName(parent)
	childTableName := tableName(parent)
	var dbAssocKeys, dbFKeys []string
	for _, k := range assocKeys {
		sf, ok := parent.FieldByName(k)
		if !ok {
			return "", "", nil, nil, fmt.Errorf("Association key %s is not found in %s", k, parent)
		}
		dbAssocKeys = append(dbAssocKeys, parentTableName+"."+columnName(&sf))
	}
	for _, k := range fKeys {
		sf, ok := parent.FieldByName(k)
		if !ok {
			return "", "", nil, nil, fmt.Errorf("Foreign key %s is not found in %s", k, child)
		}
		dbFKeys = append(dbFKeys, childTableName+"."+columnName(&sf))
	}
	return parentTableName, childTableName, dbAssocKeys, dbFKeys, nil
}
