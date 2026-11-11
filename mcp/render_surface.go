package mcp

import (
	"math"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/fogleman/fauxgl"
)

// glTF puts V=0 on an image's top row; fauxgl samples it from the bottom.
type gltfTexture struct{ fauxgl.Texture }

func (t gltfTexture) Sample(u, v float64) fauxgl.Color {
	return t.Texture.Sample(u, 1-v)
}

func (t gltfTexture) BilinearSample(u, v float64) fauxgl.Color {
	return t.Texture.BilinearSample(u, 1-v)
}

func imageTexture(tex *gltf.PolyformTexture) fauxgl.Texture {
	if tex == nil || tex.Image == nil {
		return nil
	}
	return gltfTexture{fauxgl.NewImageTexture(tex.Image)}
}

type tangentFrame struct {
	tangent fauxgl.Vector
	// Points toward the top of the image, the way a glTF normal map's green channel does.
	bitangent fauxgl.Vector
}

func faceTangentFrame(p1, p2, p3 vector3.Float64, uv1, uv2, uv3 vector2.Float64) tangentFrame {
	e1, e2 := p2.Sub(p1), p3.Sub(p1)
	d1, d2 := uv2.Sub(uv1), uv3.Sub(uv1)
	r := d1.X()*d2.Y() - d2.X()*d1.Y()
	if math.Abs(r) < 1e-14 {
		return tangentFrame{}
	}
	dPdu := e1.Scale(d2.Y()).Sub(e2.Scale(d1.Y())).DivByConstant(r)
	dPdv := e2.Scale(d1.X()).Sub(e1.Scale(d2.X())).DivByConstant(r)
	return tangentFrame{
		tangent:   fauxgl.V(dPdu.X(), dPdu.Y(), dPdu.Z()),
		bitangent: fauxgl.V(-dPdv.X(), -dPdv.Y(), -dPdv.Z()),
	}
}

// A triangle's index rides in its vertices' Texture.Z, which interpolates to that same constant.
type surface struct {
	normal            fauxgl.Texture
	normalScale       float64
	tangents          []tangentFrame
	occlusion         fauxgl.Texture
	occlusionStrength float64
	emissive          fauxgl.Color
	emissiveTexture   fauxgl.Texture
	layers            layers
}

// Stand-ins for extensions a rasterizer can't do: no refraction or scattering.
type layers struct {
	transmission   float64
	tint           fauxgl.Color
	coat           float64
	coatRoughness  float64
	sheen          fauxgl.Color
	sheenTexture   fauxgl.Texture
	sheenRoughness float64
}

func (l layers) at(v fauxgl.Vertex) layers {
	if l.sheenTexture != nil {
		l.sheen = l.sheen.Mul(l.sheenTexture.BilinearSample(v.Texture.X, v.Texture.Y))
		l.sheen.A = 0
	}
	return l
}

func (l layers) transmissive() bool { return l.transmission > 0 }

func materialLayers(mat *gltf.PolyformMaterial) layers {
	l := layers{tint: fauxgl.Color{R: 1, G: 1, B: 1, A: 1}}
	for _, ext := range mat.Extensions {
		switch e := ext.(type) {
		case gltf.PolyformTransmission:
			l.transmission = math.Min(math.Max(e.Factor, 0), 1)
		case gltf.PolyformVolume:
			if e.AttenuationColor != nil && e.ThicknessFactor > 0 {
				l.tint = fauxgl.MakeColor(e.AttenuationColor)
			}
		case gltf.PolyformClearcoat:
			l.coat = math.Min(math.Max(e.ClearcoatFactor, 0), 1)
			l.coatRoughness = math.Min(math.Max(e.ClearcoatRoughnessFactor, 0), 1)
		case gltf.PolyformSheen:
			if e.SheenColorFactor != nil {
				l.sheen = fauxgl.MakeColor(e.SheenColorFactor)
				l.sheen.A = 0
			}
			if tex := imageTexture(e.SheenColorTexture); tex != nil {
				l.sheenTexture = srgbTexture{tex}
			}
			l.sheenRoughness = math.Min(math.Max(e.SheenRoughnessFactor, 0), 1)
		}
	}
	return l
}

