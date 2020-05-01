package structfilter

import (
	"reflect"
)

// interfaceType is the reflect type of a plain interface{}.
var interfaceType = reflect.TypeOf(new(interface{})).Elem()

// getStructType returns a reflect type representing a struct type StructType
// for input types of the form StructType, *StructType, **StructType, etc.
// The second return value is the number of asterisks. If t does not have this
// form, getStructType returns nil, -1.
func getStructType(t reflect.Type) (reflect.Type, int) {
	for depth := 0; ; depth++ {
		switch t.Kind() {
		case reflect.Struct:
			return t, depth
		case reflect.Ptr:
			t = t.Elem()
		default:
			return nil, -1
		}
	}
}
