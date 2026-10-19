package mcp

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sort"

	"github.com/fogleman/fauxgl"
)

const (
	defaultFov = 28.
	maxFov     = 100.
	fitMargin  = 1.1
)

type cameraPose struct {
	eye, target, up fauxgl.Vector
	distance, fov   float64
}

func (p cameraPose) near() float64 { return p.distance * 0.05 }
func (p cameraPose) far() float64  { return p.distance * 10 }

func (p cameraPose) direction() fauxgl.Vector {
	return p.eye.Sub(p.target).Normalize()
}

// fitRadiusOf is the radius of the sphere around the target that fills the
// frame's height, less the fit margin: what fitRadius would have asked for.
func fitRadiusOf(distance, fov float64) float64 {
	return distance * math.Tan(fov*math.Pi/360) / fitMargin
}

func distanceForRadius(radius, fov float64) float64 {
	return fitMargin * radius / math.Tan(fov*math.Pi/360)
}

// azimuthElevationDirection returns the unit vector from a target toward a
// camera placed at the given horizontal/vertical angles (degrees).
func azimuthElevationDirection(azimuthDeg, elevationDeg float64) fauxgl.Vector {
	az := azimuthDeg * math.Pi / 180
	el := elevationDeg * math.Pi / 180
	return fauxgl.V(
		math.Cos(el)*math.Sin(az),
		math.Sin(el),
		math.Cos(el)*math.Cos(az),
	)
}

func directionAngles(dir fauxgl.Vector) (azimuthDeg, elevationDeg float64) {
	dir = dir.Normalize()
	return math.Atan2(dir.X, dir.Z) * 180 / math.Pi, math.Asin(math.Max(-1, math.Min(1, dir.Y))) * 180 / math.Pi
}

// cameraUp is how azimuthElevationDirection moves as elevation rises. Unlike
// world up it stays defined straight above and below, where it follows
// azimuth.
func cameraUp(azimuthDeg, elevationDeg float64) fauxgl.Vector {
	az := azimuthDeg * math.Pi / 180
	el := elevationDeg * math.Pi / 180
	return fauxgl.V(
		-math.Sin(el)*math.Sin(az),
		math.Cos(el),
		-math.Sin(el)*math.Cos(az),
	)
}

func boxCorners(box fauxgl.Box) [8]fauxgl.Vector {
	var corners [8]fauxgl.Vector
	for i := range corners {
		corners[i] = fauxgl.V(
			pick(i&1 == 0, box.Min.X, box.Max.X),
			pick(i&2 == 0, box.Min.Y, box.Max.Y),
			pick(i&4 == 0, box.Min.Z, box.Max.Z),
		)
	}
	return corners
}

func pick(first bool, a, b float64) float64 {
	if first {
		return a
	}
	return b
}

// fitDistance is how far back along dir from box's center a camera with
// this fov and aspect must sit for every corner to land inside the frame,
// with the fit margin around the widest one.
func fitDistance(box fauxgl.Box, dir, up fauxgl.Vector, aspect, fov float64) float64 {
	right := up.Cross(dir)
	if right.Length() < 1e-9 {
		right = fauxgl.V(1, 0, 0)
	}
	right = right.Normalize()
	trueUp := dir.Cross(right).Normalize()
	tanHalf := math.Tan(fov * math.Pi / 360)

	center := box.Center()
	distance := 0.
	for _, corner := range boxCorners(box) {
		rel := corner.Sub(center)
		depth := rel.Dot(dir)
		h := math.Abs(rel.Dot(trueUp)) * fitMargin / tanHalf
		w := math.Abs(rel.Dot(right)) * fitMargin / (tanHalf * aspect)
		distance = math.Max(distance, depth+math.Max(h, w))
	}
	if distance <= 0 {
		return 2.2
	}
	return distance
}

type rayHit struct {
	t        float64
	entering bool
}

