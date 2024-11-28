package potree

import (
	"fmt"
	"strings"
)

// https://github.com/potree/PotreeConverter/blob/c2328c433c7776e231d86712bb4074c82659e366/Converter/include/Attributes.h#L80

type AttributeType string

const (
	Int8AttributeType      AttributeType = "int8"
	Int16AttributeType     AttributeType = "int16"
	Int32AttributeType     AttributeType = "int32"
	Int64AttributeType     AttributeType = "int64"
	UInt8AttributeType     AttributeType = "uint8"
	UInt16AttributeType    AttributeType = "uint16"
	UInt32AttributeType    AttributeType = "uint32"
	UInt64AttributeType    AttributeType = "uint64"
	FloatAttributeType     AttributeType = "float"
	DoubleAttributeType    AttributeType = "double"
	UndefinedAttributeType AttributeType = "undefined"
)

func (at AttributeType) Size() int {
	switch at {
	case Int8AttributeType, UInt8AttributeType:
		return 1

	case Int16AttributeType, UInt16AttributeType:
		return 2

	case Int32AttributeType, UInt32AttributeType, FloatAttributeType:
		return 4

	case Int64AttributeType, UInt64AttributeType, DoubleAttributeType:
		return 8

	default:
		panic(fmt.Errorf("unimplemented byte size for attribute type: %s", at))
	}
}

type Attribute struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Size        int           `json:"size"`
	NumElements int           `json:"numElements"`
	ElementSize int           `json:"elementSize"`
	Type        AttributeType `json:"type"`
	Min         []float64     `json:"min"`
	Max         []float64     `json:"max"`
	Scale       []float64     `json:"scale,omitempty"`
	Offset      []float64     `json:"offset,omitempty"`
}

func (a Attribute) IsPosition() bool {
	return a.Name == "position" || a.Name == "POSITION_CARTESIAN"
}

func (a Attribute) IsColor() bool {
	n := strings.ToLower(a.Name)
	return n == "rgba" || n == "rgb"
}
