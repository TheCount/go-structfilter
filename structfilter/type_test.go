package structfilter

import (
	"reflect"
	"testing"
	"time"
)

// TestNilReflectType tests the ReflectType method with a nil argument.
func TestNilReflectType(t *testing.T) {
	filter := New()
	if _, err := filter.ReflectType(nil); err == nil {
		t.Error("Expected error on nil orig type")
	}
}

// TestNonStructReflectType tests the ReflectType method with a non-struct
// argument.
func TestNonStructReflectType(t *testing.T) {
	filter := New()
	if _, err := filter.ReflectType(reflect.TypeOf(0)); err == nil {
		t.Error("Expected error on non-struct orig type")
	}
}

// TestPointerToStructReflectType tests the ReflectType method with pointer
// to struct types.
func TestPointerToStructReflectType(t *testing.T) {
	pStruct := &struct{}{}
	pStructType := reflect.TypeOf(pStruct)
	ppStructType := reflect.TypeOf(&pStruct)
	filter := New()
	if _, err := filter.ReflectType(pStructType); err != nil {
		t.Errorf("Unexpected error on *struct orig type: %s", err)
	}
	if _, err := filter.ReflectType(ppStructType); err == nil {
		t.Error("Expected error on **struct orig type")
	}
}

// TestEmptyStructReflectType tests the ReflectType method with an empty struct.
func TestEmptyStructReflectType(t *testing.T) {
	filter := New()
	filtered, err := filter.ReflectType(reflect.TypeOf(struct{}{}))
	if err != nil {
		t.Fatalf("Error filtering empty struct: %s", err)
	}
	if filtered.Kind() != reflect.Struct {
		t.Fatalf("Expected filtered type to be struct type, got %s",
			filtered.Kind())
	}
	if filtered.NumField() != 0 {
		t.Errorf("Expected zero fields in filtered struct, got %d",
			filtered.NumField())
	}
}

// TestSimpleStructReflectType tests the ReflectType method with a simple
// structure.
func TestSimpleStructReflectType(t *testing.T) {
	origType := reflect.TypeOf(SimpleStruct{})
	filter := New()
	filtered, err := filter.ReflectType(origType)
	if err != nil {
		t.Fatalf("Error filtering simple struct: %s", err)
	}
	if filtered.Kind() != reflect.Struct {
		t.Fatalf("Expected filtered SimpleStruct to be struct type, got %s",
			filtered.Kind())
	}
	if filtered.NumField() != origType.NumField() {
		t.Fatal("Expected orig and filtered struct to have same number of fields "+
			"with empty filter, got %d != %d",
			origType.NumField(), filtered.NumField())
	}
	for i := 0; i != filtered.NumField(); i++ {
		origField := origType.Field(i)
		filteredField := filtered.Field(i)
		if origField.Name != filteredField.Name {
			t.Errorf("Original and filtered field have different names: '%s' != '%s'",
				origField.Name, filteredField.Name)
		}
		if filteredField.Anonymous {
			t.Errorf("Filtered field for '%s' is anonymous", origField.Name)
		}
		if origField.Type != filteredField.Type {
			t.Error("Different types from SimpleStruct")
		}
	}
}

// TestStructWithUnexportedFieldsReflectType tests the ReflectType method with
// a structure containing unexported fields.
func TestStructWithUnexportedFieldsReflectType(t *testing.T) {
	filter := New()
	filtered, err :=
		filter.ReflectType(reflect.TypeOf(StructWithUnexportedFields{}))
	if err != nil {
		t.Fatalf("Error filtering struct with unexported fields: %s", err)
	}
	if filtered.NumField() != 3 {
		t.Error("Unexported fields have not been filtered out")
	}
}

// TestRecursiveStructReflectType tests the ReflectType method with a
// recursively defined structure.
func TestRecursiveStructReflectType(t *testing.T) {
	filter := New()
	filtered, err := filter.ReflectType(reflect.TypeOf(RecursiveStruct{}))
	if err != nil {
		t.Fatalf("Error filtering recursive structure: %s", err)
	}
	for i := 0; i != filtered.NumField(); i++ {
		field := filtered.Field(i)
		if field.Type != interfaceType {
			t.Errorf("Type of field '%s' of recursive struct is not interface{}",
				field.Name)
		}
	}
}

// TestReflectTypeTwice tests whether filtering the same type twice yields the
// same result.
func TestReflectTypeTwice(t *testing.T) {
	filter := New()
	filtered1, err1 := filter.ReflectType(reflect.TypeOf(RecursiveStruct{}))
	filtered2, err2 := filter.ReflectType(reflect.TypeOf(RecursiveStruct{}))
	if err1 != nil || err2 != nil {
		t.Fatalf("Unable to create identical filtered types: %s/%s", err1, err2)
	}
	if filtered1 != filtered2 {
		t.Error("Filtered types from same original differ")
	}
}

// TestNestedReflectType tests the ReflectType method with a simple
// structure.
func TestNestedReflectType(t *testing.T) {
	origType := reflect.TypeOf(NestedStruct{})
	filter := New()
	filtered, err := filter.ReflectType(origType)
	if err != nil {
		t.Fatalf("Error filtering nested struct: %s", err)
	}
	if filtered.NumField() != origType.NumField() {
		t.Fatal("Expected orig and filtered struct to have same number of fields "+
			"with empty filter, got %d != %d",
			origType.NumField(), filtered.NumField())
	}
	for i := 0; i != filtered.NumField(); i++ {
		origField := origType.Field(i)
		filteredField := filtered.Field(i)
		if origField.Type.Kind() != filteredField.Type.Kind() {
			t.Error("Different kinds from NestedStruct")
		}
	}
}

// TestNilUnfilteredType tests the UnfilteredType and UnfilteredReflectType
// methods with nil arguments.
func TestNilUnfilteredType(t *testing.T) {
	filter := New()
	oldnum := len(filter.types)
	filter.UnfilteredType(nil)
	filter.UnfilteredReflectType(nil)
	if len(filter.types) != oldnum {
		t.Errorf("Unexpected new nil unfiltered types (%d, was %d)",
			len(filter.types), oldnum)
	}
}

// TestCuriousUnfilteredType tests UnfilteredType with self-referential types
// not involving a struct.
func TestCuriousUnfilteredType(t *testing.T) {
	filter := New()
	oldnum := len(filter.types)
	filter.UnfilteredType(new(CuriousPointer))
	filter.UnfilteredType(new(CuriousSlice))
	filter.UnfilteredType(new(CuriousMap))
	if len(filter.types) != oldnum {
		t.Errorf("Unexpected new curious unfiltered types (%d, was %d)",
			len(filter.types), oldnum)
	}
}

// TestSafeStruct tests UnfilteredType for a structure type.
func TestSafeStruct(t *testing.T) {
	filter := New()
	filter.UnfilteredType(time.Time{})
	orig := SafeStruct{
		SafeField: time.Now(),
	}
	filtered, err := filter.Convert(orig)
	if err != nil {
		t.Fatalf("Error filtering struct with safe field: %s", err)
	}
	filteredValue := reflect.ValueOf(filtered)
	if !filteredValue.Field(0).Interface().(time.Time).Equal(orig.SafeField) {
		t.Error("Values of unfiltered type do not match")
	}
}
