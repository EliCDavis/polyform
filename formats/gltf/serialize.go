package gltf

import (
	"github.com/EliCDavis/polyform/generator/manifest"
	"github.com/EliCDavis/polyform/generator/serialize"
	"github.com/EliCDavis/polyform/modeling"
)

func RegisterOutputSerializers(ts *serialize.TypeSwitch[manifest.Entry]) {
	serialize.Register(ts, func(m modeling.Mesh) manifest.Entry {
		return glb([]*PolyformModel{{Name: "Mesh", Mesh: &m}})
	})
	serialize.Register(ts, func(ms []modeling.Mesh) manifest.Entry {
		models := make([]*PolyformModel, 0, len(ms))
		for i := range ms {
			models = append(models, &PolyformModel{Name: "Mesh", Mesh: &ms[i]})
		}
		return glb(models)
	})
	serialize.Register(ts, func(m *PolyformModel) manifest.Entry {
		if m == nil {
			return glb(nil)
		}
		return glb([]*PolyformModel{m})
	})
	serialize.Register(ts, func(ms []*PolyformModel) manifest.Entry {
		return glb(ms)
	})
}

func glb(models []*PolyformModel) manifest.Entry {
	return manifest.Entry{Artifact: Artifact{Scene: PolyformScene{Models: models}}}
}
