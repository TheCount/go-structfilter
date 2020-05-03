package structfilter

import (
	"reflect"
	"testing"
)

// TestRemoveKeep tests countermanding a Remove with a Keep.
func TestRemoveKeep(t *testing.T) {
	filter := New(func(f *Field) error {
		f.Remove()
		return nil
	}, func(f *Field) error {
		f.Keep()
		return nil
	})
	orig := SimpleStruct{}
	filtered := filter.To(orig)
	origValue := reflect.ValueOf(orig)
	filteredValue := reflect.ValueOf(filtered)
	if origValue.NumField() != filteredValue.NumField() {
		t.Error("Expected all fields to be kept")
	}
}
