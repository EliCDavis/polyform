package marching

import (
	"fmt"
	"image/color"
	"math"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector"
)

func useValueOfSmallestFunc(indicators []sample.Vec3ToFloat, valuesToReturn []sample.Vec3ToVec3) sample.Vec3ToVec3 {

	if len(indicators) == 0 {
		panic("no functions to use")
	}

	if len(indicators) != len(valuesToReturn) {
		panic(fmt.Errorf("indicator count must match values count: %d != %d", len(indicators), len(valuesToReturn)))
	}

	if len(indicators) == 1 {
		return valuesToReturn[0]
	}

	return func(v vector.Vector3) vector.Vector3 {
		minIndex := 0
		minValue := indicators[0](v)
		for i := 1; i < len(indicators); i++ {
			val := indicators[i](v)
			if val < minValue {
				minValue = val
				minIndex = i
			}
		}
		return valuesToReturn[minIndex](v)
	}

}

func CombineFields(fields ...Field) Field {
	if len(fields) == 0 {
		panic("no fields to combine")
	}

	if len(fields) == 1 {
		return fields[0]
	}

	float1Aggregate := make(map[string][]sample.Vec3ToFloat)
	float2Aggregate := make(map[string][]sample.Vec3ToVec2)
	float3Aggregate := make(map[string][]sample.Vec3ToVec3)

	box := modeling.NewAABB(fields[0].Domain.Center(), fields[0].Domain.Size())
	for _, otherF := range fields {
		box.EncapsulateBounds(otherF.Domain)

		for attribute, function := range otherF.Float1Functions {
			if _, ok := float1Aggregate[attribute]; !ok {
				float1Aggregate[attribute] = make([]sample.Vec3ToFloat, 0)
			}
			float1Aggregate[attribute] = append(float1Aggregate[attribute], function)
		}

		for attribute, function := range otherF.Float2Functions {
			if _, ok := float2Aggregate[attribute]; !ok {
				float2Aggregate[attribute] = make([]sample.Vec3ToVec2, 0)
			}
			float2Aggregate[attribute] = append(float2Aggregate[attribute], function)
		}

		for attribute, function := range otherF.Float3Functions {
			if _, ok := float3Aggregate[attribute]; !ok {
				float3Aggregate[attribute] = make([]sample.Vec3ToVec3, 0)
			}
			float3Aggregate[attribute] = append(float3Aggregate[attribute], function)
		}
	}

	float1Final := make(map[string]sample.Vec3ToFloat)
	for attribute, functions := range float1Aggregate {
		float1Final[attribute] = sdf.Union(functions...)
	}

	float2Final := make(map[string]sample.Vec3ToVec2)
	for attribute, functions := range float2Aggregate {
		float2Final[attribute] = sample.AverageVec3ToVec2(functions...)
	}

	float3Final := make(map[string]sample.Vec3ToVec3)
	for attribute, functions := range float3Aggregate {
		float3Final[attribute] = useValueOfSmallestFunc(float1Aggregate[modeling.PositionAttribute], functions)
	}

	return Field{
		Float1Functions: float1Final,
		Float2Functions: float2Final,
		Float3Functions: float3Final,
		Domain:          box,
	}
}

type Field struct {
	Domain          modeling.AABB
	Float1Functions map[string]sample.Vec3ToFloat
	Float2Functions map[string]sample.Vec3ToVec2
	Float3Functions map[string]sample.Vec3ToVec3
}

func translateV2(field sample.Vec3ToVec2, translation vector.Vector3) sample.Vec3ToVec2 {
	return func(v vector.Vector3) vector.Vector2 {
		return field(v.Sub(translation))
	}
}

func translateV3(field sample.Vec3ToVec3, translation vector.Vector3) sample.Vec3ToVec3 {
	return func(v vector.Vector3) vector.Vector3 {
		return field(v.Sub(translation))
	}
}

func (f Field) Translate(translation vector.Vector3) Field {
	float1Final := make(map[string]sample.Vec3ToFloat)
	for attribute, functions := range f.Float1Functions {
		float1Final[attribute] = sdf.Translate(functions, translation)
	}

	float2Final := make(map[string]sample.Vec3ToVec2)
	for attribute, functions := range f.Float2Functions {
		float2Final[attribute] = translateV2(functions, translation)
	}

	float3Final := make(map[string]sample.Vec3ToVec3)
	for attribute, functions := range f.Float3Functions {
		float3Final[attribute] = translateV3(functions, translation)
	}

	return Field{
		Float1Functions: float1Final,
		Float2Functions: float2Final,
		Float3Functions: float3Final,
		Domain:          modeling.NewAABB(f.Domain.Center().Add(translation), f.Domain.Size()),
	}
}

