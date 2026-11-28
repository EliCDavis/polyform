package primitives

import (
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/repeat"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type Cone struct {
	Height         float64
	Radius         float64
	Sides          int
	HeightSegments int
}

func (c Cone) sideNormal(v vector3.Float64) vector3.Float64 {
	radial := vector3.New(v.X(), 0., v.Z())
	if radial.Length() == 0 {
		return vector3.Up[float64]()
	}
	return radial.Normalized().
		Scale(c.Height).
		Add(vector3.Up[float64]().Scale(c.Radius)).
		Normalized()
}

func (c Cone) ToMesh() modeling.Mesh {
	if c.Sides < 3 {
		panic("can not make cone with less that 3 sides")
	}
	segments := max(c.HeightSegments, 1)

	verts := make([]vector3.Float64, 0, c.Sides*(segments+1)+2)
	normals := make([]vector3.Float64, 0, c.Sides*(segments+1)+2)
	for ring := 0; ring < segments; ring++ {
		t := float64(ring) / float64(segments)
		for _, v := range repeat.CirclePoints(c.Sides, c.Radius*(1-t)) {
			verts = append(verts, v.SetY(c.Height*t))
			normals = append(normals, c.sideNormal(v))
		}
	}
	apex := len(verts)
	verts = append(verts, vector3.New(0., c.Height, 0.))
	normals = append(normals, vector3.Up[float64]())

	base := len(verts)
	verts = append(verts, verts[:c.Sides]...)
	verts = append(verts, vector3.Zero[float64]())
	for range c.Sides + 1 {
		normals = append(normals, vector3.Down[float64]())
	}
	baseCentre := len(verts) - 1

	uvs := make([]vector2.Float64, len(verts))
	uvs[apex] = vector2.One[float64]()

	tris := make([]int, 0, c.Sides*3*2*segments)
	for i := 0; i < c.Sides; i++ {
		tris = append(tris, baseCentre, base+i, base+(i+1)%c.Sides)
	}
	for ring := 0; ring < segments-1; ring++ {
		lower, upper := ring*c.Sides, (ring+1)*c.Sides
		for i := 0; i < c.Sides; i++ {
			next := (i + 1) % c.Sides
			tris = append(tris,
				lower+i, upper+i, lower+next,
				lower+next, upper+i, upper+next,
			)
		}
	}
	top := (segments - 1) * c.Sides
	for i := 0; i < c.Sides; i++ {
		tris = append(tris, top+i, apex, top+(i+1)%c.Sides)
	}

	return modeling.NewMesh(modeling.TriangleTopology, tris).
		SetFloat3Attribute(modeling.PositionAttribute, verts).
		SetFloat3Attribute(modeling.NormalAttribute, normals).
		SetFloat2Attribute(modeling.TexCoordAttribute, uvs)

}

type ConeNode struct {
	Height         nodes.Output[float64] `default:"1"`
	Radius         nodes.Output[float64] `default:"0.5"`
	Sides          nodes.Output[int]     `default:"3"`
	HeightSegments nodes.Output[int]     `description:"Rings of vertices between the base and the tip." default:"1"`
}

func (r ConeNode) Description() string {
	return "A closed cone standing on its base, point up the Y axis."
}

func (r ConeNode) Keywords() []string { return []string{"pyramid", "spike"} }

func (r ConeNode) Out(out *nodes.StructOutput[modeling.Mesh]) {
	cone := Cone{
		Height:         nodes.GetOutputValue(out, r.Height),
		Radius:         nodes.GetOutputValue(out, r.Radius),
		Sides:          max(nodes.GetOutputValue(out, r.Sides), 3),
		HeightSegments: max(nodes.GetOutputValue(out, r.HeightSegments), 1),
	}
	out.Set(cone.ToMesh())
}
