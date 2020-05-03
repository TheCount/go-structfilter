package structfilter

import (
	"reflect"
	"unsafe"
)

// Convert converts the specified input value to an output value based on the
// filtered type of the dynamic type of the input. If in is nil, the return
// value is nil. Maps, pointers, and slices whose type definition does
// not involve a structure type will be copied shallowly. Struct fields not
// present in the filtered type are dropped. ToValue also works with recursive
// (self-referential) values.
func (t *T) Convert(in interface{}) interface{} {
	origValue := reflect.ValueOf(in)
	if !origValue.IsValid() {
		return nil
	}
	seenPointers := make(map[unsafe.Pointer]reflect.Value)
	origType := origValue.Type()
	filteredType := t.mapType(origType)
	filteredValue := reflect.New(filteredType).Elem()
	t.convertValue(seenPointers, origValue, filteredValue)
	return filteredValue.Interface()
}

// convertValue converts the specified original value to its filtered
// counterpart and assigns it to filteredValue. The seenPointers map keeps
// track of structure, map, and slice pointers, to properly convert recursive
// values.
func (t *T) convertValue(
	seenPointers map[unsafe.Pointer]reflect.Value,
	origValue, filteredValue reflect.Value,
) {
	// If the original value is stored in an interface, we need to unwrap that
	// first.
	origType := origValue.Type()
	if origType.Kind() == reflect.Interface {
		if !origValue.IsNil() {
			t.convertValue(seenPointers, origValue.Elem(), filteredValue)
		}
		return
	}
	// Common shortcut
	filteredType := filteredValue.Type()
	if origType == filteredType {
		filteredValue.Set(origValue)
		return
	}
	// Avoid infinite recursion
	switch origType.Kind() {
	case reflect.Ptr, reflect.Slice, reflect.Map:
		seenValue, ok := seenPointers[unsafe.Pointer(origValue.Pointer())]
		if ok {
			filteredValue.Set(seenValue)
			return
		}
	}
	// The filtered type may be interface{} to avoid a recursive type definition.
	// In this case we need to allocate an actual value.
	oldFilteredValue := filteredValue
	if filteredType.Kind() == reflect.Interface {
		filteredType = t.mapType(origType)
		filteredValue = reflect.New(filteredType).Elem()
	}

	switch origType.Kind() {
	case reflect.Array:
		for i := 0; i != origType.Len(); i++ {
			origIndexValue := origValue.Index(i)
			filteredIndexValue := filteredValue.Index(i)
			t.convertValue(seenPointers, origIndexValue, filteredIndexValue)
		}
	case reflect.Struct:
		for i := 0; i != origType.NumField(); i++ {
			origStructField := origType.Field(i)
			if _, ok := filteredType.FieldByName(origStructField.Name); !ok {
				continue
			}
			t.convertValue(seenPointers, origValue.Field(i),
				filteredValue.FieldByName(origStructField.Name))
		}
	case reflect.Ptr, reflect.Slice, reflect.Map:
		if !origValue.IsNil() {
			seenPointers[unsafe.Pointer(origValue.Pointer())] = filteredValue
			t.convertPointer(seenPointers, origValue, filteredValue)
		}
	default:
		filteredValue.Set(origValue)
	}
	oldFilteredValue.Set(filteredValue)
}

// convertPointer converts the specified original value to the specified
// filtered value. Both must have the same kind, which must be pointer, slice,
// or map.
// For info on seenPointers, see T.convertValue().
func (t *T) convertPointer(
	seenPointers map[unsafe.Pointer]reflect.Value,
	origValue, filteredValue reflect.Value,
) {
	switch origValue.Kind() {
	case reflect.Ptr:
		filteredValue.Set(reflect.New(filteredValue.Type().Elem()))
		t.convertValue(seenPointers, origValue.Elem(), filteredValue.Elem())
	case reflect.Slice:
		filteredElemType := filteredValue.Type().Elem()
		for i := 0; i != origValue.Len(); i++ {
			filteredElem := reflect.New(filteredElemType).Elem()
			t.convertValue(seenPointers, origValue.Index(i), filteredElem)
			filteredValue.Set(reflect.Append(filteredValue, filteredElem))
		}
	case reflect.Map:
		filteredType := filteredValue.Type()
		filteredValue.Set(reflect.MakeMapWithSize(filteredType, origValue.Len()))
		filteredKeyType := filteredType.Key()
		filteredElemType := filteredType.Elem()
		iter := origValue.MapRange()
		for iter.Next() {
			origKeyValue := iter.Key()
			origElemValue := iter.Value()
			filteredKeyValue := reflect.New(filteredKeyType).Elem()
			filteredElemValue := reflect.New(filteredElemType).Elem()
			t.convertValue(seenPointers, origKeyValue, filteredKeyValue)
			t.convertValue(seenPointers, origElemValue, filteredElemValue)
			filteredValue.SetMapIndex(filteredKeyValue, filteredElemValue)
		}
	}
}