func (f Field) Combine(otherFields ...Field) Field {
	if len(otherFields) == 0 {
		return f
	}
	return CombineFields(append(otherFields, f)...)
}

func (f Field) Modify(attribute string, other Field, modifier func(a, b sample.Vec3ToFloat) sample.Vec3ToFloat) Field {
	newDomain := modeling.NewAABB(f.Domain.Center(), f.Domain.Size())
	newDomain.EncapsulateBounds(other.Domain)
	return Field{
		Domain: newDomain,
		Float1Functions: map[string]sample.Vec3ToFloat{
			attribute: modifier(f.Float1Functions[attribute], other.Float1Functions[attribute]),
		},
	}
}

func (f Field) SetFloat3Attribute(atr string, f3tf3 sample.Vec3ToVec3) Field {
	float1Final := make(map[string]sample.Vec3ToFloat)
	for attribute, functions := range f.Float1Functions {
		float1Final[attribute] = functions
	}

	float2Final := make(map[string]sample.Vec3ToVec2)
	for attribute, functions := range f.Float2Functions {
		float2Final[attribute] = functions
	}

	float3Final := make(map[string]sample.Vec3ToVec3)
	for attribute, functions := range f.Float3Functions {
		float3Final[attribute] = functions
	}

	float3Final[atr] = f3tf3

	return Field{
		Float1Functions: float1Final,
		Float2Functions: float2Final,
		Float3Functions: float3Final,
		Domain:          modeling.NewAABB(f.Domain.Center(), f.Domain.Size()),
	}
}

func (f Field) WithColor(c color.RGBA) Field {
	colorAsVector := vector.NewVector3(
		float64(c.R)/255.,
		float64(c.G)/255.,
		float64(c.B)/255.,
	)

	return f.SetFloat3Attribute(
		modeling.ColorAttribute,
		func(v vector.Vector3) vector.Vector3 {
			return colorAsVector
		},
	)
}

