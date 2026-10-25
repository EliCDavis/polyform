package extrude

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/gobold"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

var fontData = map[string][]byte{
	"sans":      goregular.TTF,
	"sans bold": gobold.TTF,
	"mono":      gomono.TTF,
	"mono bold": gomonobold.TTF,
}

var parsedFonts sync.Map

func loadFont(name string) (*truetype.Font, error) {
	if f, ok := parsedFonts.Load(name); ok {
		return f.(*truetype.Font), nil
	}
	data, ok := fontData[name]
	if !ok {
		return nil, fmt.Errorf("unknown font %q", name)
	}
	f, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}
	parsedFonts.Store(name, f)
	return f, nil
}

type Text struct {
	Text string

	// Height of a capital letter.
	Height float64

	// The solid spans 0 to Depth along the extrusion axis; negative goes the
	// other way. Zero leaves a single face.
	Depth float64

	// sans, sans bold, mono or mono bold. Empty is sans.
	Font string

	// "left", "center" or "right": which end of each line sits on the origin.
	Align string

	// Straight segments each curve of a letter's outline is drawn with.
	CurveSegments int
}

type placedGlyph struct {
	index truetype.Index
	x     float64
}

// The text lies in the YZ plane facing +X, its lines running along Y.
func (t Text) MeshX() (modeling.Mesh, error) {
	return t.mesh(RegionX)
}

// The text lies in the XZ plane facing +Y, its lines running along X and the
// tops of its letters toward -Z.
func (t Text) MeshY() (modeling.Mesh, error) {
	return t.mesh(func(region triangulation.Region, depth float64) modeling.Mesh {
		return extrudeRegion(region, depth, func(v vector3.Float64) vector3.Float64 {
			return vector3.New(v.X(), v.Z(), -v.Y())
		})
	})
}

// The text lies in the XY plane facing +Z, its lines running along X.
func (t Text) MeshZ() (modeling.Mesh, error) {
	return t.mesh(RegionZ)
}

// The first line's baseline runs through the origin and each further line
// sits below it. Characters the font has no glyph for are left blank and
// named in the error, which still comes with the mesh.
func (t Text) mesh(extrude func(triangulation.Region, float64) modeling.Mesh) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	if t.Height <= 0 {
		return empty, fmt.Errorf("height must be positive, got %g", t.Height)
	}
	name := t.Font
	if name == "" {
		name = "sans"
	}
	f, err := loadFont(name)
	if err != nil {
		return empty, err
	}

	// One em in 26.6 units of font units, so every coordinate is a font
	// unit times 64.
	unitsPerEm := int(f.FUnitsPerEm())
	em := fixed.I(unitsPerEm)
	var glyph truetype.GlyphBuf
	if err := glyph.Load(f, em, f.Index('H'), font.HintingNone); err != nil {
		return empty, err
	}
	capHeight := float64(glyph.Bounds.Max.Y) / 64
	if capHeight <= 0 {
		capHeight = 0.7 * float64(unitsPerEm)
	}
	scale := t.Height / capHeight
	lineHeight := 1.2 * float64(unitsPerEm)
	segments := max(t.CurveSegments, 1)

	var contours []geometry.Shape
	var missing []string
	for line, text := range strings.Split(strings.ReplaceAll(t.Text, "\r", ""), "\n") {
		var placed []placedGlyph
		pen := 0.
		for _, r := range text {
			index := f.Index(r)
			if index == 0 {
				missing = append(missing, string(r))
				pen += float64(f.HMetric(em, index).AdvanceWidth) / 64
				continue
			}
			if len(placed) > 0 {
				pen += float64(f.Kern(em, placed[len(placed)-1].index, index)) / 64
			}
			placed = append(placed, placedGlyph{index: index, x: pen})
			pen += float64(f.HMetric(em, index).AdvanceWidth) / 64
		}

		start := 0.
		switch t.Align {
		case "", "left":
		case "center":
			start = -pen / 2
		case "right":
			start = -pen
		default:
			return empty, fmt.Errorf("unknown align %q", t.Align)
		}

		baseline := -float64(line) * lineHeight
		for _, p := range placed {
			if err := glyph.Load(f, em, p.index, font.HintingNone); err != nil {
				return empty, err
			}
			origin := vector2.New(start+p.x, baseline)
			first := 0
			for _, end := range glyph.Ends {
				contours = append(contours, glyphContour(glyph.Points[first:end], segments, func(p truetype.Point) vector2.Float64 {
					return origin.Add(vector2.New(float64(p.X)/64, float64(p.Y)/64)).Scale(scale)
				}))
				first = end
			}
		}
	}

	var missingErr error
	if len(missing) > 0 {
		missingErr = fmt.Errorf("font %q has no glyph for %s, left blank", name, strings.Join(missing, " "))
	}
	region, err := triangulation.Fill(contours...)
	if err != nil {
		return empty, err
	}
	return extrude(region, t.Depth), missingErr
}

