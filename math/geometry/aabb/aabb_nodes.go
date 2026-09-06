package aabb

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

type NewNode struct {
	Center nodes.Output[vector3.Float64] `description:"Middle of the box. Defaults to the origin."`
	Size   nodes.Output[vector3.Float64] `description:"Full width, height and depth of the box - not the half-extents. Defaults to 1x1x1."`
}

func (n NewNode) Description() string {
	return "Builds an axis-aligned bounding box from a center and a full size (not half-extents)."
}

func (n NewNode) Out(out *nodes.StructOutput[geometry.AABB]) {
	out.Set(geometry.NewAABB(
		nodes.TryGetOutputValue(out, n.Center, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, n.Size, vector3.One[float64]()),
	))
}

// ============================================================================

type FromMinMaxNode struct {
	Min nodes.Output[vector3.Float64] `description:"Corner with the smallest x, y and z."`
	Max nodes.Output[vector3.Float64] `description:"Corner with the largest x, y and z."`
}

func (n FromMinMaxNode) Description() string {
	return "Builds an axis-aligned bounding box from its two opposite corners."
}

func (n FromMinMaxNode) Out(out *nodes.StructOutput[geometry.AABB]) {
	min := nodes.TryGetOutputValue(out, n.Min, vector3.Zero[float64]())
	max := nodes.TryGetOutputValue(out, n.Max, vector3.Zero[float64]())

	box := geometry.NewEmptyAABB()
	box.SetMinMax(vector3.Min(min, max), vector3.Max(min, max))
	out.Set(box)
}

// ============================================================================

type FromPointsNode struct {
	Points nodes.Output[[]vector3.Float64] `description:"Points the box must contain."`
}

func (n FromPointsNode) Description() string {
	return "Smallest axis-aligned bounding box containing every given point."
}

func (n FromPointsNode) Out(out *nodes.StructOutput[geometry.AABB]) {
	out.Set(geometry.NewAABBFromPoints(nodes.TryGetOutputValue(out, n.Points, nil)...))
}

// ============================================================================

type ExpandNode struct {
	AABB   nodes.Output[geometry.AABB] `description:"Box to grow."`
	Amount nodes.Output[float64]       `description:"How much to add to the box's total size on each axis, split evenly across both sides. Negative shrinks it."`
}

func (n ExpandNode) Description() string {
	return "Grows a bounding box outward on every side, keeping its center."
}

func (n ExpandNode) Out(out *nodes.StructOutput[geometry.AABB]) {
	box := nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB())
	box.Expand(nodes.TryGetOutputValue(out, n.Amount, 0))
	out.Set(box)
}

// ============================================================================

type SelectNode struct {
	AABB nodes.Output[geometry.AABB] `description:"Box to break apart."`
}

func (n SelectNode) Description() string {
	return "Splits a bounding box into its center, size and corners."
}

func (n SelectNode) Center(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB()).Center())
}

func (n SelectNode) Size(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB()).Size())
}

func (n SelectNode) Min(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB()).Min())
}

func (n SelectNode) Max(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB()).Max())
}

func (n SelectNode) Volume(out *nodes.StructOutput[float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, geometry.NewEmptyAABB()).Volume())
}
