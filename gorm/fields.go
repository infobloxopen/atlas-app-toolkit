package gorm

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/golang/protobuf/protoc-gen-go/generator"
	"github.com/jinzhu/gorm"

	"github.com/infobloxopen/atlas-app-toolkit/query"
)

// FieldSelectionStringToGorm is a shortcut to parse a string into FieldSelection struct and
// receive a list of associations to preload.
func FieldSelectionStringToGorm(ctx context.Context, fs string, obj interface{}) ([]string, error) {
	return FieldSelectionToGorm(ctx, query.ParseFieldSelection(fs), obj)
}

// FieldSelectionToGorm receives FieldSelection struct and returns a list of associations to preload.
func FieldSelectionToGorm(ctx context.Context, fs *query.FieldSelection, obj interface{}) ([]string, error) {
	objType := indirectType(reflect.TypeOf(obj))
	if fs.GetFields() == nil {
		return preloadEverything(objType, nil)
	}
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

	toPreload = tidySubPreload(fs.GetFields(), objType, toPreload)
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

func tidySubPreload(subs map[string]*query.Field, objType reflect.Type, subPreload []string) []string {
	edited := false

	exists := make(map[string]string, len(subPreload))
	for _, e := range subPreload {
		sp := strings.Split(e, ".")
		exists[sp[0]] = e
	}

	_, nf := subs["_assoc"]
	if nf {
		numField := objType.NumField()
		for i := 0; i < numField; i++ {
			sf := objType.Field(i)

			if _, ok := exists[sf.Name]; ok {
				continue
			}

			fType := indirectType(sf.Type)
			if isModel(fType) {
				exists[sf.Name] = sf.Name
				edited = true
			}
		}
	}

	for k := range subs {
		if strings.HasPrefix(k, "!") {
			cn := generator.CamelCase(k[1:])
			if _, ok := exists[cn]; ok {
				delete(exists, cn)
				edited = true
			}
		}
	}

	if !edited {
		return subPreload
	}

	var vals []string
	for _, v := range exists {
		vals = append(vals, v)
	}
	return vals
}

func handlePreloads(f *query.Field, objType reflect.Type) (sps []string, rer error) {
	if strings.HasPrefix(f.GetName(), "_") ||
		strings.HasPrefix(f.GetName(), "!") {
		return nil, nil
	}

	ccName := generator.CamelCase(f.GetName())
	sf, ok := objType.FieldByName(ccName)
	if !ok {
		return nil, nil
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
		toPreload = append(toPreload, subPreload...)
	}

	toPreload = tidySubPreload(f.GetSubs(), fType, toPreload)

	for i, e := range toPreload {
		toPreload[i] = ccName + "." + e
	}
	return append(toPreload, ccName), nil
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
	return nil, fmt.Errorf("Cannot preload empty association")
}