func glyphContour(points []truetype.Point, segments int, at func(truetype.Point) vector2.Float64) []vector2.Float64 {
	type anchor struct {
		p  vector2.Float64
		on bool
	}
	var expanded []anchor
	for i, p := range points {
		next := points[(i+1)%len(points)]
		expanded = append(expanded, anchor{at(p), p.Flags&1 != 0})
		if p.Flags&1 == 0 && next.Flags&1 == 0 {
			expanded = append(expanded, anchor{at(p).Add(at(next)).Scale(0.5), true})
		}
	}
	start := slices.IndexFunc(expanded, func(a anchor) bool { return a.on })
	if start == -1 {
		return nil
	}

	contour := []vector2.Float64{expanded[start].p}
	for k := 1; k < len(expanded); k++ {
		current := expanded[(start+k)%len(expanded)]
		if current.on {
			contour = append(contour, current.p)
			continue
		}
		from, to := contour[len(contour)-1], expanded[(start+k+1)%len(expanded)].p
		for j := 1; j <= segments; j++ {
			s := float64(j) / float64(segments)
			contour = append(contour, from.Scale((1-s)*(1-s)).Add(current.p.Scale(2*(1-s)*s)).Add(to.Scale(s*s)))
		}
		k++
	}
	return contour
}

type TextNode struct {
	Text          nodes.LiftedPort[string]  `description:"Text to render."`
	Height        nodes.LiftedPort[float64] `description:"Height of a capital letter. Defaults to 1."`
	Depth         nodes.LiftedPort[float64] `description:"Thickness: the solid spans 0 to Depth along the output's axis; negative goes the other way. 0 (the default) gives a single face."`
	Font          nodes.Output[string]      `description:"sans (the default), sans bold, mono or mono bold."`
	Align         nodes.LiftedPort[string]  `description:"Which end of each line sits on the origin: left (the default), center or right."`
	CurveSegments nodes.Output[int]         `description:"Straight segments per letter curve. Defaults to 6."`
}

func (TextNode) Description() string {
	return "Text as a mesh in a built-in font with the first baseline through the origin."
}

func (TextNode) Keywords() []string {
	return []string{"lettering", "emboss"}
}

func (n TextNode) X(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, Text.MeshX)
}

func (TextNode) XDescription() string {
	return "The text in the YZ plane facing +X: lines along Y, letters standing up along Z."
}

func (n TextNode) Y(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, Text.MeshY)
}

func (TextNode) YDescription() string {
	return "The text in the XZ plane facing +Y: lines along X, the tops of the letters toward -Z."
}

func (n TextNode) Z(out *nodes.Lifted[modeling.Mesh]) {
	n.out(out, Text.MeshZ)
}

func (TextNode) ZDescription() string {
	return "The text in the XY plane facing +Z: lines along X, letters standing up along Y."
}

func (n TextNode) out(out *nodes.Lifted[modeling.Mesh], mesh func(Text) (modeling.Mesh, error)) {
	font := nodes.TryGetOutputValue(out, n.Font, "")
	segments := nodes.TryGetOutputValue(out, n.CurveSegments, 6)
	nodes.Zip4(
		out,
		n.Text,
		nodes.LiftedOr(n.Height, 1.),
		n.Depth,
		n.Align,
		func(text string, height, depth float64, align string) modeling.Mesh {
			m, err := mesh(Text{
				Text:          text,
				Height:        height,
				Depth:         depth,
				Font:          font,
				Align:         align,
				CurveSegments: segments,
			})
			out.CaptureError(err)
			return m
		},
	)
}
