package gorm

import (
	"reflect"
	"time"
)

// GetFullTextSearchDBMask ...
func GetFullTextSearchDBMask(object interface{}, fields []string, separator string) string {
	mask := ""
	objectVal := indirectValue(reflect.ValueOf(object))
	if objectVal.Kind() != reflect.Struct {
		return mask
	}
	fieldsSize := len(fields)
	for i, fieldName := range fields {
		fieldVal := objectVal.FieldByName(camelCase(fieldName))
		if !fieldVal.IsValid() {
			continue
		}
		underlyingVal := indirectValue(fieldVal)
		if !underlyingVal.IsValid() {
			switch fieldVal.Interface().(type) {
			case *time.Time:
				underlyingVal = fieldVal
			default:
				continue
			}
		}
		switch underlyingVal.Interface().(type) {
		case int32:
			mask += fieldName
		case string:
			mask += fieldName
			mask += " || '" + separator + "' || "
			mask += "replace(" + fieldName + ", '@', ' ')"
			mask += " || '" + separator + "' || "
			mask += "replace(" + fieldName + ", '.', ' ')"
		case *time.Time:
			mask += "coalesce(to_char(" + fieldName + ", 'MM/DD/YY HH:MI pm'), '')"
		case bool:
			mask += fieldName
		default:
			continue
		}
		if i != fieldsSize-1 {
			mask += " || '" + separator + "' || "
		}
	}

	return mask
}

// FormFullTextSearchQuery ...
func FormFullTextSearchQuery(mask string) string {
	fullTextSearchQuery := "to_tsvector('simple', " + mask + ") @@ to_tsquery('simple', ?)"
	return fullTextSearchQuery
}
