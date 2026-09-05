package geometry

import (
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// AABB's literal parameter node takes center plus extents (half sizes),
// while these build one from live values. Size is used rather than extents
// throughout, because the numbers driving a box in a graph are almost
// always full dimensions - a part's width/height/depth variables - and
// halving them by hand at every wiring site is the easy thing to get wrong.

type AABBNode struct {
	Center nodes.Output[vector3.Float64] `description:"Middle of the box. Defaults to the origin."`
	Size   nodes.Output[vector3.Float64] `description:"Full width, height and depth of the box - not the half-extents. Defaults to 1x1x1."`
}

func (n AABBNode) Description() string {
	return "Builds an axis-aligned bounding box from a center and a full size (not half-extents)."
}

func (n AABBNode) Out(out *nodes.StructOutput[AABB]) {
	out.Set(NewAABB(
		nodes.TryGetOutputValue(out, n.Center, vector3.Zero[float64]()),
		nodes.TryGetOutputValue(out, n.Size, vector3.One[float64]()),
	))
}

// ============================================================================

type AABBFromMinMaxNode struct {
	Min nodes.Output[vector3.Float64] `description:"Corner with the smallest x, y and z."`
	Max nodes.Output[vector3.Float64] `description:"Corner with the largest x, y and z."`
}

func (n AABBFromMinMaxNode) Description() string {
	return "Builds an axis-aligned bounding box from its two opposite corners."
}

func (n AABBFromMinMaxNode) Out(out *nodes.StructOutput[AABB]) {
	min := nodes.TryGetOutputValue(out, n.Min, vector3.Zero[float64]())
	max := nodes.TryGetOutputValue(out, n.Max, vector3.Zero[float64]())

	box := NewEmptyAABB()
	box.SetMinMax(vector3.Min(min, max), vector3.Max(min, max))
	out.Set(box)
}

// ============================================================================

type AABBFromPointsNode struct {
	Points nodes.Output[[]vector3.Float64] `description:"Points the box must contain."`
}

func (n AABBFromPointsNode) Description() string {
	return "Smallest axis-aligned bounding box containing every given point."
}

func (n AABBFromPointsNode) Out(out *nodes.StructOutput[AABB]) {
	out.Set(NewAABBFromPoints(nodes.TryGetOutputValue(out, n.Points, nil)...))
}

// ============================================================================

type ExpandAABBNode struct {
	AABB   nodes.Output[AABB]    `description:"Box to grow."`
	Amount nodes.Output[float64] `description:"How much to add to the box's total size on each axis, split evenly across both sides. Negative shrinks it."`
}

func (n ExpandAABBNode) Description() string {
	return "Grows a bounding box outward on every side, keeping its center."
}

func (n ExpandAABBNode) Out(out *nodes.StructOutput[AABB]) {
	box := nodes.TryGetOutputValue(out, n.AABB, NewEmptyAABB())
	box.Expand(nodes.TryGetOutputValue(out, n.Amount, 0))
	out.Set(box)
}

// ============================================================================

type AABBPropertiesNode struct {
	AABB nodes.Output[AABB] `description:"Box to break apart."`
}

func (n AABBPropertiesNode) Description() string {
	return "Splits a bounding box into its center, size and corners."
}

func (n AABBPropertiesNode) Center(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, NewEmptyAABB()).Center())
}

func (n AABBPropertiesNode) Size(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, NewEmptyAABB()).Size())
}

func (n AABBPropertiesNode) Min(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, NewEmptyAABB()).Min())
}

func (n AABBPropertiesNode) Max(out *nodes.StructOutput[vector3.Float64]) {
	out.Set(nodes.TryGetOutputValue(out, n.AABB, NewEmptyAABB()).Max())
}
