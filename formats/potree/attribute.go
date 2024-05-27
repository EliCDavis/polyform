package potree

import "strings"

type Attribute struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Size        int       `json:"size"`
	NumElements int       `json:"numElements"`
	ElementSize int       `json:"elementSize"`
	Type        string    `json:"type"`
	Min         []float64 `json:"min"`
	Max         []float64 `json:"max"`
}

func (a Attribute) IsPosition() bool {
	return a.Name == "position" || a.Name == "POSITION_CARTESIAN"
}

func (a Attribute) IsColor() bool {
	n := strings.ToLower(a.Name)
	return n == "rgba" || n == "rgb"
}
