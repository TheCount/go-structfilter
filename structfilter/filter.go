package structfilter

import (
	"errors"
	"fmt"
	"reflect"
)

// Func is a function type for altering or removing fields as they are
// inserted into a new structure. Whenever a Filter function returns a non-nil
// error, it will be reported back to the original caller of a filter method.
type Func func(*Field) error

// T is the main structfilter type.
//
// The methods of T are unsafe for concurrent use.
type T struct {
	// filter is the filter function this structfilter uses for filtering.
	filter Func

	// types maps original structure types to their filtered structure type.
	types map[reflect.Type]reflect.Type
}

// ReflectType allows direct filtering of structure types as presented by the
// golang reflect package. orig must be a structure type, or a pointer type
// which eventually indirects to a structure. On success, the returned filtered
// type is always a structure type, not a pointer type.
func (t *T) ReflectType(orig reflect.Type) (reflect.Type, error) {
	if orig == nil {
		return nil, errors.New("orig is nil")
	}
	structType, _ := getStructType(orig)
	if structType == nil {
		return nil, errors.New("not a struct type")
	}
	if filteredType, ok := t.types[structType]; ok {
		return filteredType, nil
	}
	return t.filterType(structType)
}

// filterType returns the filtered type for the specified original type.
// orig must not be in t.types yet.
func (t *T) filterType(orig reflect.Type) (filtered reflect.Type, err error) {
	t.types[orig] = nil // reserve our spot
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic attempting to create filtered type: %v", r)
		}
		if err != nil {
			delete(t.types, orig)
		}
	}()
	filteredFields := make([]reflect.StructField, 0, orig.NumField())
	for i := 0; i != orig.NumField(); i++ {
		origField := orig.Field(i)
		if origField.PkgPath != "" {
			continue
		}
		field := Field{
			name: origField.Name,
			tag:  origField.Tag,
			keep: true,
		}
		if err = t.filter(&field); err != nil {
			return nil, fmt.Errorf("%s: %w", origField.Name, err)
		}
		if !field.keep {
			continue
		}
		filteredFields = append(filteredFields, t.newField(&origField, &field))
	}
	filtered = reflect.StructOf(filteredFields)
	t.types[orig] = filtered
	return
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

// mapType maps the specified original type to a matching generated type.
// If orig cannot be mapped (e. g., if it is recursive), nil is returned
// instead.
func (t *T) mapType(orig reflect.Type) reflect.Type {
	switch orig.Kind() {
	case reflect.Array:
		elem := t.mapType(orig.Elem())
		if elem == nil {
			return nil
		}
		if elem == orig.Elem() {
			return orig
		}
		return reflect.ArrayOf(orig.Len(), elem)
	case reflect.Map:
		key := t.mapType(orig.Key())
		elem := t.mapType(orig.Elem())
		if key == nil || elem == nil {
			return nil
		}
		if key == orig.Key() && elem == orig.Elem() {
			return orig
		}
		return reflect.MapOf(key, elem)
	case reflect.Ptr:
		elem := t.mapType(orig.Elem())
		if elem == nil {
			return nil
		}
		if elem == orig.Elem() {
			return orig
		}
		return reflect.PtrTo(elem)
	case reflect.Slice:
		elem := t.mapType(orig.Elem())
		if elem == nil {
			return nil
		}
		if elem == orig.Elem() {
			return orig
		}
		return reflect.SliceOf(elem)
	case reflect.Struct:
		elem, ok := t.types[orig]
		if ok {
			return elem // nil if recursive
		}
		elem, err := t.filterType(orig)
		if err != nil {
			return nil
		}
		return elem
	default:
		return orig
	}
}

// New creates a new structure filter based on the specified filter functions.
// The filter functions are called in order for each structure field.
func New(filters ...Func) *T {
	return &T{
		filter: combineFilters(filters),
		types:  make(map[reflect.Type]reflect.Type),
	}
}

// combineFilters combines multiple filters (or none) into a single filter.
func combineFilters(filters []Func) Func {
	switch len(filters) {
	case 0:
		return func(*Field) error {
			return nil
		}
	case 1:
		return filters[0]
	default:
		return func(field *Field) error {
			for i, filter := range filters {
				if err := filter(field); err != nil {
					return fmt.Errorf("filter[%d]: %w", i, err)
				}
			}
			return nil
		}
	}
}
