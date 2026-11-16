package extrude

import (
	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// frame places one ring of a sweep: the profile's x lands on x and its y on
// y, and x × tangent = y, so the profile is never mirrored.
type frame struct {
	geometry.PathFrame
	size vector2.Float64
}

const maxMiterStretch = geometry.MaxMiterStretch

func (f frame) place(p vector2.Float64) vector3.Float64 {
	return f.Origin.Add(f.offset(p))
}

func (f frame) offset(p vector2.Float64) vector3.Float64 {
	offset := f.direction(vector2.New(p.X()*f.size.X(), p.Y()*f.size.Y()))
	if f.Stretch > 1 {
		offset = offset.Add(f.Miter.Scale(offset.Dot(f.Miter) * (f.Stretch - 1)))
	}
	return offset
}

func (f frame) direction(p vector2.Float64) vector3.Float64 {
	return f.X.Scale(p.X()).Add(f.Y.Scale(p.Y()))
}

func (f frame) mirrored() bool {
	return f.size.X()*f.size.Y() < 0
}

// reach runs from a point on this ring to the same point on the next, so the
// normal is square to the surface however the rings taper, tilt or stretch.
// Where the two rings meet, or this one is scaled to nothing, it falls back
// to the profile's own normal.
func (f frame) normal(n vector2.Float64, scale float64, reach vector3.Float64) vector3.Float64 {
	tangent := vector2.New(-n.Y(), n.X())
	around := f.offset(tangent.Scale(scale))
	if around.Length() == 0 {
		around = f.direction(tangent)
	}

	normal := reach.Cross(around)
	if normal.Length() == 0 {
		normal = f.direction(vector2.New(n.X()*f.size.Y(), n.Y()*f.size.X()))
	}
	if normal.Length() == 0 {
		normal = f.direction(n)
	}
	if f.mirrored() {
		normal = normal.Scale(-1)
	}
	return normal.Normalized()
}

func tangents(path []vector3.Float64) []vector3.Float64 {
	if len(path) < 2 {
		return []vector3.Float64{vector3.Up[float64]()}
	}
	segments := geometry.SegmentDirections(path, false)
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

func pathFrames(path []vector3.Float64, closed bool) []frame {
	pathed := geometry.PathFrames(path, closed)
	frames := make([]frame, len(pathed))
	for i, f := range pathed {
		frames[i] = frame{PathFrame: f, size: vector2.One[float64]()}
	}
	return frames
}

// Travel runs along each transform's +Z, the profile's y along its +Y and x
// along its -X. Two transforms on the same spot make a bevel between them.
func framesFromTRS(transforms []trs.TRS, closed bool) []frame {
	count := len(transforms)
	frames := make([]frame, count)
	for i, t := range transforms {
		rotation := t.Rotation()
		scale := t.Scale()
		frames[i] = frame{
			PathFrame: geometry.PathFrame{
				Point:   i,
				Origin:  t.Position(),
				Tangent: rotation.Rotate(vector3.Forward[float64]()).Normalized(),
				X:       rotation.Rotate(vector3.Left[float64]()).Normalized(),
				Y:       rotation.Rotate(vector3.Up[float64]()).Normalized(),
			},
			size: vector2.New(scale.X(), scale.Y()),
		}
	}

	toward := func(from, to int) vector3.Float64 {
		along := frames[to].Origin.Sub(frames[from].Origin)
		if along.Length() <= 1e-12 {
			return vector3.Float64{}
		}
		return along.Normalized()
	}
	for i := range frames {
		if i > 0 || closed {
			frames[i].Incoming = toward((i+count-1)%count, i)
		}
		if i < count-1 || closed {
			frames[i].Outgoing = toward(i, (i+1)%count)
		}
		if frames[i].Incoming != (vector3.Float64{}) && frames[i].Outgoing != (vector3.Float64{}) {
			// Flattened into the frame's plane, so a frame that does not face
			// the corner evenly still places a flat section.
			miter, stretch := geometry.Miter(frames[i].Incoming, frames[i].Outgoing)
			frames[i].Miter = miter.Sub(frames[i].Tangent.Scale(miter.Dot(frames[i].Tangent)))
			frames[i].Stretch = stretch
		}
	}
	return frames
}

func distances(frames []frame) []float64 {
	walked := make([]float64, len(frames))
	for i := 1; i < len(frames); i++ {
		walked[i] = walked[i-1] + frames[i].Origin.Distance(frames[i-1].Origin)
	}
	return walked
}
