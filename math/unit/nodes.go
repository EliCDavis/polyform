package unit

import (
	"math"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/EliCDavis/vector"
)

func init() {
	factory := &refutil.TypeFactory{}

	refutil.RegisterType[nodes.Struct[FeetToMetersNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[FeetToMetersNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[MeterToFeetNode[float64]]](factory)
	refutil.RegisterType[nodes.Struct[MeterToFeetNode[int]]](factory)

	refutil.RegisterType[nodes.Struct[ParseFeetNode]](factory)

	generator.RegisterTypes(factory)
}

type FeetToMetersNode[T vector.Number] struct {
	Feet nodes.Output[T]
}

func (ftm FeetToMetersNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(nodes.TryGetOutputValue(ftm.Feet, 0)) * FeetToMeters)
}

func (ftm FeetToMetersNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(math.Round(float64(nodes.TryGetOutputValue(ftm.Feet, 0)) * FeetToMeters)))
}

type MeterToFeetNode[T vector.Number] struct {
	Meters nodes.Output[T]
}

func (ftm MeterToFeetNode[T]) Float64() nodes.StructOutput[float64] {
	return nodes.NewStructOutput(float64(nodes.TryGetOutputValue(ftm.Meters, 0)) * MetersToFeet)
}

func (ftm MeterToFeetNode[T]) Int() nodes.StructOutput[int] {
	return nodes.NewStructOutput(int(math.Round(float64(nodes.TryGetOutputValue(ftm.Meters, 0)) * MetersToFeet)))
}

type ParseFeetNode struct {
	Feet nodes.Output[string]
}

func (ftm ParseFeetNode) Float64() nodes.StructOutput[float64] {
	feet, err := ParseFeet(nodes.TryGetOutputValue(ftm.Feet, ""))
	out := nodes.NewStructOutput(feet)
	if err != nil {
		out.LogError(err)
	}
	return out
}

func (ftm ParseFeetNode) Int() nodes.StructOutput[int] {
	feet, err := ParseFeet(nodes.TryGetOutputValue(ftm.Feet, ""))
	out := nodes.NewStructOutput(int(math.Round(feet)))
	if err != nil {
		out.LogError(err)
	}
	return out
}
