package gorm

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/golang/protobuf/protoc-gen-go/generator"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

func FieldSelectionStringToGorm(fs string, obj interface{}) ([]string, []string, error) {
	return FieldSelectionToGorm(query.ParseFieldSelection(fs), obj)
}

func FieldSelectionToGorm(fs *query.FieldSelection, obj interface{}) ([]string, []string, error) {
	if fs == nil {
		return nil, nil, nil
	}
	objType := indirectType(reflect.ValueOf(obj).Type())
	var toSelect, toPreload []string
	fieldNames := getSortedFieldNames(fs.GetFields())
	for _, fieldName := range fieldNames {
		f := fs.GetFields()[fieldName]
		if f.GetSubs() == nil {
			sf, ok := objType.FieldByName(generator.CamelCase(f.GetName()))
			if !ok {
				return nil, nil, fmt.Errorf("Cannot find field %s in %s", f.GetName(), objType)
			}
			fType := indirectType(sf.Type)
			if isModel(fType) {
				toPreload = append(toPreload, generator.CamelCase(f.GetName()))
			} else {
				dbName, err := FieldPathToDBName([]string{f.GetName()}, obj)
				if err != nil {
					return nil, nil, err
				}
				toSelect = append(toSelect, dbName)
			}
		} else {
			subPreload, err := handlePreloads(f, objType)
			if err != nil {
				return nil, nil, err
			}
			toPreload = append(toPreload, subPreload...)
		}
	}
	return toSelect, toPreload, nil
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
