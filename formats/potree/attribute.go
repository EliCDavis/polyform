package potree

import "strings"

// https://github.com/potree/PotreeConverter/blob/c2328c433c7776e231d86712bb4074c82659e366/Converter/include/Attributes.h#L80

type AttributeType string

const (
	Int8AttributeType      = "int8"
	Int16AttributeType     = "int16"
	Int32AttributeType     = "int32"
	Int64AttributeType     = "int64"
	UInt8AttributeType     = "uint8"
	UInt16AttributeType    = "uint16"
	UInt32AttributeType    = "uint32"
	UInt64AttributeType    = "uint64"
	FloatAttributeType     = "float"
	DoubleAttributeType    = "double"
	UndefinedAttributeType = "undefined"
)

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
