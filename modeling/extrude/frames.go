package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// frame places one ring of a sweep: the profile's x lands on x and its y on
// y, and x × tangent = y, so the profile is never mirrored.
type frame struct {
	point   int
	origin  vector3.Float64
	tangent vector3.Float64
	x, y    vector3.Float64

	// Zero where the ring ends the path or is one side of a bevel.
	incoming, outgoing vector3.Float64

	miter   vector3.Float64
	stretch float64
}

// Turns sharper than this are bevelled with two rings instead of mitred.
const maxMiterStretch = 4.

func (f frame) place(p vector2.Float64) vector3.Float64 {
	offset := f.direction(p)
	if f.stretch > 1 {
		offset = offset.Add(f.miter.Scale(offset.Dot(f.miter) * (f.stretch - 1)))
	}
	return f.origin.Add(offset)
}

func (f frame) direction(p vector2.Float64) vector3.Float64 {
	return f.x.Scale(p.X()).Add(f.y.Scale(p.Y()))
}

func (f frame) facing(dir vector3.Float64) frame {
	turn := quaternion.RotationTo(f.tangent, dir)
	f.tangent, f.x, f.y = dir, turn.Rotate(f.x), turn.Rotate(f.y)
	return f
}

func (f frame) rolled(angle float64) frame {
	sin, cos := math.Sincos(angle)
	f.x, f.y = f.x.Scale(cos).Sub(f.y.Scale(sin)), f.y.Scale(cos).Add(f.x.Scale(sin))
	return f
}

// A zero-length segment takes its neighbour's direction, so the rings either
// side of it share an orientation instead of losing one.
func segmentDirections(path []vector3.Float64, closed bool) []vector3.Float64 {
	count := len(path) - 1
	if closed {
		count = len(path)
	}
	directions := make([]vector3.Float64, count)
	for i := range directions {
		if along := path[(i+1)%len(path)].Sub(path[i]); along.Length() > 0 {
			directions[i] = along.Normalized()
		}
	}

	var last vector3.Float64
	for i := range directions {
		if directions[i] == (vector3.Float64{}) {
			directions[i] = last
		}
		last = directions[i]
	}
	last = vector3.Up[float64]()
	for i := len(directions) - 1; i >= 0; i-- {
		if directions[i] == (vector3.Float64{}) {
			directions[i] = last
		}
		last = directions[i]
	}
	return directions
}

func tangents(path []vector3.Float64) []vector3.Float64 {
	if len(path) < 2 {
		return []vector3.Float64{vector3.Up[float64]()}
	}
	segments := segmentDirections(path, false)
	result := make([]vector3.Float64, len(path))
	for i := range path {
		switch {
		case i == 0:
			result[i] = segments[0]
		case i == len(path)-1:
			result[i] = segments[i-1]
		default:
			result[i] = segments[i-1].Add(segments[i]).Normalized()
		}
	}
	return result
}

