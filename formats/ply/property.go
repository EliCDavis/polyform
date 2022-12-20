package ply

type Property interface {
	Name() string
}

type ScalarPropertyType string

const (
	Char   ScalarPropertyType = "char"
	UChar  ScalarPropertyType = "uchar"
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

type ListProperty struct {
	name      string
	countType ScalarPropertyType
	listType  ScalarPropertyType
}

func (lp ListProperty) Name() string {
	return lp.name
}
