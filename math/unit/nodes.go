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
	Feet nodes.LiftedPort[T]
}

func (ftm FeetToMetersNode[T]) Description() string {
	return "Converts feet to meters."
}

func (ftm FeetToMetersNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, ftm.Feet, func(v T) float64 { return float64(v) * FeetToMeters })
}

func (ftm FeetToMetersNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, ftm.Feet, func(v T) int {
		return int(math.Round(float64(v) * FeetToMeters))
	})
}

type MeterToFeetNode[T vector.Number] struct {
	Meters nodes.LiftedPort[T]
}

func (ftm MeterToFeetNode[T]) Description() string {
	return "Converts meters to feet."
}

func (ftm MeterToFeetNode[T]) Float64(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, ftm.Meters, func(v T) float64 { return float64(v) * MetersToFeet })
}

func (ftm MeterToFeetNode[T]) Int(out *nodes.Lifted[int]) {
	nodes.Zip1(out, ftm.Meters, func(v T) int {
		return int(math.Round(float64(v) * MetersToFeet))
	})
}

type ParseFeetNode struct {
	Feet nodes.Output[string]
}

func (ftm ParseFeetNode) Description() string {
	return "Parses a feet and inches string like 5'11\" into a number."
}

func (ftm ParseFeetNode) Float64(out *nodes.StructOutput[float64]) {
	feet, err := ParseFeet(nodes.TryGetOutputValue(out, ftm.Feet, ""))
	out.Set(feet)
	if err != nil {
		out.CaptureError(err)
	}
}

func (ftm ParseFeetNode) Int(out *nodes.StructOutput[int]) {
	feet, err := ParseFeet(nodes.TryGetOutputValue(out, ftm.Feet, ""))
	out.Set(int(math.Round(feet)))
	if err != nil {
		out.CaptureError(err)
	}
}
