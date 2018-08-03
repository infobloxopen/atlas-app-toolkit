package gorm

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	fieldmask "google.golang.org/genproto/protobuf/field_mask"
)

// MergeWithMask will take the fields of `source` that are included as
// paths in `mask` and write them to the corresponding fields of `dest`
func MergeWithMask(source, dest interface{}, mask *fieldmask.FieldMask) error {
	if mask == nil || len(mask.Paths) == 0 {
		return nil
	}
	if source == nil {
		return errors.New("Source object is nil")
	}
	if dest == nil {
		return errors.New("Destination object is nil")
	}
	if reflect.TypeOf(source) != reflect.TypeOf(dest) {
		return errors.New("Types of source and destination objects do not match")
	}
pathsloop:
	for _, fullpath := range mask.GetPaths() {
		subpaths := strings.Split(fullpath, ".")
		srcVal := reflect.ValueOf(source).Elem()
		dstVal := reflect.ValueOf(dest).Elem()
		for _, path := range subpaths {
			for dstVal.Kind() == reflect.Ptr {
				if dstVal.IsNil() {
					dstVal.Set(reflect.New(dstVal.Type().Elem()))
				}
				dstVal = dstVal.Elem()
				srcVal = srcVal.Elem()
			}
			// For safety, skip paths that will cause a panic to call FieldByName on
			if dstVal.Kind() != reflect.Struct {
				continue pathsloop
			}
			srcVal = srcVal.FieldByName(path)
			dstVal = dstVal.FieldByName(path)
			if !srcVal.IsValid() || !dstVal.IsValid() {
				return fmt.Errorf("Field path %q doesn't exist in type %s",
					fullpath, reflect.TypeOf(source))
			}
		}
		for dstVal.Kind() == reflect.Ptr {
			if dstVal.IsNil() {
				dstVal.Set(reflect.New(dstVal.Type().Elem()))
			}
			dstVal = dstVal.Elem()
			srcVal = srcVal.Elem()
		}
		dstVal.Set(srcVal)
	}
	return nil
}
