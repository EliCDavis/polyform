package primitives

import (
	"bytes"
	_ "embed"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[CubeNode]](factory)
	refutil.RegisterType[nodes.Struct[CubeUVsNode]](factory)
	refutil.RegisterType[nodes.Struct[QuadNode]](factory)
	refutil.RegisterType[nodes.Struct[StripUVsNode]](factory)

	refutil.RegisterType[nodes.Struct[CylinderNode]](factory)
	refutil.RegisterType[nodes.Struct[HemisphereNode]](factory)
	refutil.RegisterType[nodes.Struct[UvSphereNode]](factory)

	refutil.RegisterType[nodes.Struct[CircleNode]](factory)
	refutil.RegisterType[nodes.Struct[CircleUVsNode]](factory)
	refutil.RegisterType[nodes.Struct[ConeNode]](factory)

	refutil.RegisterType[nodes.Struct[StanfordBunny]](factory)

	refutil.RegisterType[nodes.Struct[TorusNode]](factory)
	refutil.RegisterType[nodes.Struct[TorusUVNode]](factory)

	generator.RegisterTypes(factory)
}

//go:embed stanford-bunny.ply
var bunnyPLY []byte

type StanfordBunny struct {
}

func (c StanfordBunny) Bunny(out *nodes.StructOutput[modeling.Mesh]) {
	bunny, err := ply.ReadMesh(bytes.NewReader(bunnyPLY))
	if err != nil {
		out.CaptureError(err)
	}
	out.Set(*bunny)
}
