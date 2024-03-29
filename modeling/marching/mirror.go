package marching

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Axis int

const (
	XAxis Axis = iota
	YAxis
	ZAxis
)

func MirrorAxis(field Field, axisToMirror Axis) Field {
	float1Functions := make(map[string]sample.Vec3ToFloat)
	float2Functions := make(map[string]sample.Vec3ToVec2)
	float3Functions := make(map[string]sample.Vec3ToVec3)

	var newV sample.Vec3ToVec3

	topRightFrontCorner := field.Domain.
		Center().
		Add(field.Domain.Size().DivByConstant(2))
	bottomLeftBackCorner := field.Domain.
		Center().
		Sub(field.Domain.Size().DivByConstant(2))

	if axisToMirror == XAxis {
		newV = func(v vector3.Float64) vector3.Float64 {
			return vector3.New(math.Abs(v.X()), v.Y(), v.Z())
		}
		bottomLeftBackCorner = bottomLeftBackCorner.SetX(-math.Abs(topRightFrontCorner.X()))
	} else if axisToMirror == YAxis {
		newV = func(v vector3.Float64) vector3.Float64 {
			return vector3.New(v.X(), math.Abs(v.Y()), v.Z())
		}
		bottomLeftBackCorner = bottomLeftBackCorner.SetY(-math.Abs(topRightFrontCorner.Y()))
	} else if axisToMirror == ZAxis {
		newV = func(v vector3.Float64) vector3.Float64 {
			return vector3.New(v.X(), v.Y(), math.Abs(v.Z()))
		}
		bottomLeftBackCorner = bottomLeftBackCorner.SetZ(-math.Abs(topRightFrontCorner.Z()))
	} else {
		panic(fmt.Errorf("unimplemented Axis: %d", axisToMirror))
	}

	for atr, f := range field.Float1Functions {
		float1Functions[atr] = func(v vector3.Float64) float64 {
			return f(newV(v))
		}
	}

	for atr, f := range field.Float2Functions {
		float2Functions[atr] = func(v vector3.Float64) vector2.Float64 {
			return f(newV(v))
		}
	}

	for atr, f := range field.Float3Functions {
		float3Functions[atr] = func(v vector3.Float64) vector3.Float64 {
			return f(newV(v))
		}
	}

	newDomain := geometry.NewAABB(field.Domain.Center(), field.Domain.Size())
	newDomain.EncapsulatePoint(bottomLeftBackCorner)
	return Field{
		Domain:          newDomain,
		Float1Functions: float1Functions,
		Float2Functions: float2Functions,
		Float3Functions: float3Functions,
	}
}