// pathFrames carries one frame along the path without twisting it, rolled as
// a whole so the profile's y points up wherever the path runs level.
func pathFrames(path []vector3.Float64, closed bool) []frame {
	segments := segmentDirections(path, closed)
	frames := make([]frame, 0, len(path))
	for i, origin := range path {
		var incoming, outgoing vector3.Float64
		if i > 0 || closed {
			incoming = segments[(i+len(segments)-1)%len(segments)]
		}
		if i < len(segments) {
			outgoing = segments[i]
		}

		switch {
		case incoming == (vector3.Float64{}):
			frames = append(frames, frame{point: i, origin: origin, tangent: outgoing, outgoing: outgoing})
		case outgoing == (vector3.Float64{}):
			frames = append(frames, frame{point: i, origin: origin, tangent: incoming, incoming: incoming})
		default:
			// |incoming + outgoing| is 2cos(half the turn), and its inverse
			// is the stretch that keeps the sweep's width through the turn.
			bisector := incoming.Add(outgoing)
			if bisector.Length() < 2/maxMiterStretch {
				frames = append(frames,
					frame{point: i, origin: origin, tangent: incoming, incoming: incoming},
					frame{point: i, origin: origin, tangent: outgoing, outgoing: outgoing},
				)
				continue
			}
			f := frame{point: i, origin: origin, tangent: bisector.Normalized(), incoming: incoming, outgoing: outgoing}
			if across := outgoing.Sub(incoming); across.Length() > 1e-9 {
				f.miter = across.Normalized()
				f.stretch = 2 / bisector.Length()
			}
			frames = append(frames, f)
		}
	}

	frames[0].x = quaternion.RotationTo(vector3.Up[float64](), frames[0].tangent).Rotate(vector3.Right[float64]())
	frames[0].y = frames[0].x.Cross(frames[0].tangent)
	for i := 1; i < len(frames); i++ {
		frames[i].x = carried(frames[i-1], frames[i])
		frames[i].y = frames[i].x.Cross(frames[i].tangent)
	}

	if closed {
		spreadSeamTwist(frames)
	}
	return levelled(frames)
}

// carried turns from's x the way the path turns between the two rings, so a
// straight segment adds no twist and a corner only the turn itself.
func carried(from, to frame) vector3.Float64 {
	x := from.x
	if chord := to.origin.Sub(from.origin); chord.Length() > 0 {
		along := chord.Normalized()
		x = quaternion.RotationTo(from.tangent, along).Rotate(x)
		x = quaternion.RotationTo(along, to.tangent).Rotate(x)
	} else {
		x = quaternion.RotationTo(from.tangent, to.tangent).Rotate(x)
	}

	// Rotations accumulate error, so x is squared back up against the
	// tangent rather than left to shear over a long path.
	x = x.Sub(to.tangent.Scale(x.Dot(to.tangent)))
	if x.Length() == 0 {
		return to.tangent.Perpendicular().Normalized()
	}
	return x.Normalized()
}

// A frame carried round a loop does not come back to where it started. The
// shortfall is shared out by distance travelled, rather than wrung out on the
// segment that closes the loop.
func spreadSeamTwist(frames []frame) {
	last := frames[len(frames)-1]
	returned := carried(last, frames[0])
	first := frames[0]
	mismatch := math.Atan2(first.x.Cross(returned).Dot(first.tangent), first.x.Dot(returned))
	if mismatch == 0 {
		return
	}

	walked := distances(frames)
	total := walked[len(walked)-1] + last.origin.Distance(first.origin)
	if total == 0 {
		return
	}
	for i := range frames {
		frames[i] = frames[i].rolled(-mismatch * walked[i] / total)
	}
}

func distances(frames []frame) []float64 {
	walked := make([]float64, len(frames))
	for i := 1; i < len(frames); i++ {
		walked[i] = walked[i-1] + frames[i].origin.Distance(frames[i-1].origin)
	}
	return walked
}

// levelled rolls every frame by the one angle that best points y up, each
// ring weighted by how much path it stands for and how level the path runs
// there. Rings on a vertical run carry no weight, so they never twist to
// chase an up they cannot reach.
func levelled(frames []frame) []frame {
	up := vector3.Up[float64]()
	walked := distances(frames)

	var alongY, alongX float64
	for i, f := range frames {
		reach := 0.
		if i > 0 {
			reach += walked[i] - walked[i-1]
		}
		if i < len(frames)-1 {
			reach += walked[i+1] - walked[i]
		}
		level := up.Sub(f.tangent.Scale(up.Dot(f.tangent))).Length()
		alongY += reach * level * f.y.Dot(up)
		alongX += reach * level * f.x.Dot(up)
	}
	if math.Hypot(alongX, alongY) < 1e-12 {
		return frames
	}

	angle := math.Atan2(alongX, alongY)
	for i := range frames {
		frames[i] = frames[i].rolled(angle)
	}
	return frames
}
