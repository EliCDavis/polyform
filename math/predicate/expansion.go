package predicate

// Shewchuk, "Adaptive Precision Floating-Point Arithmetic and Fast Robust
// Geometric Predicates", Discrete & Computational Geometry 18 (1997).
//
// A number too precise for one float is held as an expansion: a list of
// floats, each far smaller than the last, that sum to it exactly. The
// transforms below never round; every bit lost by an operation is caught
// in a second float. Products are wrapped in float64() so the compiler can
// not fuse them into a multiply-add, which would skip a rounding the error
// terms account for.

const splitter = 134217729.0 // 2^27 + 1

// Runtime float64 arithmetic, as Shewchuk computes them, so each bound
// rounds the way the derivation assumes.
var (
	epsilon        = 1.1102230246251565e-16 // 2^-53, half an ulp of 1
	resultErrBound = (3 + 8*epsilon) * epsilon
	ccwErrBoundA   = (3 + 16*epsilon) * epsilon
	ccwErrBoundB   = (2 + 12*epsilon) * epsilon
	ccwErrBoundC   = (9 + 64*epsilon) * epsilon * epsilon
	o3dErrBoundA   = (7 + 56*epsilon) * epsilon
	o3dErrBoundB   = (3 + 28*epsilon) * epsilon
	o3dErrBoundC   = (26 + 288*epsilon) * epsilon * epsilon
	iccErrBoundA   = (10 + 96*epsilon) * epsilon
	iccErrBoundB   = (4 + 48*epsilon) * epsilon
	iccErrBoundC   = (44 + 576*epsilon) * epsilon * epsilon
)

// Requires |a| >= |b|.
func fastTwoSum(a, b float64) (x, y float64) {
	x = a + b
	bvirt := x - a
	y = b - bvirt
	return x, y
}

func twoSum(a, b float64) (x, y float64) {
	x = a + b
	bvirt := x - a
	avirt := x - bvirt
	bround := b - bvirt
	around := a - avirt
	return x, around + bround
}

func twoDiff(a, b float64) (x, y float64) {
	x = a - b
	return x, twoDiffTail(a, b, x)
}

// The bits a - b = x lost.
func twoDiffTail(a, b, x float64) float64 {
	bvirt := a - x
	avirt := x + bvirt
	bround := bvirt - b
	around := a - avirt
	return around + bround
}

func split(a float64) (hi, lo float64) {
	c := float64(splitter * a)
	abig := c - a
	hi = c - abig
	lo = a - hi
	return hi, lo
}

func twoProduct(a, b float64) (x, y float64) {
	x = float64(a * b)
	ahi, alo := split(a)
	bhi, blo := split(b)
	return x, productTail(x, ahi, alo, bhi, blo)
}

// With b already split, for scaling a whole expansion by it.
func twoProductPresplit(a, b, bhi, blo float64) (x, y float64) {
	x = float64(a * b)
	ahi, alo := split(a)
	return x, productTail(x, ahi, alo, bhi, blo)
}

func productTail(x, ahi, alo, bhi, blo float64) float64 {
	err1 := x - float64(ahi*bhi)
	err2 := err1 - float64(alo*bhi)
	err3 := err2 - float64(ahi*blo)
	return float64(alo*blo) - err3
}

// (a1 + a0) - (b1 + b0) as a four term expansion, least significant first.
func twoTwoDiff(a1, a0, b1, b0 float64) (x3, x2, x1, x0 float64) {
	i, x0 := twoDiff(a0, b0)
	j, x1a := twoSum(a1, i)
	i, x1 = twoDiff(x1a, b1)
	x3, x2 = twoSum(j, i)
	return x3, x2, x1, x0
}

// Sums two expansions into h, dropping zero terms, and reports how many
// terms h holds. h needs room for len(e) + len(f).
func fastExpansionSum(e, f, h []float64) int {
	elen, flen := len(e), len(f)
	eindex, findex := 0, 0
	enow, fnow := e[0], f[0]

	var q float64
	if (fnow > enow) == (fnow > -enow) {
		q = enow
		eindex++
	} else {
		q = fnow
		findex++
	}
	if eindex < elen {
		enow = e[eindex]
	}
	if findex < flen {
		fnow = f[findex]
	}

	hindex := 0
	var hh float64
	if eindex < elen && findex < flen {
		if (fnow > enow) == (fnow > -enow) {
			q, hh = fastTwoSum(enow, q)
			eindex++
			if eindex < elen {
				enow = e[eindex]
			}
		} else {
			q, hh = fastTwoSum(fnow, q)
			findex++
			if findex < flen {
				fnow = f[findex]
			}
		}
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
		for eindex < elen && findex < flen {
			if (fnow > enow) == (fnow > -enow) {
				q, hh = twoSum(q, enow)
				eindex++
				if eindex < elen {
					enow = e[eindex]
				}
			} else {
				q, hh = twoSum(q, fnow)
				findex++
				if findex < flen {
					fnow = f[findex]
				}
			}
			if hh != 0 {
				h[hindex] = hh
				hindex++
			}
		}
	}
	for eindex < elen {
		q, hh = twoSum(q, enow)
		eindex++
		if eindex < elen {
			enow = e[eindex]
		}
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
	}
	for findex < flen {
		q, hh = twoSum(q, fnow)
		findex++
		if findex < flen {
			fnow = f[findex]
		}
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
	}
	if q != 0 || hindex == 0 {
		h[hindex] = q
		hindex++
	}
	return hindex
}

// Multiplies an expansion by a float into h, dropping zero terms, and
// reports how many terms h holds. h needs room for 2 * len(e).
func scaleExpansion(e []float64, b float64, h []float64) int {
	bhi, blo := split(b)

	q, hh := twoProductPresplit(e[0], b, bhi, blo)
	hindex := 0
	if hh != 0 {
		h[hindex] = hh
		hindex++
	}
	for _, enow := range e[1:] {
		product1, product0 := twoProductPresplit(enow, b, bhi, blo)
		var sum float64
		sum, hh = twoSum(q, product0)
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
		q, hh = fastTwoSum(product1, sum)
		if hh != 0 {
			h[hindex] = hh
			hindex++
		}
	}
	if q != 0 || hindex == 0 {
		h[hindex] = q
		hindex++
	}
	return hindex
}

// The float nearest an expansion's value, good enough to compare against an
// error bound.
func estimate(e []float64) float64 {
	q := 0.
	for _, term := range e {
		q += term
	}
	return q
}
