package structfilter

import (
	"errors"
	"reflect"
	"testing"
)

// errFilter is an error returned by filters for testing.
var errFilter = errors.New("test filter error")

// nopFilter is a filter which doesn't do anything.
func nopFilter(*Field) error {
	return nil
}

// errorFilter is a filter function which always returns an error.
func errorFilter(*Field) error {
	return errFilter
}

// TestNopFilter tests filtering without any actual changes to fields.
func TestNopFilter(t *testing.T) {
	// Single filter
	filter := New(nopFilter)
	if _, err := filter.ReflectType(reflect.TypeOf(SimpleStruct{})); err != nil {
		t.Errorf("Error with simple struct and nop filter: %s", err)
	}
	// Multiple filters
	filter = New(nopFilter, nopFilter)
	if _, err := filter.ReflectType(reflect.TypeOf(SimpleStruct{})); err != nil {
		t.Errorf("Error with simple struct and two nop filters: %s", err)
	}
}

// TestErrFilter tests a filter which always returns an error.
func TestErrFilter(t *testing.T) {
	// Single filter
	filter := New(errorFilter)
	if _, err := filter.ReflectType(reflect.TypeOf(struct{}{})); err != nil {
		t.Errorf("Expected no filter call for empty struct, got: %s", err)
	}
	if _, err := filter.ReflectType(reflect.TypeOf(SimpleStruct{})); err == nil {
		t.Error("Expected error with error filter")
	} else if !errors.Is(err, errFilter) {
		t.Errorf("Expected filter error, got: %s", err)
	}
	// Multiple filters
	filter = New(nopFilter, errorFilter)
	if _, err := filter.ReflectType(reflect.TypeOf(SimpleStruct{})); err == nil {
		t.Error("Expected error with multiple filters")
	} else if !errors.Is(err, errFilter) {
		t.Errorf("Expected filter error, got: %s", err)
	}
}
