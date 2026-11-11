package pattern

import (
	"errors"
	"math"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

func floorMod(a, b int) int {
	return ((a % b) + b) % b
}

func fract(v float64) float64 {
	return v - math.Floor(v)
}

// Wraps by the whole count so cells keep their identity across a tiled UV seam.
func cellIndex(coordinate, count float64) int {
	i := int(math.Floor(coordinate * count))
	if whole := int(math.Round(count)); whole >= 1 {
		return floorMod(i, whole)
	}
	return i
}

func hash01(a, b, seed int) float64 {
	h := uint64(int64(a))*0x9E3779B97F4A7C15 ^ uint64(int64(b))*0xC2B2AE3D27D4EB4F ^ uint64(int64(seed))*0x165667B19E3779F9
	h ^= h >> 33
	h *= 0xff51afd7ed558ccd
	h ^= h >> 33
	h *= 0xc4ceb9fe1a85ec53
	h ^= h >> 33
	return float64(h>>11) / float64(uint64(1)<<53)
}

func smoothstep(edge0, edge1, x float64) float64 {
	if edge1 <= edge0 {
		if x < edge0 {
			return 0
		}
		return 1
	}
	t := math.Min(math.Max((x-edge0)/(edge1-edge0), 0), 1)
	return t * t * (3 - 2*t)
}

func stripe(coordinate, count, width, softness float64) float64 {
	f := fract(coordinate * count)
	half := width / 2
	dist := math.Abs(f - half)
	dist = math.Min(dist, 1-dist)
	s := softness / 2
	return 1 - smoothstep(half-s, half+s, dist)
}

func triangleWave(t float64) float64 {
	return 1 - math.Abs(2*fract(t)-1)
}

type uvTexture = texturing.Texture[vector2.Float64]

var errNoUV = errors.New("no UV texture is connected; wire a UV node, or a UV warp, into it")

func fromUV[R any](out *nodes.StructOutput[texturing.Texture[R]], uv nodes.Output[uvTexture], f func(vector2.Float64) R) {
	if uv == nil {
		out.CaptureError(errNoUV)
		return
	}
	out.Set(texturing.Map(nodes.GetOutputValue(out, uv), f))
}

// ============================================================================

type CellNode struct {
	UV        nodes.Output[uvTexture] `description:"Sets the output's size."`
	Columns   nodes.Output[float64]   `default:"8"`
	Rows      nodes.Output[float64]   `default:"8"`
	RowOffset nodes.Output[float64]   `description:"Share of a cell every other row slides along U. 0.5 is a running bond."`
	Seed      nodes.Output[int]
}

func (n CellNode) Description() string {
	return "A grid of cells: each pixel's cell index, its UV within the cell, and a random value per cell."
}

type cell struct {
	column, row int
	local       vector2.Float64
}

func (n CellNode) cells(out nodes.ExecutionRecorder) func(uv vector2.Float64) cell {
	columns := nodes.GetOutputValue(out, n.Columns)
	rows := nodes.GetOutputValue(out, n.Rows)
	offset := nodes.GetOutputValue(out, n.RowOffset)
	return func(uv vector2.Float64) cell {
		row := cellIndex(uv.Y(), rows)
		u := uv.X()
		if floorMod(row, 2) == 1 {
			u += offset / columns
		}
		return cell{
			column: cellIndex(u, columns),
			row:    row,
			local:  vector2.New(fract(u*columns), fract(uv.Y()*rows)),
		}
	}
}

func (n CellNode) Column(out *nodes.StructOutput[texturing.Texture[float64]]) {
	at := n.cells(out)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 { return float64(at(uv).column) })
}

func (n CellNode) Row(out *nodes.StructOutput[texturing.Texture[float64]]) {
	at := n.cells(out)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 { return float64(at(uv).row) })
}

func (n CellNode) LocalUV(out *nodes.StructOutput[uvTexture]) {
	at := n.cells(out)
	fromUV(out, n.UV, func(uv vector2.Float64) vector2.Float64 { return at(uv).local })
}

