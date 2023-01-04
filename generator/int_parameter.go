package generator

import (
	"fmt"
	"math/rand"
)

func NewIntParameter(min, max int) IntParameter {
	return IntParameter{
		minInclusive: min,
		maxInclusive: max,
	}
}

type IntParameter struct {
	minInclusive int
	maxInclusive int
	set          bool
	setValue     int
}

func (ip IntParameter) Value() int {
	if ip.set {
		return ip.setValue
	}
	return ip.minInclusive + rand.Intn(ip.maxInclusive-ip.minInclusive+1)
}

func (ip IntParameter) IsSet() bool {
	return ip.set
}

func (ip *IntParameter) Set(value int) {
	if value < ip.minInclusive || value > ip.maxInclusive {
		panic(fmt.Errorf("invalid int parameter value %d is not in range [%d, %d]", value, ip.minInclusive, ip.maxInclusive))
	}
	ip.set = true
	ip.setValue = value
}
