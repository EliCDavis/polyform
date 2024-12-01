package repeat

import (
	"github.com/EliCDavis/polyform/math/curves"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Spline(m modeling.Mesh, curve curves.Spline, inbetween int) modeling.Mesh {
	start := m.
		Rotate(quaternion.RotationTo(vector3.Forward[float64](), curve.Dir(0))).
		Translate(curve.At(0))

	dist := curve.Length()
	end := m.
		Rotate(quaternion.RotationTo(vector3.Forward[float64](), curve.Dir(dist))).
		Translate(curve.At(dist))

	return SplineExlusive(m, curve, inbetween).Append(end).Append(start)
}

// Like line, but we don't include meshes on the start and end points. Only the
// inbetween points
func SplineExlusive(m modeling.Mesh, curve curves.Spline, inbetween int) modeling.Mesh {

	inc := curve.Length() / float64(inbetween+1)

	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)

	for i := 1; i <= inbetween; i++ {
		dist := inc * float64(i)
		dir := curve.Dir(dist)
		rotatedMesh := m.
			Rotate(quaternion.RotationTo(vector3.Forward[float64](), dir)).
			Translate(curve.At(dist))
		finalMesh = finalMesh.Append(rotatedMesh)
	}

	return finalMesh
}

type SplineNode = nodes.StructNode[modeling.Mesh, SplineNodeData]

type SplineNodeData struct {
	Mesh  nodes.NodeOutput[modeling.Mesh]
	Curve nodes.NodeOutput[curves.Spline]
	Times nodes.NodeOutput[int]
}

func (r SplineNodeData) Process() (modeling.Mesh, error) {
	if r.Mesh == nil || r.Curve == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	times := 0
	if r.Times != nil {
		times = r.Times.Value()
	}

	if times <= 0 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	curve := r.Curve.Value()
	if curve == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	mesh := r.Mesh.Value()
	if times == 1 {
		SplineExlusive(mesh, curve, 1)
	}

	if times == 2 {
		Spline(mesh, curve, 0)
	}

	return Spline(mesh, curve, times-2), nil
}