// lineOfSight lists where the segment from origin to eye crosses the scene,
// nearest the origin first. A crossing is entering when it goes in through
// a front face: something solid stands there between the two points.
func lineOfSight(triangles []*fauxgl.Triangle, origin, eye fauxgl.Vector) []rayHit {
	span := eye.Sub(origin)
	maxT := span.Length()
	if maxT == 0 {
		return nil
	}
	dir := span.DivScalar(maxT)
	var hits []rayHit
	for _, tri := range triangles {
		t, ok := rayTriangle(origin, dir, tri.V1.Position, tri.V2.Position, tri.V3.Position)
		if !ok || t <= 0 || t >= maxT {
			continue
		}
		hits = append(hits, rayHit{t: t, entering: tri.Normal().Dot(dir) < 0})
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].t < hits[j].t })
	return hits
}

func rayTriangle(origin, dir, a, b, c fauxgl.Vector) (float64, bool) {
	e1, e2 := b.Sub(a), c.Sub(a)
	p := dir.Cross(e2)
	det := e1.Dot(p)
	if math.Abs(det) < 1e-12 {
		return 0, false
	}
	inv := 1 / det
	s := origin.Sub(a)
	u := s.Dot(p) * inv
	if u < 0 || u > 1 {
		return 0, false
	}
	q := s.Cross(e1)
	v := dir.Dot(q) * inv
	if v < 0 || u+v > 1 {
		return 0, false
	}
	return e2.Dot(q) * inv, true
}

func firstOccluder(triangles []*fauxgl.Triangle, pose cameraPose) (float64, bool) {
	eps := pose.distance * 1e-6
	for _, hit := range lineOfSight(triangles, pose.target, pose.eye) {
		if hit.t > eps && hit.entering {
			return hit.t, true
		}
	}
	return 0, false
}

// clearCamera pulls an auto-placed camera in front of whatever stands
// between it and its target, widening the field of view so the framed
// radius keeps its size in the picture.
func clearCamera(pose cameraPose, triangles []*fauxgl.Triangle) (cameraPose, string) {
	blockedAt, blocked := firstOccluder(triangles, pose)
	if !blocked {
		return pose, ""
	}
	radius := fitRadiusOf(pose.distance, pose.fov)
	pulled := blockedAt * 0.9
	if pulled < radius*0.5 {
		return pose, fmt.Sprintf("geometry %.3g from the target blocks this view and leaves no room in front of it; move target, change the angle, or exclude the part", blockedAt)
	}
	fov := math.Min(2*math.Atan(fitMargin*radius/pulled)*180/math.Pi, maxFov)
	moved := pose
	moved.distance = pulled
	moved.eye = pose.target.Add(pose.direction().MulScalar(pulled))
	moved.fov = fov
	return moved, fmt.Sprintf("camera pulled in from %.3g to %.3g (fov widened to %.0f) to get in front of geometry %.3g from the target that blocked the view", pose.distance, pulled, fov, blockedAt)
}

// gridColumns lays n views out with no empty cell when a column count from
// the square root up to five divides n; otherwise the squarest grid, whose
// spare cells the compositor fills with background.
func gridColumns(n, requested int) int {
	if requested > 0 {
		return min(requested, n)
	}
	if n <= 3 {
		return n
	}
	base := int(math.Ceil(math.Sqrt(float64(n))))
	for cols := base; cols <= 5 && cols <= n; cols++ {
		if n%cols == 0 {
			return cols
		}
	}
	return base
}

// coverage is the share of pixels that differ from the clear color.
func coverage(img image.Image, background color.NRGBA) float64 {
	bounds := img.Bounds()
	total := bounds.Dx() * bounds.Dy()
	if total == 0 {
		return 0
	}
	drawn := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if color.NRGBAModel.Convert(img.At(x, y)).(color.NRGBA) != background {
				drawn++
			}
		}
	}
	return float64(drawn) / float64(total)
}
