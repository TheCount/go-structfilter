package structfilter

import (
	"reflect"
)

// Field describes a struct field in the newly generated structure.
type Field struct {
	// name is the name of the new field. It is identical to the old name and
	// cannot be changed.
	name string

	// tag is the tag of the new struct field.
	tag reflect.StructTag

	// keep indicates whether the field should be kept.
	keep bool
}

// newField creates a new struct field based on the original field and field.
func (t *T) newField(
	orig *reflect.StructField, field *Field,
) reflect.StructField {
	result := reflect.StructField{
		Name:      field.name,
		Tag:       field.tag,
		Anonymous: orig.Anonymous,
	}
	mappedType := t.mapType(orig.Type)
	if mappedType == nil {
		result.Type = interfaceType
	} else {
		result.Type = mappedType
	}
	return result
}
