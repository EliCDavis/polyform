package sh

import (
	"fmt"
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector3"
)

const MaxDegree = 3

const C0 = 0.28209479177387814
const C1 = 0.4886025119029199

var C2 = [5]float64{1.0925484305920792, -1.0925484305920792, 0.31539156525252005, -1.0925484305920792, 0.5462742152960396}
var C3 = [7]float64{-0.5900435899266435, 2.890611442640554, -0.4570457994644658, 0.3731763325901154, -0.4570457994644658, 1.445305721320277, -0.5900435899266435}

func Band(degree int, dir vector3.Float64) []float64 {
	x, y, z := dir.X(), dir.Y(), dir.Z()
	switch degree {
	case 0:
		return []float64{C0}

	case 1:
		return []float64{-C1 * y, C1 * z, -C1 * x}

	case 2:
		xx, yy, zz := x*x, y*y, z*z
		return []float64{
			C2[0] * x * y,
			C2[1] * y * z,
			C2[2] * (2*zz - xx - yy),
			C2[3] * x * z,
			C2[4] * (xx - yy),
		}

	case 3:
		xx, yy, zz := x*x, y*y, z*z
		return []float64{
			C3[0] * y * (3*xx - yy),
			C3[1] * x * y * z,
			C3[2] * y * (4*zz - xx - yy),
			C3[3] * z * (2*zz - 3*xx - 3*yy),
			C3[4] * x * (4*zz - xx - yy),
			C3[5] * z * (xx - yy),
			C3[6] * x * (xx - 3*yy),
		}
	}

	panic(fmt.Errorf("unsupported spherical harmonic degree %d", degree))
}

func BandStart(degree int) int {
	return degree*degree - 1
}

func BandSize(degree int) int {
	return 2*degree + 1
}

var sampleDirections = []vector3.Float64{
	vector3.New(1., 0.3, -0.2).Normalized(),
	vector3.New(0.2, 1., 0.4).Normalized(),
	vector3.New(-0.3, 0.1, 1.).Normalized(),
	vector3.New(0.7, 0.6, 0.5).Normalized(),
	vector3.New(-0.5, 0.8, -0.4).Normalized(),
	vector3.New(0.6, -0.7, 0.3).Normalized(),
	vector3.New(0.2, -0.3, -0.9).Normalized(),
}

type Rotation struct {
	bands [MaxDegree][][]float64
}

func NewRotation(q quaternion.Quaternion) Rotation {
	q = q.Normalize()
	inverse := quaternion.New(q.Dir().Scale(-1), q.W())

	var r Rotation
	for degree := 1; degree <= MaxDegree; degree++ {
		n := BandSize(degree)
		basis := newMatrix(n)
		rotatedBasis := newMatrix(n)
		for k, dir := range sampleDirections[:n] {
			at := Band(degree, dir)
			rotatedAt := Band(degree, inverse.Rotate(dir))
			for m := 0; m < n; m++ {
				basis[m][k] = at[m]
				rotatedBasis[m][k] = rotatedAt[m]
			}
		}
		r.bands[degree-1] = transpose(multiply(rotatedBasis, invert(basis)))
	}
	return r
}

func (r Rotation) Band(degree int) [][]float64 {
	return r.bands[degree-1]
}

// Rotates in place; a band missing its last coefficient is left as is.
func (r Rotation) Rotate(coefficients []vector3.Float64) {
	var scratch [2*MaxDegree + 1]vector3.Float64
	for degree := 1; degree <= MaxDegree; degree++ {
		start, n := BandStart(degree), BandSize(degree)
		if start+n > len(coefficients) {
			return
		}
		copy(scratch[:n], coefficients[start:start+n])
		m := r.bands[degree-1]
		for i := 0; i < n; i++ {
			sum := vector3.Zero[float64]()
			for j := 0; j < n; j++ {
				sum = sum.Add(scratch[j].Scale(m[i][j]))
			}
			coefficients[start+i] = sum
		}
	}
}

func newMatrix(n int) [][]float64 {
	m := make([][]float64, n)
	for i := range m {
		m[i] = make([]float64, n)
	}
	return m
}

func transpose(a [][]float64) [][]float64 {
	n := len(a)
	t := newMatrix(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			t[j][i] = a[i][j]
		}
	}
	return t
}

func multiply(a, b [][]float64) [][]float64 {
	n := len(a)
	c := newMatrix(n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for k := 0; k < n; k++ {
				c[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return c
}

func invert(a [][]float64) [][]float64 {
	n := len(a)
	work := newMatrix(n)
	inv := newMatrix(n)
	for i := 0; i < n; i++ {
		copy(work[i], a[i])
		inv[i][i] = 1
	}

	for col := 0; col < n; col++ {
		pivot := col
		for row := col + 1; row < n; row++ {
			if math.Abs(work[row][col]) > math.Abs(work[pivot][col]) {
				pivot = row
			}
		}
		work[col], work[pivot] = work[pivot], work[col]
		inv[col], inv[pivot] = inv[pivot], inv[col]

		p := work[col][col]
		if p == 0 {
			panic("spherical harmonic sample directions are degenerate")
		}
		for j := 0; j < n; j++ {
			work[col][j] /= p
			inv[col][j] /= p
		}

		for row := 0; row < n; row++ {
			if row == col {
				continue
			}
			f := work[row][col]
			if f == 0 {
				continue
			}
			for j := 0; j < n; j++ {
				work[row][j] -= f * work[col][j]
				inv[row][j] -= f * inv[col][j]
			}
		}
	}
	return inv
}
