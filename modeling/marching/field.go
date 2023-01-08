package marching

import (
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
)

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
		float3Final[attribute] = sample.AverageVec3ToVec3(functions...)
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
