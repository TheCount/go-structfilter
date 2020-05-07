package structfilter

import "time"

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

// CuriousPointer is self-referential pointer type.
type CuriousPointer *CuriousPointer

// CuriousSlice is a self-referential slice type.
type CuriousSlice []CuriousSlice

// CuriousMap is a self-referential map type.
type CuriousMap map[*CuriousMap]CuriousMap

// SafeStruct is a structure type with a safe field.
type SafeStruct struct {
	SafeField time.Time
}
