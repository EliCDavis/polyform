package trees

import (
	"math"
	"math/rand"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestBoundsOfMatchesEncapsulating(t *testing.T) {
	rng := rand.New(rand.NewSource(7))
	elements := make([]elementReference, 400)
	for i := range elements {
		elements[i] = elementReference{bounds: geometry.NewAABB(
			vector3.New(rng.NormFloat64(), rng.NormFloat64(), rng.NormFloat64()).Scale(50),
			vector3.New(rng.Float64(), rng.Float64(), rng.Float64()).Scale(9),
		)}
	}

	expected := elements[0].bounds
	for _, e := range elements {
		expected.EncapsulateBounds(e.bounds)
	}

	assert.Equal(t, expected, boundsOf(elements))
}

func TestEveryElementSurvivesConstruction(t *testing.T) {
	rng := rand.New(rand.NewSource(11))
	elements := make([]Element, 5000)
	for i := range elements {
		p := vector3.New(rng.NormFloat64(), rng.NormFloat64(), rng.NormFloat64()).Normalized()
		elements[i] = BoundingBoxElement(geometry.NewAABB(p, vector3.Fill(0.05)))
	}

	tree := NewOctree(elements)

	seen := make(map[int]int)
	var walk func(*OctTree)
	walk = func(node *OctTree) {
		for _, e := range node.elements {
			seen[e.originalIndex]++
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(tree)

	assert.Len(t, seen, len(elements), "every element should be reachable exactly once")
	for index, count := range seen {
		assert.Equal(t, 1, count, "element %d", index)
	}

	whole := tree.BoundingBox()
	assert.LessOrEqual(t, whole.Min().X(), -1.0)
	assert.GreaterOrEqual(t, whole.Max().X(), 1.0)
	assert.False(t, math.IsNaN(whole.Center().X()))
}

func TestLeafSlicesCannotReachTheirNeighbours(t *testing.T) {
	rng := rand.New(rand.NewSource(19))
	elements := make([]Element, 4000)
	for i := range elements {
		p := vector3.New(rng.NormFloat64(), rng.NormFloat64(), rng.NormFloat64()).Normalized()
		elements[i] = BoundingBoxElement(geometry.NewAABB(p, vector3.Fill(0.04)))
	}

	tree := NewOctree(elements)

	var walk func(*OctTree)
	walk = func(node *OctTree) {
		assert.Equal(t, len(node.elements), cap(node.elements),
			"a node holding slack capacity could be appended into a sibling's run")
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(tree)
}
