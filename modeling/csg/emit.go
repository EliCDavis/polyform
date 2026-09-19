package csg

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

// Kept faces from one solid, with the mesh their vertex data lives in.
type kept struct {
	source modeling.Mesh
	faces  []face

	// Normals are negated where the solid's inside became the outside.
	inverted bool
}

type blendable[V any] interface {
	Scale(float64) V
	Add(V) V
}

// The corner data of a face's parent triangle, blended by the weights of one
// of its corners.
func blend[V blendable[V]](data *iter.ArrayIterator[V], corners [3]int, weights vector3.Float64) V {
	return data.At(corners[0]).Scale(weights.X()).
		Add(data.At(corners[1]).Scale(weights.Y())).
		Add(data.At(corners[2]).Scale(weights.Z()))
}

// Each corner blends its source triangle's corner data by weight. Attributes a
// source lacks are zero, except Normal, which falls back to the face normal.
func meshFromFaces(halves ...kept) (modeling.Mesh, []face) {
	count := 0
	for _, h := range halves {
		count += len(h.faces) * 3
	}

	tris := make([]int, count)
	for i := range tris {
		tris[i] = i
	}
	own := make([]face, 0, count/3)

	v4 := make(map[string][]vector4.Float64)
	v3 := map[string][]vector3.Float64{
		modeling.PositionAttribute: make([]vector3.Float64, 0, count),
		modeling.NormalAttribute:   make([]vector3.Float64, 0, count),
	}
	v2 := make(map[string][]vector2.Float64)
	v1 := make(map[string][]float64)
	for _, h := range halves {
		for _, name := range h.source.Float4Attributes() {
			v4[name] = make([]vector4.Float64, 0, count)
		}
		for _, name := range h.source.Float3Attributes() {
			v3[name] = make([]vector3.Float64, 0, count)
		}
		for _, name := range h.source.Float2Attributes() {
			v2[name] = make([]vector2.Float64, 0, count)
		}
		for _, name := range h.source.Float1Attributes() {
			v1[name] = make([]float64, 0, count)
		}
	}

	for _, h := range halves {
		indices := h.source.Indices()

		for _, f := range h.faces {
			corners := [3]int{
				indices.At(f.parent * 3),
				indices.At(f.parent*3 + 1),
				indices.At(f.parent*3 + 2),
			}
			own = append(own, face{verts: f.verts, normal: f.normal, parent: len(own), weights: ownCorners})

			for k := 0; k < 3; k++ {
				v3[modeling.PositionAttribute] = append(v3[modeling.PositionAttribute], f.verts[k])
				v3[modeling.NormalAttribute] = append(v3[modeling.NormalAttribute],
					normalAt(h, f, corners, k))

				for name := range v4 {
					var value vector4.Float64
					if h.source.HasFloat4Attribute(name) {
						value = blend(h.source.Float4Attribute(name), corners, f.weights[k])
					}
					v4[name] = append(v4[name], value)
				}
				for name := range v3 {
					if name == modeling.PositionAttribute || name == modeling.NormalAttribute {
						continue
					}
					var value vector3.Float64
					if h.source.HasFloat3Attribute(name) {
						value = blend(h.source.Float3Attribute(name), corners, f.weights[k])
					}
					v3[name] = append(v3[name], value)
				}
				for name := range v2 {
					var value vector2.Float64
					if h.source.HasFloat2Attribute(name) {
						value = blend(h.source.Float2Attribute(name), corners, f.weights[k])
					}
					v2[name] = append(v2[name], value)
				}
				for name := range v1 {
					value := 0.
					if h.source.HasFloat1Attribute(name) {
						data := h.source.Float1Attribute(name)
						value = f.weights[k].Dot(vector3.New(
							data.At(corners[0]),
							data.At(corners[1]),
							data.At(corners[2]),
						))
					}
					v1[name] = append(v1[name], value)
				}
			}
		}
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat4Data(v4).
		SetFloat3Data(v3).
		SetFloat2Data(v2).
		SetFloat1Data(v1), own
}

func normalAt(h kept, f face, corners [3]int, k int) vector3.Float64 {
	if !h.source.HasFloat3Attribute(modeling.NormalAttribute) {
		return f.normal
	}
	normal := blend(h.source.Float3Attribute(modeling.NormalAttribute), corners, f.weights[k])
	if normal.Length() == 0 {
		return f.normal
	}
	normal = normal.Normalized()
	if h.inverted {
		return normal.Scale(-1)
	}
	return normal
}
