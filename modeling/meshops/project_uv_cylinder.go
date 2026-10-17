package meshops

import (
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

const (
	cylinderSide = iota
	cylinderSideWrapped
	cylinderCapTop
	cylinderCapBottom
)

// A face tilted more than this toward the axis is a cap.
var cylinderCapCos = math.Cos(math.Pi / 4)

type cylinderFrame struct {
	center, axis, right, forward vector3.Float64
}

func newCylinderFrame(center, axis vector3.Float64) cylinderFrame {
	if axis.Length() == 0 {
		axis = vector3.Up[float64]()
	}
	axis = axis.Normalized()
	reference := vector3.Right[float64]()
	if math.Abs(axis.Dot(reference)) > 0.9 {
		reference = vector3.Forward[float64]()
	}
	right := reference.Sub(axis.Scale(axis.Dot(reference))).Normalized()
	return cylinderFrame{center: center, axis: axis, right: right, forward: axis.Cross(right)}
}

// Turns around the axis in [0, 1), increasing to the right seen from outside.
func (f cylinderFrame) turn(p vector3.Float64) float64 {
	d := p.Sub(f.center)
	t := math.Atan2(d.Dot(f.forward), d.Dot(f.right)) / (2 * math.Pi)
	if t < 0 {
		t++
	}
	return t
}

func CylinderProjectUV(m modeling.Mesh, center, axis vector3.Float64, around, height float64) (modeling.Mesh, error) {
	if !m.HasFloat3Attribute(modeling.PositionAttribute) || m.PrimitiveCount() == 0 {
		return m, nil
	}
	if m.Topology() != modeling.TriangleTopology {
		return m, errProjectionNeedsTriangles
	}
	frame := newCylinderFrame(center, axis)
	height = nonZero(height)
	around = nonZero(around)

	positions := m.Float3Attribute(modeling.PositionAttribute)
	groups := make([]int, m.PrimitiveCount())
	for triangle := range groups {
		n := triangleNormal(m, triangle)
		if length := n.Length(); length > 0 {
			switch along := n.Dot(frame.axis) / length; {
			case along > cylinderCapCos:
				groups[triangle] = cylinderCapTop
				continue
			case along < -cylinderCapCos:
				groups[triangle] = cylinderCapBottom
				continue
			}
		}
		tri := m.Tri(triangle)
		lo, hi := 1.0, 0.0
		for _, p := range []vector3.Float64{
			tri.P1Vec3Attr(modeling.PositionAttribute),
			tri.P2Vec3Attr(modeling.PositionAttribute),
			tri.P3Vec3Attr(modeling.PositionAttribute),
		} {
			t := frame.turn(p)
			lo, hi = math.Min(lo, t), math.Max(hi, t)
		}
		groups[triangle] = cylinderSide
		if hi-lo > 0.5 {
			groups[triangle] = cylinderSideWrapped
		}
	}

	split, sources, vertexGroups := splitByGroup(m, func(corner int) int { return groups[corner/3] })
	uvs := make([]vector2.Float64, len(sources))
	for i, source := range sources {
		p := positions.At(source)
		d := p.Sub(frame.center)
		switch vertexGroups[i] {
		case cylinderCapTop:
			uvs[i] = vector2.New(d.Dot(frame.right)/height, d.Dot(frame.forward)/height)
		case cylinderCapBottom:
			uvs[i] = vector2.New(-d.Dot(frame.right)/height, d.Dot(frame.forward)/height)
		default:
			t := frame.turn(p)
			if vertexGroups[i] == cylinderSideWrapped && t < 0.5 {
				t++
			}
			uvs[i] = vector2.New(t*around, d.Dot(frame.axis)/height)
		}
	}
	return split.SetFloat2Attribute(modeling.TexCoordAttribute, uvs), nil
}

type CylinderProjectUVNode struct {
	Mesh   nodes.Output[modeling.Mesh]
	Axis   nodes.Output[vector3.Float64] `description:"Direction the cylinder runs along. Defaults to +Y."`
	Center nodes.Output[vector3.Float64] `description:"A point on the axis. Defaults to the center of the mesh's bounds."`
	Around nodes.Output[float64]         `description:"Texture repeats once around the circumference this many times. Defaults to 1."`
	Height nodes.Output[float64]         `description:"World length of one texture repeat along the axis, and across the caps, in meters. Defaults to 1."`
}

func (CylinderProjectUVNode) Description() string {
	return "Gives a mesh UVs wrapped around an axis: u runs around, v along the axis, and faces pointing along the axis get a flat projection."
}

func (CylinderProjectUVNode) Keywords() []string {
	return []string{"unwrap", "wrap", "cylindrical", "tube"}
}

func (n CylinderProjectUVNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	out.Set(modeling.EmptyMesh(modeling.TriangleTopology))
	if n.Mesh == nil {
		return
	}
	m := nodes.GetOutputValue(out, n.Mesh)
	center := vector3.Zero[float64]()
	if m.HasFloat3Attribute(modeling.PositionAttribute) {
		center = m.BoundingBox(modeling.PositionAttribute).Center()
	}
	mesh, err := CylinderProjectUV(
		m,
		nodes.TryGetOutputValue(out, n.Center, center),
		nodes.TryGetOutputValue(out, n.Axis, vector3.Up[float64]()),
		nodes.TryGetOutputValue(out, n.Around, 1.),
		nodes.TryGetOutputValue(out, n.Height, 1.),
	)
	if err != nil {
		out.CaptureError(err)
		return
	}
	out.Set(mesh)
}
