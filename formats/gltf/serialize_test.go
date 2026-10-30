package gltf_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeshAndModelOutputsSerializeAsGLB(t *testing.T) {
	ts := &serialize.TypeSwitch[manifest.Entry]{}
	gltf.RegisterOutputSerializers(ts)

	cube := primitives.UnitCube()
	ports := []nodes.OutputPort{
		nodes.ConstOutput[modeling.Mesh]{Val: cube},
		nodes.ConstOutput[[]modeling.Mesh]{Val: []modeling.Mesh{cube, cube}},
		nodes.ConstOutput[*gltf.PolyformModel]{Val: &gltf.PolyformModel{Name: "Cube", Mesh: &cube}},
		nodes.ConstOutput[[]*gltf.PolyformModel]{Val: []*gltf.PolyformModel{{Name: "Cube", Mesh: &cube}}},
	}

	for _, port := range ports {
		entry := ts.Run(port)
		assert.Equal(t, "model/gltf-binary", entry.Artifact.Mime())

		var out bytes.Buffer
		require.NoError(t, entry.Artifact.Write(&out))
		assert.Equal(t, "glTF", out.String()[:4])
	}
}
