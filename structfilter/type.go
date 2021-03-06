package structfilter

import (
	"errors"
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

// ReflectType allows direct filtering of structure types as presented by the
// golang reflect package. orig must be a structure type, or a pointer to a
// a structure type. On success, the returned filtered
// type is always a structure type, not a pointer type.
func (t *T) ReflectType(orig reflect.Type) (reflect.Type, error) {
	if orig == nil {
		return nil, errors.New("orig is nil")
	}
	structType, depth := getStructType(orig)
	if structType == nil {
		return nil, errors.New("not a struct type or pointer to struct type")
	}
	if depth > 1 {
		return nil, errors.New("at most one pointer indirection allowed")
	}
	if filteredType, ok := t.types[structType]; ok {
		return filteredType, nil
	}
	return t.filterType(structType)
}

// mapType maps the specified original type to a matching generated type.
// If orig cannot be mapped because it is recursive, nil is returned
// instead.
func (t *T) mapType(orig reflect.Type) (reflect.Type, error) {
	switch orig.Kind() {
	case reflect.Array:
		elem, err := t.mapType(orig.Elem())
		if err != nil {
			return nil, err
		}
		if elem == nil {
			return nil, nil
		}
		if elem == orig.Elem() {
			return orig, nil
		}
		return reflect.ArrayOf(orig.Len(), elem), nil
	case reflect.Interface:
		// Generated type has no methods, so we need to downgrade all interfaces
		// to plain interface{}.
		return interfaceType, nil
	case reflect.Map:
		key, err := t.mapType(orig.Key())
		if err != nil {
			return nil, err
		}
		elem, err := t.mapType(orig.Elem())
		if err != nil {
			return nil, err
		}
		if key == nil || elem == nil {
			return nil, nil
		}
		if key == orig.Key() && elem == orig.Elem() {
			return orig, nil
		}
		return reflect.MapOf(key, elem), nil
	case reflect.Ptr:
		elem, err := t.mapType(orig.Elem())
		if err != nil {
			return nil, err
		}
		if elem == nil {
			return nil, nil
		}
		if elem == orig.Elem() {
			return orig, nil
		}
		return reflect.PtrTo(elem), nil
	case reflect.Slice:
		elem, err := t.mapType(orig.Elem())
		if err != nil {
			return nil, err
		}
		if elem == nil {
			return nil, nil
		}
		if elem == orig.Elem() {
			return orig, nil
		}
		return reflect.SliceOf(elem), nil
	case reflect.Struct:
		elem, ok := t.types[orig]
		if ok {
			return elem, nil // elem == nil if recursive
		}
		elem, err := t.filterType(orig)
		if err != nil {
			return nil, err
		}
		return elem, nil
	default:
		return orig, nil
	}
}
