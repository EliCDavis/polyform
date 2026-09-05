package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Cone struct {
	Height float64
	Radius float64
	Sides  int
}

func (c Cone) ToMesh() modeling.Mesh {
	if c.Sides < 3 {
		panic("can not make cone with less that 3 sides")
	}

	verts := repeat.CirclePoints(c.Sides, c.Radius)
	lastVert := len(verts)

	normals := make([]vector3.Float64, 0, len(verts)+1)
	for _, v := range verts {
		radial := vector3.New(v.X(), 0., v.Z())
		if radial.Length() == 0 {
			normals = append(normals, vector3.Up[float64]())
			continue
		}
		normals = append(normals, radial.Normalized().
			Scale(c.Height).
			Add(vector3.Up[float64]().Scale(c.Radius)).
			Normalized())
	}
	normals = append(normals, vector3.Up[float64]())

	verts = append(verts, vector3.New(0., c.Height, 0.))
	uvs := make([]vector2.Float64, len(verts))
	uvs[len(uvs)-1] = vector2.One[float64]()

	tris := make([]int, 0, c.Sides*3)
	for i := 0; i < c.Sides; i++ {
		tris = append(tris, i, lastVert, i+1)
	}
	tris[len(tris)-1] = 0

	return modeling.NewMesh(modeling.TriangleTopology, tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat3Attribute(modeling.NormalAttribute, normals).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

}

type ConeNode struct {
	Height nodes.Output[float64]
	Radius nodes.Output[float64]
	Sides  nodes.Output[int]
}

func (r ConeNode) Description() string {
	return "A cone standing on its base, point up the Y axis. No bottom cap. A low Sides count gives a faceted spike; 4 gives a pyramid."
}

func (r ConeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	cone := Cone{
		Height: nodes.TryGetOutputValue(out, r.Height, 1),
		Radius: nodes.TryGetOutputValue(out, r.Radius, 0.5),
		Sides:  max(nodes.TryGetOutputValue(out, r.Sides, 3), 3),
	}
	out.Set(cone.ToMesh())
}
