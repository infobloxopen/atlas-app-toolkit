package gorm

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/generator"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

// FieldSelectionStringToGorm is a shortcut to parse a string into FieldSelection struct and
// receive a list of associations to preload.
func FieldSelectionStringToGorm(fs string, obj interface{}) ([]string, error) {
	return FieldSelectionToGorm(query.ParseFieldSelection(fs), obj)
}

// FieldSelectionToGorm receives FieldSelection struct and returns a list of associations to preload.
func FieldSelectionToGorm(fs *query.FieldSelection, obj interface{}) ([]string, error) {
	if fs == nil {
		return nil, nil
	}
	objType := indirectType(reflect.ValueOf(obj).Type())
	var toPreload []string
	fieldNames := getSortedFieldNames(fs.GetFields())
	for _, fieldName := range fieldNames {
		f := fs.GetFields()[fieldName]
		subPreload, err := handlePreloads(f, objType)
		if err != nil {
			return nil, err
		}
		toPreload = append(toPreload, subPreload...)
	}
	return toPreload, nil
}

func handlePreloads(f *query.Field, objType reflect.Type) ([]string, error) {
	sf, ok := objType.FieldByName(generator.CamelCase(f.GetName()))
	if !ok {
		return nil, fmt.Errorf("Cannot find field %s in %s", f.GetName(), objType)
	}
	fType := indirectType(sf.Type)
	if f.GetSubs() == nil {
		if isModel(fType) {
			return []string{generator.CamelCase(f.GetName())}, nil
		} else {
			return nil, nil
		}
	}
	if !isModel(fType) {
		return nil, fmt.Errorf("%s is expected to be a model, but got %s ", f.GetName(), fType)
	}
	var toPreload []string
	fieldNames := getSortedFieldNames(f.GetSubs())
	for _, fieldName := range fieldNames {
		subField := f.GetSubs()[fieldName]
		subPreload, err := handlePreloads(subField, fType)
		if err != nil {
			return nil, err
		}
		for i, e := range subPreload {
			subPreload[i] = generator.CamelCase(f.GetName()) + "." + e
		}
		toPreload = append(toPreload, subPreload...)
	}
	if toPreload == nil {
		return []string{generator.CamelCase(f.GetName())}, nil
	}
	return toPreload, nil
}

func getSortedFieldNames(fields map[string]*query.Field) []string {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
