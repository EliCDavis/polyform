package repeat

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func Line(m modeling.Mesh, start, end vector3.Float64, inbetween int) modeling.Mesh {
	return LineExlusive(m, start, end, inbetween).Append(m.Translate(end)).Append(m.Translate(start))
}

// Like line, but we don't include meshes on the start and end points. Only the
// inbetween points
func LineExlusive(m modeling.Mesh, start, end vector3.Float64, inbetween int) modeling.Mesh {

	dir := end.Sub(start)
	inc := dir.DivByConstant(float64(inbetween + 1))

	finalMesh := modeling.EmptyMesh(modeling.TriangleTopology)

	for i := 1; i <= inbetween; i++ {
		finalMesh = finalMesh.Append(m.Translate(start.Add(inc.Scale(float64(i)))))
	}

	return finalMesh
}

type LineNode = nodes.StructNode[modeling.Mesh, LineNodeData]

type LineNodeData struct {
	Mesh  nodes.NodeOutput[modeling.Mesh]
	Start nodes.NodeOutput[vector3.Float64]
	End   nodes.NodeOutput[vector3.Float64]
	Times nodes.NodeOutput[int]
}

func (r LineNodeData) Process() (modeling.Mesh, error) {
	if r.Mesh == nil || r.Start == nil || r.End == nil {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	times := 0
	if r.Times != nil {
		times = r.Times.Value()
	}

	if times <= 0 {
		return modeling.EmptyMesh(modeling.TriangleTopology), nil
	}

	mesh := r.Mesh.Value()
	start := r.Start.Value()
	end := r.End.Value()

	if times == 1 {
		LineExlusive(mesh, start, end, 1)
	}

	if times == 2 {
		Line(mesh, start, end, 0)
	}

	return Line(mesh, start, end, times-2), nil
}
