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
	Tag reflect.StructTag

	// keep indicates whether the field should be kept.
	keep bool
}

// Name returns the name of this field.
func (f *Field) Name() string {
	return f.name
}

// Remove indicates that this field should not be part of the
// filtered structure. A later filter might cause the field to be included
// after all by calling Keep.
func (f *Field) Remove() {
	f.keep = false
}

// Keep indicates that this field should be part of the filtered structure.
// This is the default. However, calling Keep explicitly may be necessary to
// countermand a Remove call by an earlier filter. A later filter might cause
// the field to be expluded after all by calling Remove again.
func (f *Field) Keep() {
	f.keep = true
}

// newField creates a new struct field based on the original field and field.
func (t *T) newField(
	orig *reflect.StructField, field *Field,
) (reflect.StructField, error) {
	result := reflect.StructField{
		Name:      field.name,
		Tag:       field.Tag,
		Anonymous: orig.Anonymous,
	}
	mappedType, err := t.mapType(orig.Type)
	if err != nil {
		return reflect.StructField{}, err
	}
	if mappedType == nil {
		result.Type = interfaceType
	} else {
		result.Type = mappedType
	}
	return result, nil
}
