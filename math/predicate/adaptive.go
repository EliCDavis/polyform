package predicate

import (
	"math"

	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// Each predicate runs in stages, each exact to a tighter error bound than
// the last, and stops at the first whose result clears its bound. Nearly
// every call stops at stage A, plain floating point; the rest is only paid
// for inputs near degenerate.

// orient2D is the full adaptive predicate: stage D is exact.
func orient2D(a, b, c vector2.Float64) float64 {
	ax, ay := a.Values()
	bx, by := b.Values()
	cx, cy := c.Values()
	detleft := float64((ax - cx) * (by - cy))
	detright := float64((ay - cy) * (bx - cx))
	det := detleft - detright

	var detsum float64
	switch {
	case detleft > 0:
		if detright <= 0 {
			return det
		}
		detsum = detleft + detright
	case detleft < 0:
		if detright >= 0 {
			return det
		}
		detsum = -detleft - detright
	default:
		return det
	}

	if errbound := ccwErrBoundA * detsum; det >= errbound || -det >= errbound {
		return det
	}
	return orient2DAdapt(ax, ay, bx, by, cx, cy, detsum)
}

func orient2DAdapt(ax, ay, bx, by, cx, cy, detsum float64) float64 {
	acx, acy := ax-cx, ay-cy
	bcx, bcy := bx-cx, by-cy

	detleft, detlefttail := twoProduct(acx, bcy)
	detright, detrighttail := twoProduct(acy, bcx)

	var b [4]float64
	b[3], b[2], b[1], b[0] = twoTwoDiff(detleft, detlefttail, detright, detrighttail)

	det := estimate(b[:])
	if errbound := ccwErrBoundB * detsum; det >= errbound || -det >= errbound {
		return det
	}

	acxtail := twoDiffTail(ax, cx, acx)
	bcxtail := twoDiffTail(bx, cx, bcx)
	acytail := twoDiffTail(ay, cy, acy)
	bcytail := twoDiffTail(by, cy, bcy)
	if acxtail == 0 && acytail == 0 && bcxtail == 0 && bcytail == 0 {
		return det
	}

	errbound := ccwErrBoundC*detsum + resultErrBound*math.Abs(det)
	det += (float64(acx*bcytail) + float64(bcy*acxtail)) - (float64(acy*bcxtail) + float64(bcx*acytail))
	if det >= errbound || -det >= errbound {
		return det
	}

	var u [4]float64
	var c1 [8]float64
	var c2 [12]float64
	var d [16]float64

	s1, s0 := twoProduct(acxtail, bcy)
	t1, t0 := twoProduct(acytail, bcx)
	u[3], u[2], u[1], u[0] = twoTwoDiff(s1, s0, t1, t0)
	c1len := fastExpansionSum(b[:], u[:], c1[:])

	s1, s0 = twoProduct(acx, bcytail)
	t1, t0 = twoProduct(acy, bcxtail)
	u[3], u[2], u[1], u[0] = twoTwoDiff(s1, s0, t1, t0)
	c2len := fastExpansionSum(c1[:c1len], u[:], c2[:])

	s1, s0 = twoProduct(acxtail, bcytail)
	t1, t0 = twoProduct(acytail, bcxtail)
	u[3], u[2], u[1], u[0] = twoTwoDiff(s1, s0, t1, t0)
	dlen := fastExpansionSum(c2[:c2len], u[:], d[:])

	return d[dlen-1]
}

// orient3D runs stages A to C. ok is false when the answer is still inside
// the error bound after those, which needs exact arithmetic.
//
// Shewchuk's sign is positive with d below the plane; this returns the
// opposite, positive on the side the normal (b-a)x(c-a) points to.
func orient3D(a, b, c, d vector3.Float64) (float64, bool) {
	ax, ay, az := a.Values()
	bx, by, bz := b.Values()
	cx, cy, cz := c.Values()
	dx, dy, dz := d.Values()
	adx, ady, adz := ax-dx, ay-dy, az-dz
	bdx, bdy, bdz := bx-dx, by-dy, bz-dz
	cdx, cdy, cdz := cx-dx, cy-dy, cz-dz

	bdxcdy := float64(bdx * cdy)
	cdxbdy := float64(cdx * bdy)
	cdxady := float64(cdx * ady)
	adxcdy := float64(adx * cdy)
	adxbdy := float64(adx * bdy)
	bdxady := float64(bdx * ady)

	det := float64(adz*(bdxcdy-cdxbdy)) + float64(bdz*(cdxady-adxcdy)) + float64(cdz*(adxbdy-bdxady))

	permanent := float64((math.Abs(bdxcdy)+math.Abs(cdxbdy))*math.Abs(adz)) +
		float64((math.Abs(cdxady)+math.Abs(adxcdy))*math.Abs(bdz)) +
		float64((math.Abs(adxbdy)+math.Abs(bdxady))*math.Abs(cdz))

	if errbound := o3dErrBoundA * permanent; det > errbound || -det > errbound {
		return -det, true
	}

	// Stage B: the same determinant with nothing rounded, from the rounded
	// differences.
	var bc, ca, ab [4]float64
	var adet, bdet, cdet [8]float64
	var abdet [16]float64
	var fin [24]float64

	bdxcdy1, bdxcdy0 := twoProduct(bdx, cdy)
	cdxbdy1, cdxbdy0 := twoProduct(cdx, bdy)
	bc[3], bc[2], bc[1], bc[0] = twoTwoDiff(bdxcdy1, bdxcdy0, cdxbdy1, cdxbdy0)
	alen := scaleExpansion(bc[:], adz, adet[:])

	cdxady1, cdxady0 := twoProduct(cdx, ady)
	adxcdy1, adxcdy0 := twoProduct(adx, cdy)
	ca[3], ca[2], ca[1], ca[0] = twoTwoDiff(cdxady1, cdxady0, adxcdy1, adxcdy0)
	blen := scaleExpansion(ca[:], bdz, bdet[:])

	adxbdy1, adxbdy0 := twoProduct(adx, bdy)
	bdxady1, bdxady0 := twoProduct(bdx, ady)
	ab[3], ab[2], ab[1], ab[0] = twoTwoDiff(adxbdy1, adxbdy0, bdxady1, bdxady0)
	clen := scaleExpansion(ab[:], cdz, cdet[:])

	ablen := fastExpansionSum(adet[:alen], bdet[:blen], abdet[:])
	finlen := fastExpansionSum(abdet[:ablen], cdet[:clen], fin[:])

	det = estimate(fin[:finlen])
	if errbound := o3dErrBoundB * permanent; det >= errbound || -det >= errbound {
		return -det, true
	}

	// Stage C: fold in the bits the differences lost, to first order.
	adxtail := twoDiffTail(ax, dx, adx)
	adytail := twoDiffTail(ay, dy, ady)
	adztail := twoDiffTail(az, dz, adz)
	bdxtail := twoDiffTail(bx, dx, bdx)
	bdytail := twoDiffTail(by, dy, bdy)
	bdztail := twoDiffTail(bz, dz, bdz)
	cdxtail := twoDiffTail(cx, dx, cdx)
	cdytail := twoDiffTail(cy, dy, cdy)
	cdztail := twoDiffTail(cz, dz, cdz)
	if adxtail == 0 && bdxtail == 0 && cdxtail == 0 &&
		adytail == 0 && bdytail == 0 && cdytail == 0 &&
		adztail == 0 && bdztail == 0 && cdztail == 0 {
		return -det, true
	}

	errbound := o3dErrBoundC*permanent + resultErrBound*math.Abs(det)
	det += (float64(adz*((float64(bdx*cdytail)+float64(cdy*bdxtail))-(float64(bdy*cdxtail)+float64(cdx*bdytail)))) +
		float64(adztail*(float64(bdx*cdy)-float64(bdy*cdx)))) +
		(float64(bdz*((float64(cdx*adytail)+float64(ady*cdxtail))-(float64(cdy*adxtail)+float64(adx*cdytail)))) +
			float64(bdztail*(float64(cdx*ady)-float64(cdy*adx)))) +
		(float64(cdz*((float64(adx*bdytail)+float64(bdy*adxtail))-(float64(ady*bdxtail)+float64(bdx*adytail)))) +
			float64(cdztail*(float64(adx*bdy)-float64(ady*bdx))))
	if det >= errbound || -det >= errbound {
		return -det, true
	}
	return 0, false
}

// inCircle runs stages A to C. ok is false when the answer is still inside
// the error bound after those, which needs exact arithmetic.
func inCircle(a, b, c, d vector2.Float64) (float64, bool) {
	ax, ay := a.Values()
	bx, by := b.Values()
	cx, cy := c.Values()
	dx, dy := d.Values()
	adx, ady := ax-dx, ay-dy
	bdx, bdy := bx-dx, by-dy
	cdx, cdy := cx-dx, cy-dy

	bdxcdy := float64(bdx * cdy)
	cdxbdy := float64(cdx * bdy)
	alift := float64(adx*adx) + float64(ady*ady)

	cdxady := float64(cdx * ady)
	adxcdy := float64(adx * cdy)
	blift := float64(bdx*bdx) + float64(bdy*bdy)

	adxbdy := float64(adx * bdy)
	bdxady := float64(bdx * ady)
	clift := float64(cdx*cdx) + float64(cdy*cdy)

	det := float64(alift*(bdxcdy-cdxbdy)) + float64(blift*(cdxady-adxcdy)) + float64(clift*(adxbdy-bdxady))

	permanent := float64((math.Abs(bdxcdy)+math.Abs(cdxbdy))*alift) +
		float64((math.Abs(cdxady)+math.Abs(adxcdy))*blift) +
		float64((math.Abs(adxbdy)+math.Abs(bdxady))*clift)

	if errbound := iccErrBoundA * permanent; det > errbound || -det > errbound {
		return det, true
	}

	// Stage B.
	var bc, ca, ab [4]float64
	var axbc, aybc, bxca, byca, cxab, cyab [8]float64
	var axxbc, ayybc, bxxca, byyca, cxxab, cyyab [16]float64
	var adet, bdet, cdet [32]float64
	var abdet [64]float64
	var fin [96]float64

	bdxcdy1, bdxcdy0 := twoProduct(bdx, cdy)
	cdxbdy1, cdxbdy0 := twoProduct(cdx, bdy)
	bc[3], bc[2], bc[1], bc[0] = twoTwoDiff(bdxcdy1, bdxcdy0, cdxbdy1, cdxbdy0)
	axbclen := scaleExpansion(bc[:], adx, axbc[:])
	axxbclen := scaleExpansion(axbc[:axbclen], adx, axxbc[:])
	aybclen := scaleExpansion(bc[:], ady, aybc[:])
	ayybclen := scaleExpansion(aybc[:aybclen], ady, ayybc[:])
	alen := fastExpansionSum(axxbc[:axxbclen], ayybc[:ayybclen], adet[:])

	cdxady1, cdxady0 := twoProduct(cdx, ady)
	adxcdy1, adxcdy0 := twoProduct(adx, cdy)
	ca[3], ca[2], ca[1], ca[0] = twoTwoDiff(cdxady1, cdxady0, adxcdy1, adxcdy0)
	bxcalen := scaleExpansion(ca[:], bdx, bxca[:])
	bxxcalen := scaleExpansion(bxca[:bxcalen], bdx, bxxca[:])
	bycalen := scaleExpansion(ca[:], bdy, byca[:])
	byycalen := scaleExpansion(byca[:bycalen], bdy, byyca[:])
	blen := fastExpansionSum(bxxca[:bxxcalen], byyca[:byycalen], bdet[:])

	adxbdy1, adxbdy0 := twoProduct(adx, bdy)
	bdxady1, bdxady0 := twoProduct(bdx, ady)
	ab[3], ab[2], ab[1], ab[0] = twoTwoDiff(adxbdy1, adxbdy0, bdxady1, bdxady0)
	cxablen := scaleExpansion(ab[:], cdx, cxab[:])
	cxxablen := scaleExpansion(cxab[:cxablen], cdx, cxxab[:])
	cyablen := scaleExpansion(ab[:], cdy, cyab[:])
	cyyablen := scaleExpansion(cyab[:cyablen], cdy, cyyab[:])
	clen := fastExpansionSum(cxxab[:cxxablen], cyyab[:cyyablen], cdet[:])

	ablen := fastExpansionSum(adet[:alen], bdet[:blen], abdet[:])
	finlen := fastExpansionSum(abdet[:ablen], cdet[:clen], fin[:])

	det = estimate(fin[:finlen])
	if errbound := iccErrBoundB * permanent; det >= errbound || -det >= errbound {
		return det, true
	}

	// Stage C.
	adxtail := twoDiffTail(ax, dx, adx)
	adytail := twoDiffTail(ay, dy, ady)
	bdxtail := twoDiffTail(bx, dx, bdx)
	bdytail := twoDiffTail(by, dy, bdy)
	cdxtail := twoDiffTail(cx, dx, cdx)
	cdytail := twoDiffTail(cy, dy, cdy)
	if adxtail == 0 && bdxtail == 0 && cdxtail == 0 &&
		adytail == 0 && bdytail == 0 && cdytail == 0 {
		return det, true
	}

	errbound := iccErrBoundC*permanent + resultErrBound*math.Abs(det)
	det += (float64((float64(adx*adx)+float64(ady*ady))*((float64(bdx*cdytail)+float64(cdy*bdxtail))-(float64(bdy*cdxtail)+float64(cdx*bdytail)))) +
		float64(2*(float64(adx*adxtail)+float64(ady*adytail))*(float64(bdx*cdy)-float64(bdy*cdx)))) +
		(float64((float64(bdx*bdx)+float64(bdy*bdy))*((float64(cdx*adytail)+float64(ady*cdxtail))-(float64(cdy*adxtail)+float64(adx*cdytail)))) +
			float64(2*(float64(bdx*bdxtail)+float64(bdy*bdytail))*(float64(cdx*ady)-float64(cdy*adx)))) +
		(float64((float64(cdx*cdx)+float64(cdy*cdy))*((float64(adx*bdytail)+float64(bdy*adxtail))-(float64(ady*bdxtail)+float64(bdx*adytail)))) +
			float64(2*(float64(cdx*cdxtail)+float64(cdy*cdytail))*(float64(adx*bdy)-float64(ady*bdx))))
	if det >= errbound || -det >= errbound {
		return det, true
	}
	return 0, false
}
