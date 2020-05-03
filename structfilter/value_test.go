package structfilter

import (
	"reflect"
	"regexp"
	"testing"
)

// TestToSimpleValue tests simple value filtering.
func TestToSimpleValue(t *testing.T) {
	filter := New()
	if filter.To(nil) != nil {
		t.Error("Nil to value does not yield nil")
	}
	if v, ok := filter.To(42).(int); !ok {
		t.Error("Integer to value does not yield integer")
	} else if v != 42 {
		t.Error("Integer to value does not yield same result")
	}
	var simple SimpleStruct
	filtered := filter.To(simple)
	filteredValue := reflect.ValueOf(filtered)
	if filteredValue.Kind() != reflect.Struct {
		t.Error("Expected filtered value to be a struct")
	}
}

// TestToValueWithUnexportedFields tests value filtering with unexported fields.
func TestToValueWithUnexportedFields(t *testing.T) {
	filter := New()
	filtered := filter.To(StructWithUnexportedFields{
		Exported1:   1,
		unexported1: -1,
		Exported2:   2,
		unexported2: -2,
		Exported3:   3,
		unexported3: -3,
	})
	filteredValue := reflect.ValueOf(filtered)
	if filteredValue.NumField() != 3 {
		t.Fatalf("Expected 3 fields in filtered value, got %d",
			filteredValue.NumField())
	}
	for i := 0; i != 3; i++ {
		if filteredValue.Field(i).Interface().(int) != i+1 {
			t.Errorf("Wrong value for filtered field Exported%d", i+1)
		}
	}
}

// TestToValueRecursive tests value filtering with a recursive value.
func TestToValueRecursive(t *testing.T) {
	filter := New()
	rec1 := RecursiveStruct{}
	rec2 := RecursiveStruct{
		Map:   make(map[*RecursiveStruct]RecursiveStruct),
		Ptr:   &rec1,
		Slice: make([]RecursiveStruct, 0),
	}
	rec2.Array[21] = &rec1
	rec2.Map[&rec1] = rec2
	rec2.Map[&rec2] = rec1
	rec2.Slice = append(rec2.Slice, rec1, rec2)
	rec1.Ptr = &rec2
	filtered1 := filter.To(rec1)
	filtered2 := filter.To(rec2)
	filteredValue1 := reflect.ValueOf(filtered1)
	filteredValue2 := reflect.ValueOf(filtered2)
	if filteredValue1.Kind() != reflect.Struct ||
		filteredValue2.Kind() != reflect.Struct {
		t.Error("Expected filtered recursive values to be structs")
	}
}

// TestToValueInterface tests value filtering with an intervening interface.
func TestToValueInterface(t *testing.T) {
	filter := New(RemoveFieldFilter(regexp.MustCompile("^U.*$")))
	orig := SimpleStruct{
		Interface: SimpleStruct{
			Interface: 42,
		},
	}
	filtered := filter.To(orig)
	filteredValue := reflect.ValueOf(filtered)
	if filteredValue.FieldByName("Uint").IsValid() {
		t.Error("Top level field Uint should have been removed")
	}
	filteredValue = filteredValue.FieldByName("Interface").Elem()
	if filteredValue.FieldByName("Uint").IsValid() {
		t.Error("Nested field Uint should have been removed")
	}
}
