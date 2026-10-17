package marching

import (
	"math"
	"sort"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/noise"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Reference: halves every axis together until the longest side is under a voxel.
func floatOctree(field sample.Vec3ToFloat, bounds geometry.AABB, cubeSize float64, res map[vector3.Int32]float64) {
	center := bounds.Center()
	index := center.DivByConstant(cubeSize).RoundToInt()
	at := index.ToFloat64().Scale(cubeSize)
	value := field(at)
	size := bounds.Size()
	if math.Abs(value) > size.Length()*2+cubeSize+center.Distance(at) {
		return
	}
	res[vector3.New(int32(index.X()), int32(index.Y()), int32(index.Z()))] = value
	if size.MaxComponent() < cubeSize {
		return
	}
	half := size.Scale(0.5)
	for _, sx := range []float64{1, -1} {
		for _, sy := range []float64{1, -1} {
			for _, sz := range []float64{1, -1} {
				offset := vector3.New(half.X()*sx, half.Y()*sy, half.Z()*sz).Scale(0.5)
				floatOctree(field, geometry.NewAABB(center.Add(offset), half), cubeSize, res)
			}
		}
	}
}

func triangleSet(m modeling.Mesh, cubeSize float64) map[[9]int64]bool {
	pos := m.Float3Attribute(modeling.PositionAttribute)
	idx := m.Indices()
	snap := func(v vector3.Float64) [3]int64 {
		s := v.DivByConstant(cubeSize).Scale(1000)
		return [3]int64{int64(math.Round(s.X())), int64(math.Round(s.Y())), int64(math.Round(s.Z()))}
	}
	set := map[[9]int64]bool{}
	for i := 0; i+2 < idx.Len(); i += 3 {
		corners := [][3]int64{snap(pos.At(idx.At(i))), snap(pos.At(idx.At(i + 1))), snap(pos.At(idx.At(i + 2)))}
		sort.Slice(corners, func(a, b int) bool {
			for k := range 3 {
				if corners[a][k] != corners[b][k] {
					return corners[a][k] < corners[b][k]
				}
			}
			return false
		})
		var key [9]int64
		for c := range 3 {
			copy(key[c*3:], corners[c][:])
		}
		set[key] = true
	}
	return set
}

func openEdges(m modeling.Mesh) int {
	uses := map[[2]int]int{}
	idx := m.Indices()
	for i := 0; i+2 < idx.Len(); i += 3 {
		tri := [3]int{idx.At(i), idx.At(i + 1), idx.At(i + 2)}
		for k := range 3 {
			a, b := tri[k], tri[(k+1)%3]
			uses[[2]int{min(a, b), max(a, b)}]++
		}
	}
	open := 0
	for _, n := range uses {
		if n != 2 {
			open++
		}
	}
	return open
}

func displaced(base sample.Vec3ToFloat, amplitude, frequency float64) sample.Vec3ToFloat {
	return func(p vector3.Float64) float64 { return base(p) + amplitude*noise.Perlin3D(p.Scale(frequency)) }
}

func TestMarchMatchesTheFloatOctreeWhereNoSurfaceCrossesTheDomain(t *testing.T) {
	o := vector3.Zero[float64]()
	cases := map[string]struct {
		field  sample.Vec3ToFloat
		domain geometry.AABB
	}{
		"sphere":       {sdf.Sphere(vector3.New(0.13, -0.07, 0.31), 1), geometry.NewAABB(vector3.New(0.1, 0., 0.3), vector3.Fill(2.4))},
		"torus":        {sdf.Torus(o, 1, 0.3), geometry.NewAABB(o, vector3.New(2.8, 0.8, 2.8))},
		"long capsule": {sdf.Line(vector3.New(-3., 0., 0.), vector3.New(3., 0., 0.), 0.3), geometry.NewAABB(o, vector3.New(6.8, 0.8, 0.8))},
		"thin plate":   {sdf.Box(o, vector3.New(2., 0.1, 2.)), geometry.NewAABB(o, vector3.New(2.4, 0.5, 2.4))},
		"tilted box":   {sdf.Rotate(sdf.Box(o, vector3.New(1., 0.1, 1.)), quaternion.FromTheta(0.4, vector3.New(1., 1., 0.).Normalized())), geometry.NewAABB(o, vector3.Fill(2.))},
		"smooth union": {sdf.SmoothUnion(0.3, sdf.Sphere(o, 0.8), sdf.Sphere(vector3.New(0.9, 0.3, 0.), 0.6)), geometry.NewAABB(o, vector3.Fill(3.))},
		"subtract":     {sdf.Subtract(sdf.Box(o, vector3.Fill(1.6)), sdf.Sphere(vector3.Fill(0.5), 0.8)), geometry.NewAABB(o, vector3.Fill(2.2))},
		"gentle noise": {displaced(sdf.Sphere(o, 1), 0.05, 3), geometry.NewAABB(o, vector3.Fill(2.6))},
	}
	for name, c := range cases {
		for _, resolution := range []float64{16, 34} {
			cubeSize := 1 / resolution
			before := map[vector3.Int32]float64{}
			floatOctree(c.field, c.domain, cubeSize, before)

			want := triangleSet(meshFromSamples(before, cubeSize), cubeSize)
			got := triangleSet(March(c.field, c.domain, cubeSize, 0), cubeSize)
			require.NotEmpty(t, want, "%s at %g", name, resolution)
			assert.Equal(t, want, got, "%s at %g", name, resolution)
		}
	}
}

func TestMarchSamplesALongThinDomainAboutOncePerVoxel(t *testing.T) {
	boom := sdf.Line(vector3.New(0., 0., -3.), vector3.New(0., 0., 3.), 0.15)
	domain := geometry.NewAABB(vector3.Zero[float64](), vector3.New(0.5, 0.5, 6.5))
	evaluations := 0
	counted := func(p vector3.Float64) float64 { evaluations++; return boom(p) }

	samples := map[vector3.Int32]float64{}
	marchRecurse(counted, domain, 1./34, 0, samples)

	assert.LessOrEqual(t, float64(evaluations), 1.1*float64(len(samples)))
}

func TestMarchOpensNoMoreHolesThanTheFloatOctreeOnStrongNoise(t *testing.T) {
	field := displaced(sdf.Sphere(vector3.Zero[float64](), 1), 0.2, 5)
	domain := geometry.NewAABB(vector3.Zero[float64](), vector3.Fill(3.))
	cubeSize := 1. / 34

	before := map[vector3.Int32]float64{}
	floatOctree(field, domain, cubeSize, before)

	assert.LessOrEqual(t, openEdges(March(field, domain, cubeSize, 0)), openEdges(meshFromSamples(before, cubeSize)))
}
