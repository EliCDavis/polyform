package operators

import (
	"github.com/EliCDavis/polyform/modeling/pipeline"
	"github.com/EliCDavis/vector/vector3"
)

type SmoothLaplacianCommand struct {
	readPermission  pipeline.MeshReadPermission
	writePermission pipeline.MeshWritePermission
	attribute       string
	iterations      int
	smoothingFactor float64
}

func NewSmoothLaplacianCommand(attribute string) SmoothLaplacianCommand {
	return SmoothLaplacianCommand{
		attribute: attribute,
		readPermission: pipeline.MeshReadPermission{
			Indices: &pipeline.ReadIndicesPermission{},
		},
		writePermission: pipeline.MeshWritePermission{
			V3Permissions: map[string]pipeline.WriteArrayPermission[vector3.Float64]{
				attribute: {},
			},
		},
	}
}

func (slc SmoothLaplacianCommand) ReadPermissions() pipeline.MeshReadPermission {
	return slc.readPermission
}

func (slc SmoothLaplacianCommand) WritePermissions() pipeline.MeshWritePermission {
	return slc.writePermission
}

func (slc SmoothLaplacianCommand) Run() {
	attributeData := slc.writePermission.V3Permissions[slc.attribute].Data()
	if len(attributeData) == 0 {
		return
	}

	lut := slc.readPermission.Indices.VertexNeighborTable()

	for i := 0; i < slc.iterations; i++ {
		for vi, vertex := range attributeData {
			vs := vector3.Zero[float64]()

			for vn := range lut.Lookup(vi) {
				vs = vs.Add(attributeData[vn])
			}

			attributeData[vi] = vertex.Add(
				vs.
					DivByConstant(float64(lut.Count(vi))).
					Sub(vertex).
					Scale(slc.smoothingFactor))
		}
	}
}
