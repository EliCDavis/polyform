package ply

import (
	"fmt"
	"io"
)

type Property interface {
	Name() string              // Name of the property as found in the PLY header
	Write(out io.Writer) error // Writes out the definition of the property in PLY format
}

type ScalarPropertyType string

const (
	Char   ScalarPropertyType = "char"
	UChar  ScalarPropertyType = "uchar" //  uint8
	Short  ScalarPropertyType = "short"
	UShort ScalarPropertyType = "ushort"
	Int    ScalarPropertyType = "int"
	UInt   ScalarPropertyType = "uint"
	Float  ScalarPropertyType = "float"
	Double ScalarPropertyType = "double"
)

type ScalarProperty struct {
	PropertyName string             `json:"name"` // Name of the property
	Type         ScalarPropertyType `json:"type"` // Property type
}

// Name of the property as found in the PLY header
func (sp ScalarProperty) Name() string {
	return sp.PropertyName
}

// Size of the property on a per point basis when serialized to binary format
func (sp ScalarProperty) Size() int {
	switch sp.Type {
	case Char, UChar:
		return 1

	case Short, UShort:
		return 2

	case Int, UInt, Float:
		return 4

	case Double:
		return 8

	default:
		panic(fmt.Errorf("unimplemented byte size for scalar property type: %s", sp.Type))
	}
}

// Writes out the definition of the property in PLY format
func (sp ScalarProperty) Write(out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "property %s %s\n", sp.Type, sp.PropertyName)
	return
}

type ListProperty struct {
	PropertyName string             // Name of the property
	CountType    ScalarPropertyType // Data type of the number used to define how many elements are in the list
	ListType     ScalarPropertyType // Data type of the elements in the list
}

// Name of the property as found in the PLY header
func (lp ListProperty) Name() string {
	return lp.PropertyName
}

// Writes out the definition of the property in PLY format
func (lp ListProperty) Write(out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "property list %s %s %s\n", lp.CountType, lp.ListType, lp.PropertyName)
	return
}
