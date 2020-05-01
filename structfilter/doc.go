// Package structfilter converts a struct type to a new struct type with some
// fields removed or their tags altered. This package provides facilities to
// convert values from the old struct type to the new one.
//
// A typical use case is to remove sensitive or superfluous fields from or
// adjust tags in a structure type before marshalling the structure, or before
// handing it over to a logging framework.
//
// The generated types have neither methods nor unexported fields.
package structfilter
