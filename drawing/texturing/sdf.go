package texturing

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type SampleSDFNode struct {
	Texture nodes.Output[Texture[float64]]
	Size    nodes.Output[vector3.Float64]
}

func (n SampleSDFNode) SDF(out *nodes.StructOutput[sample.Vec3ToFloat]) {
	if n.Texture == nil {
		return
	}
	tex := nodes.GetOutputValue(out, n.Texture)
	size := nodes.TryGetOutputValue(out, n.Size, vector3.One[float64]())
	bounds := geometry.NewAABB(vector3.Zero[float64](), size)
	half := size.Scale(0.5)
	out.Set(func(f vector3.Float64) float64 {
		v := bounds.ClosestPoint(f)
		p := v.XZ().
			DivByVector(half.XZ()).
			Scale(0.5).
			Add(vector2.Fill(0.5))

		t := tex.UV(p.X(), p.Y())
		return t + f.Distance(v)
	})
}

type MaskToSDFNode struct {
	Mask nodes.Output[Texture[bool]]
}

func (n MaskToSDFNode) SDF(out *nodes.StructOutput[Texture[float64]]) {
	if n.Mask == nil {
		return
	}
	out.Set(ToSDF(nodes.GetOutputValue(out, n.Mask)))
}

func (n MaskToSDFNode) NormalizedSDF(out *nodes.StructOutput[Texture[float64]]) {
	if n.Mask == nil {
		return
	}

	sdf := ToSDF(nodes.GetOutputValue(out, n.Mask))

	maxV := -math.MaxFloat64
	for _, v := range sdf.data {
		maxV = max(maxV, v)
	}

	sdf.MutateParallel(func(x, y int, v float64) float64 {
		return v / maxV
	})

	out.Set(sdf)
}

func edt1D(input, result []float64, n int) {
	paras := make([]int, n)
	ranges := make([]float64, n+1)

	k := 0
	ranges[0] = math.Inf(-1)
	ranges[1] = math.Inf(1)

	for q := 1; q < n; q++ {
		q2 := float64(q * q)
		s := ((input[q] + q2) -
			(input[paras[k]] + float64(paras[k]*paras[k]))) /
			float64(2*(q-paras[k]))

		for s <= ranges[k] {
			k--
			s = ((input[q] + q2) -
				(input[paras[k]] + float64(paras[k]*paras[k]))) /
				float64(2*(q-paras[k]))
		}

		k++
		paras[k] = q
		ranges[k] = s
		ranges[k+1] = math.Inf(1)
	}

	k = 0
	for q := range n {
		for ranges[k+1] < float64(q) {
			k++
		}
		v := q - paras[k]
		result[q] = float64(v*v) + input[paras[k]]
	}
}

// Implementation of:
//
//	Distance Transforms of Sampled Functions
//	by Pedro F. Felzenszwalb and Daniel P. Huttenlocher
//	https://cs.brown.edu/people/pfelzens/papers/dt-final.pdf
func ToSDF(tex Texture[bool]) Texture[float64] {
	w, h := tex.width, tex.height
	n := w * h

	// Output
	sdf := Texture[float64]{
		width:  w,
		height: h,
		data:   make([]float64, n),
	}

	// --- Boundary detection ---
	isBoundary := func(x, y int) bool {
		if !tex.data[y*w+x] {
			return false
		}
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				if dx == 0 && dy == 0 {
					continue
				}
				nx, ny := x+dx, y+dy
				if nx < 0 || ny < 0 || nx >= w || ny >= h {
					return true
				}
				if !tex.data[ny*w+nx] {
					return true
				}
			}
		}
		return false
	}

	// Distance buffers
	inside := make([]float64, n)
	for y := range h {
		for x := range w {
			i := y*w + x
			if isBoundary(x, y) {
				inside[i] = 0
			} else {
				inside[i] = math.MaxFloat64
			}
		}
	}

	// --- 2D EDT ---
	edt2D := func(grid []float64) []float64 {
		tmp := make([]float64, n)
		out := make([]float64, n)

		// Rows
		for y := range h {
			row := grid[y*w : (y+1)*w]
			edt1D(row, tmp[y*w:], w)
		}

		// Columns
		result := make([]float64, max(w, h))
		col := make([]float64, h)
		for x := range w {
			for y := range h {
				col[y] = tmp[y*w+x]
			}
			edt1D(col, result, h)
			for y := range h {
				out[y*w+x] = math.Sqrt(result[y])
			}
		}

		return out
	}

	// Compute distances
	dist := edt2D(inside)

	// --- Combine into signed field ---
	for i := range n {
		if tex.data[i] {
			sdf.data[i] = -dist[i]
		} else {
			sdf.data[i] = dist[i]
		}
	}

	return sdf
}