func (n CellNode) Random(out *nodes.StructOutput[texturing.Texture[float64]]) {
	at := n.cells(out)
	seed := nodes.GetOutputValue(out, n.Seed)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		c := at(uv)
		return hash01(c.column, c.row, seed)
	})
}

// ============================================================================

type StripesNode struct {
	UV       nodes.Output[uvTexture] `description:"Sets the output's size."`
	Count    nodes.Output[float64]   `default:"8"`
	Width    nodes.Output[float64]   `description:"Share of each period covered, 0 to 1." default:"0.5"`
	Softness nodes.Output[float64]   `description:"Share of each period the edge blends over."`
	Angle    nodes.Output[float64]   `description:"Degrees from running along V."`
}

func (n StripesNode) Description() string {
	return "Repeating stripes: 1 on a stripe, 0 between them."
}

func (n StripesNode) Value(out *nodes.StructOutput[texturing.Texture[float64]]) {
	count := nodes.GetOutputValue(out, n.Count)
	width := nodes.GetOutputValue(out, n.Width)
	softness := nodes.GetOutputValue(out, n.Softness)
	angle := nodes.GetOutputValue(out, n.Angle) * math.Pi / 180
	cos, sin := math.Cos(angle), math.Sin(angle)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		return stripe(uv.X()*cos+uv.Y()*sin, count, width, softness)
	})
}

// ============================================================================

type CheckerNode struct {
	UV      nodes.Output[uvTexture] `description:"Sets the output's size."`
	Columns nodes.Output[float64]   `default:"8"`
	Rows    nodes.Output[float64]   `default:"8"`
}

func (n CheckerNode) Description() string {
	return "A checkerboard: alternating cells of 0 and 1."
}

func (n CheckerNode) Value(out *nodes.StructOutput[texturing.Texture[float64]]) {
	columns := nodes.GetOutputValue(out, n.Columns)
	rows := nodes.GetOutputValue(out, n.Rows)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		return float64(floorMod(cellIndex(uv.X(), columns)+cellIndex(uv.Y(), rows), 2))
	})
}

// ============================================================================

type CrossedStripesNode struct {
	UV       nodes.Output[uvTexture] `description:"Sets the output's size."`
	Columns  nodes.Output[float64]   `default:"8"`
	Rows     nodes.Output[float64]   `default:"8"`
	Width    nodes.Output[float64]   `description:"Share of each period covered, 0 to 1." default:"0.5"`
	Softness nodes.Output[float64]   `description:"Share of each period the edge blends over."`
}

func (n CrossedStripesNode) Description() string {
	return "Stripes along U crossed with stripes along V: 0 under none, 0.5 under one, 1 where two cross."
}

func (n CrossedStripesNode) Keywords() []string {
	return []string{"gingham", "plaid"}
}

func (n CrossedStripesNode) Value(out *nodes.StructOutput[texturing.Texture[float64]]) {
	columns := nodes.GetOutputValue(out, n.Columns)
	rows := nodes.GetOutputValue(out, n.Rows)
	width := nodes.GetOutputValue(out, n.Width)
	softness := nodes.GetOutputValue(out, n.Softness)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		return (stripe(uv.X(), columns, width, softness) + stripe(uv.Y(), rows, width, softness)) / 2
	})
}

// ============================================================================

type WeaveNode struct {
	UV          nodes.Output[uvTexture] `description:"Sets the output's size."`
	Columns     nodes.Output[float64]   `description:"Strands along V." default:"16"`
	Rows        nodes.Output[float64]   `description:"Strands along U." default:"16"`
	Over        nodes.Output[int]       `description:"Cells a U strand passes over before going under." default:"1"`
	Under       nodes.Output[int]       `description:"Cells a U strand passes under." default:"1"`
	Herringbone nodes.Output[int]       `description:"Columns between reversals of the diagonal. 0 keeps one diagonal."`
	Gap         nodes.Output[float64]   `description:"Share of each cell left open, 0 to 1."`
	Seed        nodes.Output[int]
}

