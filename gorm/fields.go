package gorm

import (
	"context"
	"fmt"
	"github.com/infobloxopen/atlas-app-toolkit/util"
	"reflect"
	"sort"
	"strings"

	"github.com/jinzhu/gorm"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

// DefaultFieldSelectionConverter performs default convertion for FieldSelection collection operator
type DefaultFieldSelectionConverter struct{}

// FieldSelectionStringToGorm is a shortcut to parse a string into FieldSelection struct and
// receive a list of associations to preload.
func FieldSelectionStringToGorm(ctx context.Context, fs string, obj interface{}) ([]string, error) {
	c := &DefaultFieldSelectionConverter{}
	return c.FieldSelectionToGorm(ctx, query.ParseFieldSelection(fs), obj)
}

// FieldSelectionToGorm receives FieldSelection struct and returns a list of associations to preload.
func (converter *DefaultFieldSelectionConverter) FieldSelectionToGorm(ctx context.Context, fs *query.FieldSelection, obj interface{}) ([]string, error) {
	objType := indirectType(reflect.TypeOf(obj))
	selectedFields := fs.GetFields()
	if selectedFields == nil {
		return preloadEverything(objType, nil)
	}
	var toPreload []string
	fieldNames := getSortedFieldNames(selectedFields)
	for _, fieldName := range fieldNames {
		f := selectedFields[fieldName]
		subPreload, err := handlePreloads(f, objType)
		if err != nil {
			return nil, err
		}
		toPreload = append(toPreload, subPreload...)
	}
	return toPreload, nil
}

func preloadEverything(objType reflect.Type, path []reflect.Type) ([]string, error) {
	if !isModel(objType) {
		return nil, fmt.Errorf("%s is not a model", objType)
	}
	numField := objType.NumField()
	var toPreload []string
fields:
	for i := 0; i < numField; i++ {
		sf := objType.Field(i)
		fType := indirectType(sf.Type)
		for _, e := range path {
			if fType == e {
				continue fields
			}
		}
		if isModel(fType) {
			if ok, flag := gormTag(&sf, "preload"); ok && flag == "false" {
				continue
			}
			subPreload, err := preloadEverything(fType, append(path, objType))
			if err != nil {
				return nil, err
			}
			for i, e := range subPreload {
				subPreload[i] = sf.Name + "." + e
			}
			toPreload = append(toPreload, subPreload...)
			toPreload = append(toPreload, sf.Name)
		}
	}
	return toPreload, nil
}

func handlePreloads(f *query.Field, objType reflect.Type) ([]string, error) {
	queryFieldName := f.GetName()
	fmt.Printf("Query name = %v\n", queryFieldName)

	findFieldNameFunc := func(objType reflect.Type, s string) (string, reflect.StructField, bool) {
		var findFieldByName reflect.StructField
		var ok bool
		camelQueryFieldName := util.Camel(s)
		// find field as camel string
		findFieldByName, ok = objType.FieldByName(camelQueryFieldName)
		if ok {
			fmt.Printf("Found field by name: %v\n", camelQueryFieldName)
			return camelQueryFieldName, findFieldByName, ok
		}

		for _, subString := range strings.Split(s, "_") {
			// TODO: if there are multiple same substrings, only first one is replaced
			fieldName := strings.Replace(s, subString, strings.ToUpper(subString), 1)
			fieldName = util.Camel(fieldName)
			findFieldByName, ok = objType.FieldByName(fieldName)
			if ok {
				fmt.Printf("Found field by name: %v\n", fieldName)
				return fieldName, findFieldByName, ok
			}
		}

		return "", findFieldByName, false
	}

	fieldTypeName, sf, ok := findFieldNameFunc(objType, queryFieldName)

	if !ok {
		return nil, nil
	}

	fType := indirectType(sf.Type)
	fieldSubs := f.GetSubs()

	if fieldSubs == nil {
		if isModel(fType) {
			return []string{fieldTypeName}, nil
		} else {
			return nil, nil
		}
	}
	if !isModel(fType) {
		return nil, fmt.Errorf("%s is expected to be a model, but got %s ", queryFieldName, fType)
	}
	var toPreload []string
	fieldNames := getSortedFieldNames(fieldSubs)
	for _, fieldName := range fieldNames {
		subField := fieldSubs[fieldName]
		subPreload, err := handlePreloads(subField, fType)
		if err != nil {
			return nil, err
		}
		for i, e := range subPreload {
			subPreload[i] = fieldTypeName + "." + e
		}
		toPreload = append(toPreload, subPreload...)
	}
	return append(toPreload, fieldTypeName), nil
}

func getSortedFieldNames(fields map[string]*query.Field) []string {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func preload(db *gorm.DB, obj interface{}, assoc string) (*gorm.DB, error) {
	objType := indirectType(reflect.TypeOf(obj))
	if !isModel(objType) {
		return nil, fmt.Errorf("%s is not a model", objType)
	}
	assocPath := strings.Split(assoc, ".")
	pathLength := len(assocPath)
	for i, part := range assocPath {
		sf, ok := objType.FieldByName(part)
		if !ok {
			return nil, fmt.Errorf("Cannot find %s in %s", part, objType)
		}
		objType = indirectType(sf.Type)
		if !isModel(objType) {
			return nil, fmt.Errorf("%s is not a model", objType)
		}
		if i == pathLength-1 {
			ok, pos := atlasTag(&sf, "position")
			if !ok {
				return db.Preload(assoc), nil
			} else {
				return db.Preload(assoc, func(db *gorm.DB) *gorm.DB {
					return db.Order(gorm.ToDBName(pos))
				}), nil
			}
		}
	}
	return nil, fmt.Errorf("cannot preload empty association")
}
