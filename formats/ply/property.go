package ply

import (
	"fmt"
	"io"
)

type Property interface {
	Name() string
	Write(out io.Writer) error
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
	PropertyName string             `json:"name"`
	Type         ScalarPropertyType `json:"type"`
}

func (sp ScalarProperty) Name() string {
	return sp.PropertyName
}

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

func (sp ScalarProperty) Write(out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "property %s %s\n", sp.Type, sp.PropertyName)
	return
}

type ListProperty struct {
	name      string
	countType ScalarPropertyType
	listType  ScalarPropertyType
}

func (lp ListProperty) Name() string {
	return lp.name
}

func (lp ListProperty) Write(out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "property list %s %s %s\n", lp.countType, lp.listType, lp.name)
	return
}