func (n WeaveNode) Description() string {
	return "Strands woven over and under: the height of the strand on top, a mask of which direction shows, and a random value per strand."
}

type weaveSample struct {
	height float64
	alongU bool
	strand int
}

type weave struct {
	columns, rows float64
	over, under   int
	herringbone   int
	gap           float64
}

func strandProfile(across, gap float64) float64 {
	if gap >= 1 {
		return 0
	}
	t := (across - gap/2) / (1 - gap)
	if t <= 0 || t >= 1 {
		return 0
	}
	return math.Sqrt(math.Sin(math.Pi * t))
}

func (w weave) sample(u, v float64) weaveSample {
	period := w.over + w.under
	column, row := cellIndex(u, w.columns), cellIndex(v, w.rows)
	localU, localV := fract(u*w.columns), fract(v*w.rows)

	step := 1
	if w.herringbone > 0 && floorMod(column/w.herringbone, 2) == 1 {
		step = -1
	}
	phase := floorMod(column+step*row, period)

	uOnTop := phase < w.over
	var along float64
	if uOnTop {
		along = (float64(phase) + localU) / float64(w.over)
	} else {
		k := phase - w.over
		if step < 0 {
			k = w.under - 1 - k
		}
		along = (float64(k) + localV) / float64(w.under)
	}
	arch := math.Sin(math.Pi * along)

	uProfile := strandProfile(localV, w.gap)
	vProfile := strandProfile(localU, w.gap)
	top, bottom := uProfile, vProfile
	if !uOnTop {
		top, bottom = vProfile, uProfile
	}

	switch {
	case top > 0:
		return weaveSample{height: 0.5 + 0.5*top*arch, alongU: uOnTop, strand: w.strand(uOnTop, column, row)}
	case bottom > 0:
		return weaveSample{height: 0.35 * bottom, alongU: !uOnTop, strand: w.strand(!uOnTop, column, row)}
	}
	return weaveSample{alongU: uOnTop, strand: -1}
}

func (w weave) strand(alongU bool, column, row int) int {
	if alongU {
		return row
	}
	return column
}

func (n WeaveNode) weave(out *nodes.StructOutput[texturing.Texture[float64]], f func(s weaveSample, seed int) float64) {
	w := weave{
		columns:     nodes.GetOutputValue(out, n.Columns),
		rows:        nodes.GetOutputValue(out, n.Rows),
		over:        max(nodes.GetOutputValue(out, n.Over), 1),
		under:       max(nodes.GetOutputValue(out, n.Under), 1),
		herringbone: nodes.GetOutputValue(out, n.Herringbone),
		gap:         math.Min(math.Max(nodes.GetOutputValue(out, n.Gap), 0), 1),
	}
	seed := nodes.GetOutputValue(out, n.Seed)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		return f(w.sample(uv.X(), uv.Y()), seed)
	})
}

func (n WeaveNode) Height(out *nodes.StructOutput[texturing.Texture[float64]]) {
	n.weave(out, func(s weaveSample, seed int) float64 { return s.height })
}

func strandMask(s weaveSample, alongU bool) float64 {
	if s.strand >= 0 && s.alongU == alongU {
		return 1
	}
	return 0
}

func (n WeaveNode) Weft(out *nodes.StructOutput[texturing.Texture[float64]]) {
	n.weave(out, func(s weaveSample, seed int) float64 { return strandMask(s, true) })
}

func (n WeaveNode) Warp(out *nodes.StructOutput[texturing.Texture[float64]]) {
	n.weave(out, func(s weaveSample, seed int) float64 { return strandMask(s, false) })
}

func (n WeaveNode) Random(out *nodes.StructOutput[texturing.Texture[float64]]) {
	n.weave(out, func(s weaveSample, seed int) float64 {
		if s.strand < 0 {
			return 0
		}
		direction := 0
		if s.alongU {
			direction = 1
		}
		return hash01(s.strand, direction, seed)
	})
}

// ============================================================================

