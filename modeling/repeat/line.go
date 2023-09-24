package repeat

import (
	"github.com/EliCDavis/polyform/modeling"
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
