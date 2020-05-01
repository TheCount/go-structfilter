package structfilter

import (
	"reflect"
	"testing"
)

// SimpleStruct is a simple, non-recursive structure type for testing.
type SimpleStruct struct {
	Bool       bool
	Int        int
	Int8       int8
	Int16      int16
	Int32      int32
	Int64      int64
	Uint       uint
	Uint8      uint8
	Uint16     uint16
	Uint32     uint32
	Uint64     uint64
	Uintptr    uintptr
	Float32    float32
	Float64    float64
	Complex64  complex64
	Complex128 complex128
	Array      [42]int
	Chan       chan SimpleStruct
	Func       func(SimpleStruct) SimpleStruct
	Interface  interface{}
	Map        map[int]int
	Ptr        *int
	Slice      []int
	String     string
	Struct     struct{}
}

// StructWithUnexportedFields is a structure with some unexported fields for
// testing.
type StructWithUnexportedFields struct {
	Exported1   int
	unexported1 int
	Exported2   int
	unexported2 int
	Exported3   int
	unexported3 int
}

// RecursiveStruct is a recursively defined structure for testing.
type RecursiveStruct struct {
	Array [42]*RecursiveStruct
	Map   map[*RecursiveStruct]RecursiveStruct
	Ptr   *RecursiveStruct
	Slice []RecursiveStruct
}

// nested is a simple structure type to nest into other structure definitions
// for testing.
type nested struct {
	Field int
}

// NestedStruct is a simple, non-recursively nested structure type.
type NestedStruct struct {
	Field nested
	Array [42]nested
	Map   map[nested]nested
	Ptr   *nested
	Slice []nested
}

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
