package repeat_test

import (
	"testing"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeshSurfaceSamplesLandOnTheSurface(t *testing.T) {
	cube := primitives.UnitCube()
	surface := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 500, Seed: 3}

	transforms := surface.TRS()
	require.Len(t, transforms, 500)

	onFace := func(p vector3.Float64) bool {
		for _, c := range []float64{p.X(), p.Y(), p.Z()} {
			if abs(abs(c)-0.5) < 1e-9 {
				return true
			}
		}
		return false
	}
	faces := map[int]int{}
	for _, tr := range transforms {
		p := tr.Position()
		assert.True(t, onFace(p), "%v is not on the cube", p)
		assert.LessOrEqual(t, p.MaxComponent(), 0.5+1e-9)
		assert.GreaterOrEqual(t, p.MinComponent(), -0.5-1e-9)
		faces[faceOf(p)]++
	}
	assert.Len(t, faces, 6, "every face gets samples")
	for face, count := range faces {
		assert.InDelta(t, 500/6, count, 35, "face %d", face)
	}
}

func TestMeshSurfaceIsRepeatableBySeed(t *testing.T) {
	cube := primitives.UnitCube()
	a := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 20, Seed: 1}.TRS()
	b := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 20, Seed: 1}.TRS()
	c := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 20, Seed: 2}.TRS()
	assert.Equal(t, a, b)
	assert.NotEqual(t, a, c)
}

func TestMeshSurfaceKeepsEarlySamplesWhenCountGrows(t *testing.T) {
	cube := primitives.UnitCube()
	few := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 5, Seed: 1}.TRS()
	many := repeat.MeshSurface{Mesh: cube, Attribute: modeling.PositionAttribute, Samples: 50, Seed: 1}.TRS()
	assert.Equal(t, few, many[:5])
}

func abs(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}

func faceOf(p vector3.Float64) int {
	switch {
	case abs(p.X()-0.5) < 1e-9:
		return 0
	case abs(p.X()+0.5) < 1e-9:
		return 1
	case abs(p.Y()-0.5) < 1e-9:
		return 2
	case abs(p.Y()+0.5) < 1e-9:
		return 3
	case abs(p.Z()-0.5) < 1e-9:
		return 4
	default:
		return 5
	}
}
