package enum

import (
	"github.com/EliCDavis/polyform/generator/registry"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

func init() {
	factory := &refutil.TypeFactory{}
	refutil.RegisterType[nodes.Struct[MemberNameNode]](factory)
	refutil.RegisterType[nodes.Struct[MemberIndexNode]](factory)
	registry.Register(factory)
}
