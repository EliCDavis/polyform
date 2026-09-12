package aabb_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/geometry/aabb"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildAABB(center, size vector3.Float64) geometry.AABB {
	return nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.NewNode]{
		Data: aabb.NewNode{
			Center: nodes.ConstOutput[vector3.Float64]{Val: center},
			Size:   nodes.ConstOutput[vector3.Float64]{Val: size},
		},
	}, "Out").Value()
}

func TestAABBNodeTakesFullSizeNotExtents(t *testing.T) {
	box := buildAABB(vector3.Zero[float64](), vector3.New(2., 4., 6.))

	assert.Equal(t, vector3.New(2., 4., 6.), box.Size(), "Size in should be Size out")
	assert.Equal(t, vector3.New(-1., -2., -3.), box.Min(), "half the size should land on each side")
	assert.Equal(t, vector3.New(1., 2., 3.), box.Max())
}

func TestAABBNodeMatchesLiteralEncoding(t *testing.T) {
	computed := buildAABB(vector3.New(0., 1., 0.), vector3.New(2., 1., 2.))

	var literal geometry.AABB
	require.NoError(t, json.Unmarshal(
		[]byte(`{"center":{"x":0,"y":1,"z":0},"extents":{"x":1,"y":0.5,"z":1}}`),
		&literal,
	))

	assert.Equal(t, literal.Center(), computed.Center())
	assert.Equal(t, literal.Size(), computed.Size())
	assert.Equal(t, literal.Min(), computed.Min())
	assert.Equal(t, literal.Max(), computed.Max())
}

func TestAABBNodeDefaults(t *testing.T) {
	box := nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.NewNode]{
		Data: aabb.NewNode{},
	}, "Out").Value()

	assert.Equal(t, vector3.Zero[float64](), box.Center(), "an unwired box should sit at the origin")
	assert.Equal(t, vector3.One[float64](), box.Size(), "an unwired box should be unit sized, not empty")
}

func TestAABBFromMinMaxNode(t *testing.T) {
	box := nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.FromMinMaxNode]{
		Data: aabb.FromMinMaxNode{
			Min: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(-1., 0., -2.)},
			Max: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(3., 4., 2.)},
		},
	}, "Out").Value()

	assert.Equal(t, vector3.New(-1., 0., -2.), box.Min())
	assert.Equal(t, vector3.New(3., 4., 2.), box.Max())
	assert.Equal(t, vector3.New(4., 4., 4.), box.Size())
}

func TestAABBFromMinMaxNodeToleratesSwappedCorners(t *testing.T) {
	box := nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.FromMinMaxNode]{
		Data: aabb.FromMinMaxNode{
			Min: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(3., 4., 2.)},
			Max: nodes.ConstOutput[vector3.Float64]{Val: vector3.New(-1., 0., -2.)},
		},
	}, "Out").Value()

	assert.Equal(t, vector3.New(-1., 0., -2.), box.Min())
	assert.Equal(t, vector3.New(4., 4., 4.), box.Size())
}

func TestExpandAABBNodeGrowsTotalSize(t *testing.T) {
	expanded := nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.ExpandNode]{
		Data: aabb.ExpandNode{
			AABB:   nodes.ConstOutput[geometry.AABB]{Val: buildAABB(vector3.Zero[float64](), vector3.New(2., 2., 2.))},
			Amount: nodes.ConstOutput[float64]{Val: 1},
		},
	}, "Out").Value()

	assert.Equal(t, vector3.New(3., 3., 3.), expanded.Size(), "expanding by 1 should add 1 to each total dimension")
	assert.Equal(t, vector3.Zero[float64](), expanded.Center(), "expanding should keep the center put")
}

func TestAABBPropertiesNodeRoundTrips(t *testing.T) {
	original := buildAABB(vector3.New(1., 2., 3.), vector3.New(4., 6., 8.))

	props := &nodes.Struct[aabb.SelectNode]{
		Data: aabb.SelectNode{AABB: nodes.ConstOutput[geometry.AABB]{Val: original}},
	}

	center := nodes.GetNodeOutputPort[vector3.Float64](props, "Center").Value()
	size := nodes.GetNodeOutputPort[vector3.Float64](props, "Size").Value()

	assert.Equal(t, vector3.New(1., 2., 3.), center)
	assert.Equal(t, vector3.New(4., 6., 8.), size)

	// Feeding the pieces back in must reproduce the box exactly, so a
	// domain can be measured, adjusted and rebuilt without drift.
	assert.Equal(t, original, buildAABB(center, size))
}

func TestComputedDomainTracksItsInputs(t *testing.T) {
	partSize := func(w, h, d float64) geometry.AABB {
		return nodes.GetNodeOutputPort[geometry.AABB](&nodes.Struct[aabb.ExpandNode]{
			Data: aabb.ExpandNode{
				AABB:   nodes.ConstOutput[geometry.AABB]{Val: buildAABB(vector3.Zero[float64](), vector3.New(w, h, d))},
				Amount: nodes.ConstOutput[float64]{Val: 0.5},
			},
		}, "Out").Value()
	}

	small := partSize(1, 1, 1)
	tripled := partSize(3, 3, 3)

	assert.Equal(t, vector3.New(1.5, 1.5, 1.5), small.Size())
	assert.Equal(t, vector3.New(3.5, 3.5, 3.5), tripled.Size(),
		"tripling the part's dimensions must grow the domain with it")
}
