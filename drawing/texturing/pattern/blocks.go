package pattern

import (
	"math"

	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/math/norm"
	"github.com/EliCDavis/polyform/math/scalar"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
)

type BlocksNode struct {
	UV                 nodes.Output[uvTexture]     `description:"Sets the output's size."`
	Rows               nodes.Output[int]           `default:"8"`
	Units              nodes.Output[int]           `description:"Joint slots along each row." default:"12"`
	JointChance        nodes.Output[norm.UFloat64] `description:"Chance of a joint at each slot." default:"0.5"`
	MaxUnits           nodes.Output[int]           `description:"Longest block, in units." default:"4"`
	RowHeightVariation nodes.Output[norm.UFloat64]
	RowSlide           nodes.Output[norm.UFloat64] `description:"Random slide of each row, as a share of the row." default:"1"`
	Round              nodes.Output[float64]       `description:"Corner radius in pixels."`
	Seed               nodes.Output[int]
}

func (n BlocksNode) Description() string {
	return "Courses of blocks of varying length that tile: a random value per block, the UV within it, and the distance in pixels to its edge."
}

type blockLayout struct {
	rows, units int
	heights     []float64
	slides      []float64
	starts      [][]int
	lengths     [][]int
	seed        int
	round       float64
}

type blockSample struct {
	row, start   int
	local        vector2.Float64
	edgeDistance float64
}

func salted(seed, salt int) int {
	return int(hash01(seed, salt, 0x51ed) * (1 << 31))
}

func newBlockLayout(rows, units, maxUnits int, jointChance, heightVariation, slide, round float64, seed int) blockLayout {
	rows, units, maxUnits = max(rows, 1), max(units, 1), max(maxUnits, 1)
	l := blockLayout{
		rows:    rows,
		units:   units,
		heights: make([]float64, rows),
		slides:  make([]float64, rows),
		starts:  make([][]int, rows),
		lengths: make([][]int, rows),
		seed:    seed,
		round:   math.Max(round, 0),
	}
	jointSeed, heightSeed, slideSeed := salted(seed, 1), salted(seed, 2), salted(seed, 3)
	for r := range rows {
		l.heights[r] = (hash01(r, 0, heightSeed) - 0.5) * 0.9 * heightVariation
		l.slides[r] = hash01(r, 0, slideSeed) * slide * float64(units)

		starts, lengths := make([]int, units), make([]int, units)
		start := 0
		for k := 1; k <= units; k++ {
			if k < units && k-start < maxUnits && hash01(k, r, jointSeed) >= jointChance {
				continue
			}
			for i := start; i < k; i++ {
				starts[i], lengths[i] = start, k-start
			}
			start = k
		}
		l.starts[r], l.lengths[r] = starts, lengths
	}
	return l
}

func (l blockLayout) boundary(j int) float64 {
	return float64(j) + l.heights[scalar.Modulo(j, l.rows)]
}

func (l blockLayout) sample(uv vector2.Float64, width, height float64) blockSample {
	vt := uv.Y() * float64(l.rows)
	j := int(math.Floor(vt))
	switch {
	case vt < l.boundary(j):
		j--
	case vt >= l.boundary(j+1):
		j++
	}
	bottom, top := l.boundary(j), l.boundary(j+1)
	row := scalar.Modulo(j, l.rows)

	x := uv.X()*float64(l.units) + l.slides[row]
	k := int(math.Floor(x))
	unit := scalar.Modulo(k, l.units)
	start, length := l.starts[row][unit], l.lengths[row][unit]
	left := float64(k - (unit - start))

	pxPerUnit := width / float64(l.units)
	pxPerRow := height / float64(l.rows)
	halfW := float64(length) * pxPerUnit / 2
	halfH := (top - bottom) * pxPerRow / 2
	px := (x-left)*pxPerUnit - halfW
	py := (vt-bottom)*pxPerRow - halfH

	r := math.Min(l.round, math.Min(halfW, halfH))
	qx, qy := math.Abs(px)-(halfW-r), math.Abs(py)-(halfH-r)
	outside := math.Hypot(math.Max(qx, 0), math.Max(qy, 0))
	inside := math.Min(math.Max(qx, qy), 0)

	return blockSample{
		row:          row,
		start:        start,
		local:        vector2.New((x-left)/float64(length), (vt-bottom)/(top-bottom)),
		edgeDistance: r - outside - inside,
	}
}

func (n BlocksNode) blocks(out nodes.ExecutionRecorder, f func(s blockSample, l blockLayout) float64) texturing.Texture[float64] {
	uv := nodes.GetOutputValue(out, n.UV)
	l := n.layout(out)
	width, height := float64(uv.Width()), float64(uv.Height())
	return texturing.Map(uv, func(uv vector2.Float64) float64 {
		return f(l.sample(uv, width, height), l)
	})
}

func (n BlocksNode) layout(out nodes.ExecutionRecorder) blockLayout {
	return newBlockLayout(
		nodes.GetOutputValue(out, n.Rows),
		nodes.GetOutputValue(out, n.Units),
		nodes.GetOutputValue(out, n.MaxUnits),
		nodes.GetOutputValue(out, n.JointChance).Value(),
		nodes.GetOutputValue(out, n.RowHeightVariation).Value(),
		nodes.GetOutputValue(out, n.RowSlide).Value(),
		nodes.GetOutputValue(out, n.Round),
		nodes.GetOutputValue(out, n.Seed),
	)
}

func requireUV(out nodes.ExecutionRecorder, uv nodes.OutputPort) bool {
	if uv == nil {
		out.CaptureError(errNoUV)
		return false
	}
	return true
}

func (n BlocksNode) Random(out *nodes.StructOutput[texturing.Texture[float64]]) {
	if !requireUV(out, n.UV) {
		return
	}
	out.Set(n.blocks(out, func(s blockSample, l blockLayout) float64 {
		return hash01(s.start, s.row, l.seed)
	}))
}

func (n BlocksNode) EdgeDistance(out *nodes.StructOutput[texturing.Texture[float64]]) {
	if !requireUV(out, n.UV) {
		return
	}
	out.Set(n.blocks(out, func(s blockSample, l blockLayout) float64 { return s.edgeDistance }))
}

func (n BlocksNode) LocalUV(out *nodes.StructOutput[uvTexture]) {
	if !requireUV(out, n.UV) {
		return
	}
	uv := nodes.GetOutputValue(out, n.UV)
	l := n.layout(out)
	width, height := float64(uv.Width()), float64(uv.Height())
	out.Set(texturing.Map(uv, func(uv vector2.Float64) vector2.Float64 {
		return l.sample(uv, width, height).local
	}))
}

// ============================================================================

type RehashNode struct {
	Texture nodes.LiftedPort[texturing.Texture[float64]]
	Seed    nodes.LiftedPort[int]
}

func (n RehashNode) Description() string {
	return "A random value from 0 to 1 per distinct input value."
}

func (n RehashNode) Value(out *nodes.Lifted[texturing.Texture[float64]]) {
	if n.Texture == nil {
		return
	}
	nodes.Zip2(out, n.Texture, n.Seed, func(in texturing.Texture[float64], seed int) texturing.Texture[float64] {
		return texturing.Map(in, func(v float64) float64 {
			bits := math.Float64bits(v)
			return hash01(int(bits>>32), int(bits&0xffffffff), seed)
		})
	})
}
