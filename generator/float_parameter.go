package generator

import (
	"fmt"
	"math/rand"
)

func NewFloatParameter(min, max float64) FloatParameter {
	return FloatParameter{
		minInclusive: min,
		maxInclusive: max,
		set:          false,
	}
}

type FloatParameter struct {
	minInclusive float64
	maxInclusive float64
	set          bool
	setValue     float64
}

func (fp FloatParameter) Value() float64 {
	if fp.set {
		return fp.setValue
	}
	return fp.minInclusive + (rand.Float64() * (fp.maxInclusive - fp.minInclusive))
}

func (fp FloatParameter) IsSet() bool {
	return fp.set
}

func (fp *FloatParameter) Set(value float64) {
	if value < fp.minInclusive || value > fp.maxInclusive {
		panic(fmt.Errorf("invalid float parameter value %f is not in range [%f, %f]", value, fp.minInclusive, fp.maxInclusive))
	}
	fp.set = true
	fp.setValue = value
}
