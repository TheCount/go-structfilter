package structfilter

import (
	"errors"
	"reflect"
	"regexp"
	"testing"
)

// StructKeepRemove is a structure type for testing the RemoveFieldFilter
// function.
type StructKeepRemove struct {
	Keep1   int
	Remove1 int
	Keep2   int
	Remove2 int
}

// StructTag is a structure type for testing the InsertTagFilter function.
type StructTag struct {
	TagMe         int
	NoThanks      int
	TagMeNotAgain int `test:"alreadypresent"`
}

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

// TestOccasionalErrFilter tests filters throwing an occasional error.
func TestOccasionalErrFilter(t *testing.T) {
	i := 0
	filter := New(func(*Field) error {
		i++
		if i%4 == 0 {
			return errFilter
		}
		return nil
	})
	if _, err := filter.Convert(SimpleStruct{}); err == nil {
		t.Error("Expected error in value conversion with occasional error filter")
	}
}

// TestNilRemoveFieldFilter tests RemoveFieldFilter with a nil matcher.
func TestNilRemoveFieldFilter(t *testing.T) {
	filter := New(RemoveFieldFilter(nil))
	filtered, err := filter.Convert(StructKeepRemove{})
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.ValueOf(filtered)
	if value.NumField() != 4 {
		t.Error("Expected no removed fields with nil matcher")
	}
}

// TestRemoveFieldFilter tests RemoveFieldFilter.
func TestRemoveFieldFilter(t *testing.T) {
	re := regexp.MustCompile("^Remove.*$")
	filter := New(RemoveFieldFilter(re))
	filtered, err := filter.Convert(StructKeepRemove{})
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.ValueOf(filtered)
	if value.NumField() != 2 {
		t.Errorf("Expected 2 remaining fields, got %d", value.NumField())
	}
	if value.FieldByName("Remove1").IsValid() {
		t.Error("Field that should have been removed is still present")
	}
}

// TestNilInsertTagFilter tests InsertTagFilter with a nil matcher.
func TestNilInsertTagFilter(t *testing.T) {
	filter := New(InsertTagFilter(nil, `test:"foo"`))
	filtered, err := filter.Convert(StructTag{})
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.ValueOf(filtered)
	tag0 := value.Type().Field(0).Tag
	tag1 := value.Type().Field(1).Tag
	if tag0 != "" || tag1 != "" {
		t.Errorf("Expected unchanged tags, got `%s` and `%s` instead", tag0, tag1)
	}
}

// TestInsertTagFilter tests InsertTagFilter.
func TestInsertTagFilter(t *testing.T) {
	re := regexp.MustCompile("^TagMe.*$")
	// First test with bad tag format.
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic on bad tag format")
			}
		}()
		InsertTagFilter(re, "badtag")
	}()
	// Now proper test
	const tag = `test:"inserted"`
	filter := New(InsertTagFilter(re, tag))
	filtered, err := filter.Convert(StructTag{})
	if err != nil {
		t.Fatal(err)
	}
	value := reflect.ValueOf(filtered)
	tag0 := value.Type().Field(0).Tag
	tag1 := value.Type().Field(1).Tag
	tag2 := value.Type().Field(2).Tag
	if tag0 != tag {
		t.Errorf("Expected inserted tag `%s`, got `%s`", tag, tag0)
	}
	if tag1 != "" {
		t.Errorf("Expected tag to be empty, got `%s`", tag1)
	}
	if tag2.Get(`test`) == "inserted" {
		t.Error("Already present tag was overwritten")
	}
}