func (l layers) apply(shaded, albedo fauxgl.Color, normal, toCamera, toLight fauxgl.Vector) (fauxgl.Color, float64) {
	facing := math.Max(normal.Dot(toCamera), 0)
	grazing := math.Pow(1-facing, 5)
	lit := math.Max(normal.Dot(toLight), 0)
	alpha := 1.

	if l.sheen != (fauxgl.Color{}) {
		rim := math.Pow(1-facing, 2.5-1.5*l.sheenRoughness)
		shaded = shaded.Add(l.sheen.MulScalar((0.06 + 0.7*rim) * (0.5 + 0.5*lit)))
	}

	if l.transmission > 0 {
		edge := math.Pow(1-facing, 2)
		reflected := toLight.Negate().Reflect(normal)
		glint := math.Pow(math.Max(toCamera.Dot(reflected), 0), 300) * lit
		through := albedo.Mul(l.tint)
		shaded = shaded.Lerp(through, 0.6*l.transmission)
		shaded = shaded.Add(fauxgl.Color{R: 1, G: 1, B: 1}.MulScalar((0.45*edge + 1.5*glint) * l.transmission))
		alpha = 1 - l.transmission*(1-edge)*0.7
	}

	if l.coat > 0 {
		fresnel := 0.04 + 0.96*grazing
		spread := math.Max(l.coatRoughness*l.coatRoughness, 0.002)
		shininess := math.Min((2/(spread*spread)-2)/4, 1024)
		reflected := toLight.Negate().Reflect(normal)
		highlight := math.Pow(math.Max(toCamera.Dot(reflected), 0), shininess) * lit
		up := toCamera.Negate().Reflect(normal).Y
		band := 0.05 + l.coatRoughness
		sky := 0.15 + math.Min(math.Max((up+band)/(2*band), 0), 1)
		gloss := fresnel*sky + 0.8*highlight
		shaded = shaded.MulScalar(1 - l.coat*fresnel).Add(fauxgl.Color{R: 1, G: 1, B: 1}.MulScalar(l.coat * gloss))
	}

	return shaded, alpha
}

func (s surface) readsUVs() bool {
	return s.normal != nil || s.occlusion != nil || s.emissiveTexture != nil || s.layers.sheenTexture != nil
}

func materialSurface(mat *gltf.PolyformMaterial) surface {
	var s surface
	if mat == nil {
		return s
	}
	s.layers = materialLayers(mat)
	if mat.NormalTexture != nil {
		s.normal = imageTexture(mat.NormalTexture.PolyformTexture)
		s.normalScale = 1
		if mat.NormalTexture.Scale != nil {
			s.normalScale = *mat.NormalTexture.Scale
		}
	}
	if mat.OcclusionTexture != nil {
		s.occlusion = imageTexture(mat.OcclusionTexture.PolyformTexture)
		s.occlusionStrength = 1
		if mat.OcclusionTexture.Strength != nil {
			s.occlusionStrength = *mat.OcclusionTexture.Strength
		}
	}
	if mat.EmissiveFactor != nil {
		strength := 1.
		for _, ext := range mat.Extensions {
			if e, ok := ext.(gltf.PolyformEmissiveStrength); ok && e.EmissiveStrength != nil {
				strength = *e.EmissiveStrength
			}
		}
		s.emissive = fauxgl.MakeColor(mat.EmissiveFactor).MulScalar(strength)
		s.emissive.A = 0
		if tex := imageTexture(mat.EmissiveTexture); tex != nil {
			s.emissiveTexture = srgbTexture{tex}
		}
	}
	return s
}

func (s surface) triangleFrame(v fauxgl.Vertex) (tangentFrame, bool) {
	i := int(math.Round(v.Texture.Z))
	if i < 0 || i >= len(s.tangents) {
		return tangentFrame{}, false
	}
	return s.tangents[i], true
}

func (s surface) shadingNormal(v fauxgl.Vertex) fauxgl.Vector {
	n := v.Normal
	if s.normal == nil {
		return n
	}
	frame, ok := s.triangleFrame(v)
	if !ok {
		return n
	}
	t := frame.tangent.Sub(n.MulScalar(n.Dot(frame.tangent)))
	if t.Length() < 1e-12 {
		return n
	}
	t = t.Normalize()
	b := n.Cross(t)
	if b.Dot(frame.bitangent) < 0 {
		b = b.Negate()
	}

	c := s.normal.BilinearSample(v.Texture.X, v.Texture.Y)
	perturbed := t.MulScalar((2*c.R - 1) * s.normalScale).
		Add(b.MulScalar((2*c.G - 1) * s.normalScale)).
		Add(n.MulScalar(2*c.B - 1))
	if perturbed.Length() < 1e-12 {
		return n
	}
	return perturbed.Normalize()
}

func (s surface) occlusionAt(v fauxgl.Vertex) float64 {
	if s.occlusion == nil {
		return 1
	}
	return 1 + s.occlusionStrength*(s.occlusion.BilinearSample(v.Texture.X, v.Texture.Y).R-1)
}

func (s surface) emissionAt(v fauxgl.Vertex) fauxgl.Color {
	if s.emissiveTexture == nil {
		return s.emissive
	}
	e := s.emissive.Mul(s.emissiveTexture.BilinearSample(v.Texture.X, v.Texture.Y))
	e.A = 0
	return e
}

type uvCheckerTexture struct{}

func (uvCheckerTexture) Sample(u, v float64) fauxgl.Color {
	fu, fv := u-math.Floor(u), v-math.Floor(v)
	shade := 1.
	if (int(math.Floor(fu*8))+int(math.Floor(fv*8)))%2 == 1 {
		shade = 0.3
	}
	return fauxgl.Color{R: (0.15 + 0.85*fu) * shade, G: (0.15 + 0.85*fv) * shade, B: 0.5 * shade, A: 1}
}

func (c uvCheckerTexture) BilinearSample(u, v float64) fauxgl.Color {
	return c.Sample(u, v)
}