type TransformUVNode struct {
	In       nodes.Output[uvTexture]
	Scale    nodes.Output[vector2.Float64] `description:"Repeats across the texture." default:"1,1"`
	Rotation nodes.Output[float64]         `description:"Degrees, around the centre."`
	Offset   nodes.Output[vector2.Float64] `description:"In UV units."`
}

func (n TransformUVNode) Description() string {
	return "Scales, rotates and slides UVs."
}

func (n TransformUVNode) UV(out *nodes.StructOutput[uvTexture]) {
	scale := nodes.GetOutputValue(out, n.Scale)
	angle := nodes.GetOutputValue(out, n.Rotation) * math.Pi / 180
	offset := nodes.GetOutputValue(out, n.Offset)
	cos, sin := math.Cos(angle), math.Sin(angle)
	centre := vector2.New(0.5, 0.5)
	fromUV(out, n.In, func(uv vector2.Float64) vector2.Float64 {
		p := uv.Sub(centre)
		p = vector2.New(p.X()*cos-p.Y()*sin, p.X()*sin+p.Y()*cos)
		return p.MultByVector(scale).Add(centre).Add(offset)
	})
}

// ============================================================================

type ZigzagUVNode struct {
	In        nodes.Output[uvTexture]
	Count     nodes.Output[float64] `description:"Zigzags across U. Whole numbers tile." default:"1"`
	Amplitude nodes.Output[float64] `description:"V shift at each peak, in UV units." default:"0.1"`
}

func (n ZigzagUVNode) Description() string {
	return "Shifts V in a zigzag along U, bending straight bands into chevrons."
}

func (n ZigzagUVNode) UV(out *nodes.StructOutput[uvTexture]) {
	count := nodes.GetOutputValue(out, n.Count)
	amplitude := nodes.GetOutputValue(out, n.Amplitude)
	fromUV(out, n.In, func(uv vector2.Float64) vector2.Float64 {
		return vector2.New(uv.X(), uv.Y()+amplitude*triangleWave(uv.X()*count))
	})
}

// ============================================================================

type DisplaceUVNode struct {
	In     nodes.Output[uvTexture]
	Amount nodes.Output[texturing.Texture[float64]] `description:"How far each pixel moves. Must match the UVs' size."`
	Shift  nodes.Output[vector2.Float64]            `description:"UV offset for an Amount of 1." default:"0.1,0.1"`
}

func (n DisplaceUVNode) Description() string {
	return "Moves each UV by Shift scaled by the matching pixel of Amount."
}

func (n DisplaceUVNode) UV(out *nodes.StructOutput[uvTexture]) {
	if n.In == nil {
		out.CaptureError(errNoUV)
		return
	}
	uv := nodes.GetOutputValue(out, n.In)
	if n.Amount == nil {
		out.Set(uv)
		return
	}
	amount := nodes.GetOutputValue(out, n.Amount)
	if amount.Width() != uv.Width() || amount.Height() != uv.Height() {
		out.CaptureError(texturing.ErrMismatchDimensions)
		return
	}
	shift := nodes.GetOutputValue(out, n.Shift)
	displaced := texturing.Empty[vector2.Float64](uv.Width(), uv.Height())
	displaced.MutateParallel(func(x, y int, _ vector2.Float64) vector2.Float64 {
		return uv.Get(x, y).Add(shift.Scale(amount.Get(x, y)))
	})
	out.Set(displaced)
}

// ============================================================================

type DistanceNode struct {
	UV     nodes.Output[uvTexture]       `description:"Sets the output's size."`
	Center nodes.Output[vector2.Float64] `description:"In UV units." default:"0.5,0.5"`
}

func (n DistanceNode) Description() string {
	return "Each pixel's UV distance from Center."
}

func (n DistanceNode) Value(out *nodes.StructOutput[texturing.Texture[float64]]) {
	center := nodes.GetOutputValue(out, n.Center)
	fromUV(out, n.UV, func(uv vector2.Float64) float64 {
		return uv.Distance(center)
	})
}
