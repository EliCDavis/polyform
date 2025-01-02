package main

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type ShiftNode = nodes.Struct[[]float64, ShiftNodeData]

type ShiftNodeData struct {
	In    nodes.NodeOutput[[]float64]
	Shift nodes.NodeOutput[float64]
}

func (snd ShiftNodeData) Process() ([]float64, error) {
	if snd.In == nil {
		return nil, nil
	}

	if snd.Shift == nil {
		return snd.In.Value(), nil
	}

	in := snd.In.Value()
	shift := snd.Shift.Value()

	out := make([]float64, len(in))
	for i, v := range in {
		out[i] = v + shift
	}

	return out, nil
}

type VectorArrayNode = nodes.Struct[[]vector3.Float64, VectoryArrayNodeData]

type VectoryArrayNodeData struct {
	X nodes.NodeOutput[[]float64]
	Y nodes.NodeOutput[[]float64]
	Z nodes.NodeOutput[[]float64]
}

func (snd VectoryArrayNodeData) Process() ([]vector3.Float64, error) {
	var xArr []float64
	var yArr []float64
	var zArr []float64

	if snd.X != nil {
		xArr = snd.X.Value()
	}

	if snd.Y != nil {
		yArr = snd.Y.Value()
	}

	if snd.Z != nil {
		zArr = snd.Z.Value()
	}

	out := make([]vector3.Float64, max(len(xArr), len(yArr), len(zArr)))
	for i := 0; i < len(out); i++ {
		x := 0.
		y := 0.
		z := 0.

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		out[i] = vector3.New(x, y, z)
	}

	return out, nil
}

type QuaternionArrayFromThetaNode = nodes.Struct[[]quaternion.Quaternion, QuaternionArrayFromThetaNodeData]

type QuaternionArrayFromThetaNodeData struct {
	X nodes.NodeOutput[[]float64]
	Y nodes.NodeOutput[[]float64]
	Z nodes.NodeOutput[[]float64]
	W nodes.NodeOutput[[]float64]
}

func (snd QuaternionArrayFromThetaNodeData) Process() ([]quaternion.Quaternion, error) {
	var xArr []float64
	var yArr []float64
	var zArr []float64
	var wArr []float64

	if snd.X != nil {
		xArr = snd.X.Value()
	}

	if snd.Y != nil {
		yArr = snd.Y.Value()
	}

	if snd.Z != nil {
		zArr = snd.Z.Value()
	}

	if snd.W != nil {
		wArr = snd.W.Value()
	}

	out := make([]quaternion.Quaternion, max(len(xArr), len(yArr), len(zArr)))
	for i := 0; i < len(out); i++ {
		x := 0.
		y := 0.
		z := 0.
		w := 0.

		if i < len(xArr) {
			x = xArr[i]
		}

		if i < len(yArr) {
			y = yArr[i]
		}

		if i < len(zArr) {
			z = zArr[i]
		}

		if i < len(wArr) {
			w = wArr[i]
		}

		out[i] = quaternion.FromTheta(w, vector3.New(x, y, z))
	}

	return out, nil
}
