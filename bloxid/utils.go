package bloxid

import (
	"reflect"
)

// isNilInterface returns whether the interface parameter is nil
func isNilInterface(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}
