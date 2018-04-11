package pb

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/wrappers"
	"reflect"
	"strings"
)

// Strips a pointer, or pointer(-to-pointer)^* to base object if needed
func indirect(reflectValue reflect.Value) reflect.Value {
	for reflectValue.Kind() == reflect.Ptr {
		reflectValue = reflectValue.Elem()
	}
	return reflectValue
}

// Strips a pointer, slice, pointer-to-slice, slice-of-pointers,
// pointer-to-slice-of-pointers-to-pointers, etc... to base type if needed
func indirectType(reflectType reflect.Type) reflect.Type {
	for reflectType.Kind() == reflect.Ptr || reflectType.Kind() == reflect.Slice {
		reflectType = reflectType.Elem()
	}
	return reflectType
}

//----- Types, and zero values to avoid having to recreate them every time
var pStringValueType = reflect.TypeOf(&wrappers.StringValue{})
var pStringValueZeroValue = reflect.Zero(pStringValueType)

var pUInt32ValueType = reflect.TypeOf(&wrappers.UInt32Value{})
var pUInt32ValueZeroValue = reflect.Zero(pUInt32ValueType)

var exString = ""
var pStringType = reflect.TypeOf(&exString)
var pStringZeroValue = reflect.Zero(pStringType)

var exUInt32 = uint32(0)
var pUInt32Type = reflect.TypeOf(&exUInt32)
var pUInt32ZeroValue = reflect.Zero(pUInt32Type)

type typeToType struct {
	from reflect.Type
	to   reflect.Type
}

// Cache for different conversions, dest fields in order, -1 if no dest
var fieldMapsByType = make(map[typeToType][]int)

// Convert Copies data between fields at ORM and service levels.
// Works under the assumption that any WKT fields in proto map to * fields at ORM.
func Convert(source interface{}, dest interface{}) error {
	// If dest object is unaddressable, that won't work. Unfortunately, a code
	// error that will only be caught at runtime
	toObject := indirect(reflect.ValueOf(dest))
	if toObject.CanAddr() == false {
		return fmt.Errorf("Dest, type %s, is unaddressable", reflect.TypeOf(dest))
	}
	if indirectType(reflect.TypeOf(source)).Kind() != reflect.Struct {
		return fmt.Errorf("Cannot convert a non-struct")
	}
	destType := toObject.Type()
	fromObject := indirect(reflect.ValueOf(source))
	fromType := fromObject.Type()

	// Check for mapping, populate mapping if not already present
	fieldMap, exists := fieldMapsByType[typeToType{fromType, destType}]
	if !exists {
		for i := 0; i < fromType.NumField(); i++ {
			found := false
			for j := 0; j < destType.NumField(); j++ {
				if strings.EqualFold(fromType.Field(i).Name, destType.Field(j).Name) {
					found = true
					fieldMap = append(fieldMap, j)
					break
				}
			}
			// Store -1 if no dest corresponds to field
			if !found {
				fieldMap = append(fieldMap, -1)
			}
		}
		fieldMapsByType[typeToType{fromType, destType}] = fieldMap
	}

	for i := 0; i < fromType.NumField(); i++ {
		if fieldMap[i] == -1 {
			continue
		}
		to := toObject.Field(fieldMap[i])
		if to.IsValid() {
			fromFieldDesc := fromType.Field(i)
			fromData := fromObject.Field(i)

			switch fromFieldDesc.Type {
			case to.Type(): // Matching type
				to.Set(fromData)
			case pStringValueType: // WKT *StringValue{} --> *string
				if fromData.IsNil() || !fromData.IsValid() {
					to.Set(pStringZeroValue)
				} else {
					value := fromData.Elem().Field(0).String()
					to.Set(reflect.ValueOf(&value))
				}
			case pStringType: // *string --> WKT *StringValue{}
				if fromData.IsNil() || !fromData.IsValid() {
					to.Set(pStringValueZeroValue)
				} else {
					strValue := fromData.Elem().String()
					to.Set(reflect.ValueOf(&wrappers.StringValue{strValue}))
				}
			case pUInt32ValueType: // WKT *UInt32Value{} --> *uint32
				if fromData.IsNil() || !fromData.IsValid() {
					to.Set(pUInt32ZeroValue)
				} else {
					value := uint32(fromData.Elem().Field(0).Uint())
					to.Set(reflect.ValueOf(&value))
				}
			case pUInt32Type: // *uint32 --> WKT *UInt32Value{}
				if fromData.IsNil() || !fromData.IsValid() {
					to.Set(pUInt32ValueZeroValue)
				} else {
					intValue := uint32(fromData.Elem().Uint())
					to.Set(reflect.ValueOf(&wrappers.UInt32Value{intValue}))
				}
				//Additional WKTs to be used should be included here
			default:
				kind := fromFieldDesc.Type.Kind()
				if kind == reflect.Slice &&
					indirectType(fromFieldDesc.Type).Kind() == reflect.Struct &&
					indirectType(to.Type()).Kind() == reflect.Struct { // Copy slice one at a time

					len := fromData.Len()
					to.Set(reflect.MakeSlice(to.Type(), len, len))
					for k := 0; k < len; k++ {
						dest := to.Index(k)
						if dest.Kind() == reflect.Ptr {
							dest.Set(reflect.New(indirectType(dest.Type())).Elem().Addr())
						}
						err := Convert(fromData.Index(k).Interface(), dest.Addr().Interface())
						if err != nil {
							fmt.Printf("%s", err.Error())
						}
					}
				} else if kind == reflect.Struct && !fromData.IsNil() { // A nested struct
					err := Convert(fromData.Interface(), to.Addr().Interface())
					if err != nil {
						fmt.Printf("%s", err.Error())
					}
				} else if kind == reflect.Int32 && to.Type().Kind() == reflect.Int32 { // Probably an enum
					to.Set(reflect.ValueOf(int32(fromData.Int())))
				}
			}
		}
	}
	return nil
}
