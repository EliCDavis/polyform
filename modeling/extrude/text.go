package extrude

import (
	"fmt"
	"math"
	"slices"
	"strings"
	"sync"

	"github.com/EliCDavis/polyform/generator/enum"
	"github.com/EliCDavis/polyform/generator/parameter"
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/primitives"
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

type Font int

const (
	Sans Font = iota
	SansBold
	Mono
	MonoBold
)

var fonts = parameter.DefineEnum(
	enum.Value(Sans, "Sans", "Go Regular"),
	enum.Value(SansBold, "Sans Bold", "Go Bold"),
	enum.Value(Mono, "Mono", "Go Mono"),
	enum.Value(MonoBold, "Mono Bold", "Go Mono Bold"),
)

var fontData = map[Font][]byte{
	Sans:     goregular.TTF,
	SansBold: gobold.TTF,
	Mono:     gomono.TTF,
	MonoBold: gomonobold.TTF,
}

type Align int

const (
	AlignLeft Align = iota
	AlignCenter
	AlignRight
)

var _ = parameter.DefineEnum(
	enum.Value(AlignLeft, "Left", "The line starts at the origin."),
	enum.Value(AlignCenter, "Center", "The line is centred on the origin."),
	enum.Value(AlignRight, "Right", "The line ends at the origin."),
)

var parsedFonts sync.Map

func loadFont(name Font) (*truetype.Font, error) {
	if f, ok := parsedFonts.Load(name); ok {
		return f.(*truetype.Font), nil
	}
	data, ok := fontData[name]
	if !ok {
		return nil, fmt.Errorf("unknown font %d", name)
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

	Font Font

	// Which end of each line sits on the origin.
	Align Align

	// Straight segments each curve of a letter's outline is drawn with.
	CurveSegments int

	UVs *primitives.StripUVs
}

type placedGlyph struct {
	index truetype.Index
	x     float64
}

// The text lies in the YZ plane facing +X, its lines running along Y.
func (t Text) MeshX() (modeling.Mesh, error) {
	return t.mesh(func(region triangulation.Region, depth float64) modeling.Mesh { return RegionX(region, depth, t.UVs) })
}

// The text lies in the XZ plane facing +Y, its lines running along X and the
// tops of its letters toward -Z.
func (t Text) MeshY() (modeling.Mesh, error) {
	return t.mesh(func(region triangulation.Region, depth float64) modeling.Mesh {
		return extrudeRegion(region, depth, func(v vector3.Float64) vector3.Float64 {
			return vector3.New(v.X(), v.Z(), -v.Y())
		}, t.UVs)
	})
}

// The text lies in the XY plane facing +Z, its lines running along X.
func (t Text) MeshZ() (modeling.Mesh, error) {
	return t.mesh(func(region triangulation.Region, depth float64) modeling.Mesh { return RegionZ(region, depth, t.UVs) })
}

type textLayout struct {
	font    *truetype.Font
	em      fixed.Int26_6
	scale   float64
	lines   [][]placedGlyph
	missing []string
}

// Glyph x positions are in font units, with each line's alignment applied.
func (t Text) layout() (textLayout, error) {
	if t.Height <= 0 {
		return textLayout{}, fmt.Errorf("height must be positive, got %g", t.Height)
	}
	f, err := loadFont(t.Font)
	if err != nil {
		return textLayout{}, err
	}

	// One em in 26.6 units of font units, so every coordinate is a font
	// unit times 64.
	em := fixed.I(int(f.FUnitsPerEm()))
	var glyph truetype.GlyphBuf
	if err := glyph.Load(f, em, f.Index('H'), font.HintingNone); err != nil {
		return textLayout{}, err
	}
	capHeight := float64(glyph.Bounds.Max.Y) / 64
	if capHeight <= 0 {
		capHeight = 0.7 * float64(f.FUnitsPerEm())
	}

	out := textLayout{font: f, em: em, scale: t.Height / capHeight}
	for _, text := range strings.Split(strings.ReplaceAll(t.Text, "\r", ""), "\n") {
		var placed []placedGlyph
		pen := 0.
		for _, r := range text {
			index := f.Index(r)
			if index == 0 {
				out.missing = append(out.missing, string(r))
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
		case AlignCenter:
			start = -pen / 2
		case AlignRight:
			start = -pen
		}
		for i := range placed {
			placed[i].x += start
		}
		out.lines = append(out.lines, placed)
	}
	return out, nil
}

// Width spans the ink of every line, not their advance widths.
func (t Text) Measure() (width float64, lines int, err error) {
	l, err := t.layout()
	if err != nil {
		return 0, 0, err
	}
	segments := max(t.CurveSegments, 1)
	var glyph truetype.GlyphBuf
	left, right := math.Inf(1), math.Inf(-1)
	for _, line := range l.lines {
		for _, p := range line {
			if err := glyph.Load(l.font, l.em, p.index, font.HintingNone); err != nil {
				return 0, 0, err
			}
			first := 0
			for _, end := range glyph.Ends {
				for _, v := range glyphContour(glyph.Points[first:end], segments, func(pt truetype.Point) vector2.Float64 {
					return vector2.New(p.x+float64(pt.X)/64, 0)
				}) {
					left, right = min(left, v.X()), max(right, v.X())
				}
				first = end
			}
		}
	}
	if right > left {
		width = (right - left) * l.scale
	}
	return width, len(l.lines), nil
}

// The first line's baseline runs through the origin and each further line
// sits below it. Characters the font has no glyph for are left blank and
// named in the error, which still comes with the mesh.
func (t Text) mesh(extrude func(triangulation.Region, float64) modeling.Mesh) (modeling.Mesh, error) {
	empty := modeling.EmptyMesh(modeling.TriangleTopology)
	l, err := t.layout()
	if err != nil {
		return empty, err
	}
	lineHeight := 1.2 * float64(l.font.FUnitsPerEm())
	segments := max(t.CurveSegments, 1)

	var glyph truetype.GlyphBuf
	var contours []geometry.Shape
	for line, placed := range l.lines {
		baseline := -float64(line) * lineHeight
		for _, p := range placed {
			if err := glyph.Load(l.font, l.em, p.index, font.HintingNone); err != nil {
				return empty, err
			}
			origin := vector2.New(p.x, baseline)
			first := 0
			for _, end := range glyph.Ends {
				contours = append(contours, glyphContour(glyph.Points[first:end], segments, func(p truetype.Point) vector2.Float64 {
					return origin.Add(vector2.New(float64(p.X)/64, float64(p.Y)/64)).Scale(l.scale)
				}))
				first = end
			}
		}
	}

	var missingErr error
	if len(l.missing) > 0 {
		missingErr = fmt.Errorf("font %s has no glyph for %s, left blank", fonts.Name(t.Font), strings.Join(l.missing, " "))
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
	Text          nodes.LiftedPort[string]
	Height        nodes.LiftedPort[float64]         `description:"Height of a capital letter." default:"1"`
	Depth         nodes.LiftedPort[float64]         `description:"Spans 0 to Depth along the output's axis; 0 gives one face."`
	Font          nodes.Output[Font]                `description:"Typeface the letters are cut from." default:"Sans"`
	Align         nodes.LiftedPort[Align]           `description:"Which end of each line sits on the origin."`
	CurveSegments nodes.Output[int]                 `description:"Straight segments per letter curve." default:"6"`
	UVs           nodes.Output[primitives.StripUVs] `description:"Walls as strips in world units; caps use the text's X and Y."`
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

func (n TextNode) Width(out *nodes.Lifted[float64]) {
	font := nodes.GetOutputValue(out, n.Font)
	segments := nodes.GetOutputValue(out, n.CurveSegments)
	nodes.Zip2(out, n.Text, n.Height, func(text string, height float64) float64 {
		width, _, err := Text{Text: text, Height: height, Font: font, CurveSegments: segments}.Measure()
		out.CaptureError(err)
		return width
	})
}

func (TextNode) WidthDescription() string {
	return "Width of the text's geometry."
}

func (n TextNode) Lines(out *nodes.Lifted[int]) {
	nodes.Zip1(out, n.Text, func(text string) int {
		return strings.Count(strings.ReplaceAll(text, "\r", ""), "\n") + 1
	})
}

func (n TextNode) out(out *nodes.Lifted[modeling.Mesh], mesh func(Text) (modeling.Mesh, error)) {
	font := nodes.GetOutputValue(out, n.Font)
	segments := nodes.GetOutputValue(out, n.CurveSegments)
	strip := nodes.TryGetOutputValue(out, n.UVs, primitives.UnitStripUVs())
	nodes.Zip4(
		out,
		n.Text,
		n.Height,
		n.Depth,
		n.Align,
		func(text string, height, depth float64, align Align) modeling.Mesh {
			m, err := mesh(Text{
				Text:          text,
				Height:        height,
				Depth:         depth,
				Font:          font,
				Align:         align,
				CurveSegments: segments,
				UVs:           &strip,
			})
			out.CaptureError(err)
			return m
		},
	)
}
