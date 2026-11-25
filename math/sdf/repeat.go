package sdf

import (
	"math"
	"sort"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

const (
	boundProbeDistance = 1e5
	boundProbeCoverage = 0.85
	boundProbeSlack    = 1e-6
	repeatLeafSize     = 4
)

var boundProbeDirections = func() []vector3.Float64 {
	directions := make([]vector3.Float64, 0, 26)
	for x := -1; x <= 1; x++ {
		for y := -1; y <= 1; y++ {
			for z := -1; z <= 1; z++ {
				if x == 0 && y == 0 && z == 0 {
					continue
				}
				directions = append(directions, vector3.New(float64(x), float64(y), float64(z)).Normalized())
			}
		}
	}
	return directions
}()

func boundingRadius(field sample.Vec3ToFloat) (float64, bool) {
	d := boundProbeDistance
	c := boundProbeCoverage
	uncovered := d * d * (1 - c*c)
	radius := 0.
	for _, dir := range boundProbeDirections {
		probe := field(dir.Scale(d))
		if !(probe > 0) || probe*probe <= uncovered {
			return 0, false
		}
		// A field growing faster than distance does would put the radius too small.
		if growth := field(dir.Scale(2*d)) - probe; !(growth <= d+boundProbeSlack) {
			return 0, false
		}
		radius = math.Max(radius, d*c-math.Sqrt(probe*probe-uncovered))
	}
	if radius > d/4 {
		return 0, false
	}
	return radius, true
}

func minMaxAbs(v vector3.Float64) (float64, float64) {
	x, y, z := math.Abs(v.X()), math.Abs(v.Y()), math.Abs(v.Z())
	return math.Min(x, math.Min(y, z)), math.Max(x, math.Max(y, z))
}

type repeatCopy struct {
	inverse  mat.Matrix4x4
	center   vector3.Float64
	factor   float64
	slope    float64
	inflated float64
}

func (c repeatCopy) eval(field sample.Vec3ToFloat, v vector3.Float64) float64 {
	return field(c.inverse.MulPosition(v)) * c.factor
}

type repeatNode struct {
	centers     geometry.AABB
	slope       float64
	inflated    float64
	left, right *repeatNode
	copies      []repeatCopy
}

func (n *repeatNode) lowerBound(v vector3.Float64) float64 {
	return n.slope*n.centers.ClosestPoint(v).Distance(v) - n.inflated
}

func buildRepeatTree(copies []repeatCopy) *repeatNode {
	node := &repeatNode{slope: math.Inf(1), inflated: math.Inf(-1)}
	centers := make([]vector3.Float64, len(copies))
	for i, c := range copies {
		centers[i] = c.center
		node.slope = math.Min(node.slope, c.slope)
		node.inflated = math.Max(node.inflated, c.inflated)
	}
	node.centers = geometry.NewAABBFromPoints(centers...)

	if len(copies) <= repeatLeafSize {
		node.copies = copies
		return node
	}

	size := node.centers.Size()
	component := vector3.Float64.X
	if size.Y() > size.X() && size.Y() >= size.Z() {
		component = vector3.Float64.Y
	} else if size.Z() > size.X() && size.Z() > size.Y() {
		component = vector3.Float64.Z
	}
	sort.Slice(copies, func(i, j int) bool {
		return component(copies[i].center) < component(copies[j].center)
	})
	half := len(copies) / 2
	node.left = buildRepeatTree(copies[:half])
	node.right = buildRepeatTree(copies[half:])
	return node
}

type repeatQuery struct {
	field      sample.Vec3ToFloat
	point      vector3.Float64
	radius     float64
	min1, min2 float64
}

func (q *repeatQuery) threshold() float64 {
	return math.Min(q.min2, q.min1+q.radius)
}

func (q *repeatQuery) visit(n *repeatNode) {
	if n.copies != nil {
		for _, c := range n.copies {
			if c.slope*c.center.Distance(q.point)-c.inflated >= q.threshold() {
				continue
			}
			val := c.eval(q.field, q.point)
			if val < q.min1 {
				q.min1, q.min2 = val, q.min1
			} else if val < q.min2 {
				q.min2 = val
			}
		}
		return
	}

	first, second := n.left, n.right
	firstBound, secondBound := first.lowerBound(q.point), second.lowerBound(q.point)
	if secondBound < firstBound {
		first, second = second, first
		firstBound, secondBound = secondBound, firstBound
	}
	if firstBound < q.threshold() {
		q.visit(first)
	}
	if secondBound < q.threshold() {
		q.visit(second)
	}
}

func Repeat(field sample.Vec3ToFloat, transforms []trs.TRS, radius float64) sample.Vec3ToFloat {
	copies := make([]repeatCopy, 0, len(transforms))
	for _, t := range transforms {
		smallest, largest := minMaxAbs(t.Scale())
		if smallest == 0 {
			continue
		}
		copies = append(copies, repeatCopy{
			inverse: t.Inverse(),
			center:  t.Position(),
			factor:  smallest,
			slope:   smallest / largest,
		})
	}

	if len(copies) == 0 {
		return nullField
	}

	if len(copies) == 1 {
		only := copies[0]
		return func(v vector3.Float64) float64 {
			return only.eval(field, v)
		}
	}

	radius = math.Max(radius, 0)
	bound, bounded := boundingRadius(field)
	if !bounded {
		return func(v vector3.Float64) float64 {
			min1, min2 := math.Inf(1), math.Inf(1)
			for _, c := range copies {
				val := c.eval(field, v)
				if val < min1 {
					min1, min2 = val, min1
				} else if val < min2 {
					min2 = val
				}
			}
			if radius == 0 {
				return min1
			}
			return smoothUnionBlend(min1, min2, radius)
		}
	}

	for i := range copies {
		copies[i].inflated = copies[i].factor * bound
	}
	root := buildRepeatTree(copies)

	return func(v vector3.Float64) float64 {
		q := repeatQuery{field: field, point: v, radius: radius, min1: math.Inf(1), min2: math.Inf(1)}
		q.visit(root)
		if radius == 0 {
			return q.min1
		}
		return smoothUnionBlend(q.min1, q.min2, radius)
	}
}

type RepeatNode struct {
	Transforms nodes.Output[[]trs.TRS]          `description:"One TRS per copy to place. Defaults to a single identity transform (one unmoved copy). An empty array produces an empty field (nothing)."`
	Field      nodes.Output[sample.Vec3ToFloat] `description:"The field to repeat. Nothing is set on the output until this is connected."`
	Radius     nodes.Output[float64]            `description:"Width of the blend region in world units. Zero or less is a regular union."`
}

func (cn RepeatNode) Description() string {
	return "Places one copy of an SDF field at each given transform and unions them together."
}

func (cn RepeatNode) Result(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if cn.Field == nil {
		return
	}

	out.Set(Repeat(
		nodes.GetOutputValue(out, cn.Field),
		nodes.TryGetOutputValue(out, cn.Transforms, []trs.TRS{trs.Identity()}),
		nodes.GetOutputValue(out, cn.Radius),
	))
}