func (f Field) March(atr string, cubesPerUnit, cutoff float64) modeling.Mesh {

	v1Data := make(map[string][]float64)
	v2Data := make(map[string][]vector.Vector2)
	v3Data := make(map[string][]vector.Vector3)

	var atrFunc sample.Vec3ToFloat
	for atrs, f1f := range f.Float1Functions {
		v1Data[atrs] = make([]float64, 0)
		if atrs == atr {
			atrFunc = f1f
		}
	}

	for atr := range f.Float2Functions {
		v2Data[atr] = make([]vector.Vector2, 0)
	}

	for atr := range f.Float3Functions {
		v3Data[atr] = make([]vector.Vector3, 0)
	}

	if atrFunc == nil {
		panic(fmt.Errorf("Field doesn't contain f1 function for attribute %s", atr))
	}

	min := f.Domain.Min()
	max := f.Domain.Max()

	minCanvas := modeling.VectorInt{
		X: int(math.Floor(min.X()*cubesPerUnit)) - 1,
		Y: int(math.Floor(min.Y()*cubesPerUnit)) - 1,
		Z: int(math.Floor(min.Z()*cubesPerUnit)) - 1,
	}

	maxCanvas := modeling.VectorInt{
		X: int(math.Ceil(max.X()*cubesPerUnit)) + 1,
		Y: int(math.Ceil(max.Y()*cubesPerUnit)) + 1,
		Z: int(math.Ceil(max.Z()*cubesPerUnit)) + 1,
	}

	cubesToUnit := 1. / cubesPerUnit

	tris := make([]int, 0)

	for x := minCanvas.X; x < maxCanvas.X-1; x++ {
		for y := minCanvas.Y; y < maxCanvas.Y-1; y++ {
			for z := minCanvas.Z; z < maxCanvas.Z-1; z++ {
				v := vector.NewVector3(float64(x), float64(y), float64(z)).MultByConstant(cubesToUnit)

				cubeCornerPositions := []vector.Vector3{
					v,
					v.Add(vector.NewVector3(cubesToUnit, 0, 0)),
					v.Add(vector.NewVector3(cubesToUnit, 0, cubesToUnit)),
					v.Add(vector.NewVector3(0, 0, cubesToUnit)),
					v.Add(vector.NewVector3(0, cubesToUnit, 0)),
					v.Add(vector.NewVector3(cubesToUnit, cubesToUnit, 0)),
					v.Add(vector.NewVector3(cubesToUnit, cubesToUnit, cubesToUnit)),
					v.Add(vector.NewVector3(0, cubesToUnit, cubesToUnit)),
				}

				cubeCorners := []float64{
					atrFunc(cubeCornerPositions[0]),
					atrFunc(cubeCornerPositions[1]),
					atrFunc(cubeCornerPositions[2]),
					atrFunc(cubeCornerPositions[3]),
					atrFunc(cubeCornerPositions[4]),
					atrFunc(cubeCornerPositions[5]),
					atrFunc(cubeCornerPositions[6]),
					atrFunc(cubeCornerPositions[7]),
				}

				cubeCornersExistence := []bool{
					cubeCorners[0] < cutoff,
					cubeCorners[1] < cutoff,
					cubeCorners[2] < cutoff,
					cubeCorners[3] < cutoff,
					cubeCorners[4] < cutoff,
					cubeCorners[5] < cutoff,
					cubeCorners[6] < cutoff,
					cubeCorners[7] < cutoff,
				}

				var lookupIndex = 0
				if cubeCornersExistence[0] {
					lookupIndex |= 1
				}
				if cubeCornersExistence[1] {
					lookupIndex |= 2
				}
				if cubeCornersExistence[2] {
					lookupIndex |= 4
				}
				if cubeCornersExistence[3] {
					lookupIndex |= 8
				}
				if cubeCornersExistence[4] {
					lookupIndex |= 16
				}
				if cubeCornersExistence[5] {
					lookupIndex |= 32
				}
				if cubeCornersExistence[6] {
					lookupIndex |= 64
				}
				if cubeCornersExistence[7] {
					lookupIndex |= 128
				}

				for i := 0; triangulation[lookupIndex][i] != -1; i += 3 {
					// Get indices of corner points A and B for each of the three edges
					// of the cube that need to be joined to form the triangle.
					a0 := cornerIndexAFromEdge[triangulation[lookupIndex][i]]
					b0 := cornerIndexBFromEdge[triangulation[lookupIndex][i]]

					a1 := cornerIndexAFromEdge[triangulation[lookupIndex][i+1]]
					b1 := cornerIndexBFromEdge[triangulation[lookupIndex][i+1]]

					a2 := cornerIndexAFromEdge[triangulation[lookupIndex][i+2]]
					b2 := cornerIndexBFromEdge[triangulation[lookupIndex][i+2]]

					t1 := interpolationValueFromCutoff(cubeCorners[a0], cubeCorners[b0], cutoff)
					t2 := interpolationValueFromCutoff(cubeCorners[a1], cubeCorners[b1], cutoff)
					t3 := interpolationValueFromCutoff(cubeCorners[a2], cubeCorners[b2], cutoff)

					v1 := interpolateV3(cubeCornerPositions[a0], cubeCornerPositions[b0], t1)
					v2 := interpolateV3(cubeCornerPositions[a1], cubeCornerPositions[b1], t2)
					v3 := interpolateV3(cubeCornerPositions[a2], cubeCornerPositions[b2], t3)

					v3Data[atr] = append(v3Data[atr], v1, v2, v3)

					for atr, f := range f.Float1Functions {
						v1Data[atr] = append(
							v1Data[atr],
							interpolateV1(f(cubeCornerPositions[a0]), f(cubeCornerPositions[b0]), t1),
							interpolateV1(f(cubeCornerPositions[a1]), f(cubeCornerPositions[b1]), t2),
							interpolateV1(f(cubeCornerPositions[a2]), f(cubeCornerPositions[b2]), t3),
						)
					}

					for atr, f := range f.Float2Functions {
						v2Data[atr] = append(
							v2Data[atr],
							interpolateV2(f(cubeCornerPositions[a0]), f(cubeCornerPositions[b0]), t1),
							interpolateV2(f(cubeCornerPositions[a1]), f(cubeCornerPositions[b1]), t2),
							interpolateV2(f(cubeCornerPositions[a2]), f(cubeCornerPositions[b2]), t3),
						)
					}

					for atr, f := range f.Float3Functions {
						v3Data[atr] = append(
							v3Data[atr],
							interpolateV3(f(cubeCornerPositions[a0]), f(cubeCornerPositions[b0]), t1),
							interpolateV3(f(cubeCornerPositions[a1]), f(cubeCornerPositions[b1]), t2),
							interpolateV3(f(cubeCornerPositions[a2]), f(cubeCornerPositions[b2]), t3),
						)
					}

					startIndex := len(tris)
					tris = append(
						tris,
						startIndex,
						startIndex+1,
						startIndex+2,
					)
				}
			}
		}
	}
	return modeling.NewMesh(tris, v3Data, v2Data, v1Data, nil).WeldByFloat3Attribute(modeling.PositionAttribute, 3)
}
