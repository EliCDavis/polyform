package operators

import (
	"math"

	"github.com/EliCDavis/polyform/modeling/pipeline"
	"github.com/EliCDavis/vector/vector3"
)

type CenterAttribute3DCommand struct {
	writePermission pipeline.MeshWritePermission
	attribute       string
}

func NewCenterAttribute3DCommand(attribute string) CenterAttribute3DCommand {
	return CenterAttribute3DCommand{
		attribute: attribute,
		writePermission: pipeline.MeshWritePermission{
			V3Permissions: map[string]pipeline.WriteArrayPermission[vector3.Float64]{
				attribute: {},
			},
		},
	}
}

func (ca3dc CenterAttribute3DCommand) ReadPermissions() pipeline.MeshReadPermission {
	return pipeline.MeshReadPermission{}
}

func (ca3dc CenterAttribute3DCommand) WritePermissions() pipeline.MeshWritePermission {
	return ca3dc.writePermission
}


func (ca3dc CenterAttribute3DCommand) Run() {
	attributeData := ca3dc.writePermission.V3Permissions[ca3dc.attribute].Data()
	if len(attributeData) == 0 {
		return
	}

	min := vector3.New(math.Inf(1), math.Inf(1), math.Inf(1))
	max := vector3.New(math.Inf(-1), math.Inf(-1), math.Inf(-1))
	for i := 0; i < len(attributeData); i++ {
		v := attributeData[i]
		min = vector3.Min(min, v)
		max = vector3.Max(max, v)
	}

	center := min.Midpoint(max)
	for i := 0; i < len(attributeData); i++ {
		attributeData[i] = attributeData[i].Sub(center)
	}
}

