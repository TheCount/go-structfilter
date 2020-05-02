package structfilter

import (
	"reflect"
	"unsafe"
)

// To converts the specified input value to an output value based on the
// filtered type of the dynamic type of the input. If in is nil, the return
// value is nil. Maps, pointers, and slices whose type definition does
// not involve a structure type will be copied shallowly. Struct fields not
// present in the filtered type are dropped. ToValue also works with recursive
// (self-referential) values.
func (t *T) To(in interface{}) interface{} {
	origValue := reflect.ValueOf(in)
	if !origValue.IsValid() {
		return nil
	}
	seenPointers := make(map[unsafe.Pointer]reflect.Value)
	origType := origValue.Type()
	filteredType := t.mapType(origType)
	filteredValue := reflect.New(filteredType).Elem()
	t.toValue(seenPointers, origValue, filteredValue)
	return filteredValue.Interface()
}

// toValue converts the specified original value to its filtered counterpart
// and assigns it to filteredValue. The seenPointers map keeps track of
// structure, map, and slice pointers, to properly convert recursive values.
func (t *T) toValue(
	seenPointers map[unsafe.Pointer]reflect.Value,
	origValue, filteredValue reflect.Value,
) {
	origType := origValue.Type()
	filteredType := filteredValue.Type()
	if origType == filteredType {
		filteredValue.Set(origValue)
		return
	}
	// Different types means origType is array, struct, pointer, slice, or map.
	// In the latter three cases, we may have seen the value already.
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
			t.toValue(seenPointers, origIndexValue, filteredIndexValue)
		}
	case reflect.Struct:
		for i := 0; i != origType.NumField(); i++ {
			origStructField := origType.Field(i)
			if _, ok := filteredType.FieldByName(origStructField.Name); !ok {
				continue
			}
			t.toValue(seenPointers, origValue.Field(i),
				filteredValue.FieldByName(origStructField.Name))
		}
	default:
		// Pointer, slice, or map: store as seen.
		if !origValue.IsNil() {
			seenPointers[unsafe.Pointer(origValue.Pointer())] = filteredValue
			t.toPointer(seenPointers, origValue, filteredValue)
		}
	}
	oldFilteredValue.Set(filteredValue)
}

// toPointer converts the specified original value to the specified filtered
// value. Both must have the same kind, which must be pointer, slice, or map.
// For info on seenPointers, see T.toValue().
func (t *T) toPointer(
	seenPointers map[unsafe.Pointer]reflect.Value,
	origValue, filteredValue reflect.Value,
) {
	switch origValue.Kind() {
	case reflect.Ptr:
		filteredValue.Set(reflect.New(filteredValue.Type().Elem()))
		t.toValue(seenPointers, origValue.Elem(), filteredValue.Elem())
	case reflect.Slice:
		filteredElemType := filteredValue.Type().Elem()
		for i := 0; i != origValue.Len(); i++ {
			filteredElem := reflect.New(filteredElemType).Elem()
			t.toValue(seenPointers, origValue.Index(i), filteredElem)
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
			t.toValue(seenPointers, origKeyValue, filteredKeyValue)
			t.toValue(seenPointers, origElemValue, filteredElemValue)
			filteredValue.SetMapIndex(filteredKeyValue, filteredElemValue)
		}
	}
}
