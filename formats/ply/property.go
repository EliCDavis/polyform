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
	Short  ScalarPropertyType = "Short"
	Ushort ScalarPropertyType = "ushort"
	Int    ScalarPropertyType = "int"
	Uint   ScalarPropertyType = "uint"
	Float  ScalarPropertyType = "float"
	Double ScalarPropertyType = "double"
)

type ScalarProperty struct {
	name string
	Type ScalarPropertyType
}

func (sp ScalarProperty) Name() string {
	return sp.name
}

func (sp ScalarProperty) Write(out io.Writer) (err error) {
	_, err = fmt.Fprintf(out, "property %s %s\n", sp.Type, sp.name)
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
