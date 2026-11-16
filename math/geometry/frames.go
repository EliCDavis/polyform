package geometry

import (
	"math"

	"github.com/EliCDavis/polyform/math/mat"
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

// PathFrame orients one cross-section of a path. X is the section's right
// looking along the path and Y its up, with X × Tangent = Y.
type PathFrame struct {
	Point   int
	Origin  vector3.Float64
	Tangent vector3.Float64
	X, Y    vector3.Float64

	// Zero where the frame ends the path or is one side of a bevel.
	Incoming, Outgoing vector3.Float64

	Miter   vector3.Float64
	Stretch float64
}

// Turns sharper than this are bevelled with two frames instead of mitred.
const MaxMiterStretch = 4.

// The frame's travel runs along local +Z, its up along +Y and its right
// along -X.
func (f PathFrame) TRS() trs.TRS {
	right := f.X.Scale(-1)
	basis := mat.Matrix4x4{
		X00: right.X(), X01: f.Y.X(), X02: f.Tangent.X(),
		X10: right.Y(), X11: f.Y.Y(), X12: f.Tangent.Y(),
		X20: right.Z(), X21: f.Y.Z(), X22: f.Tangent.Z(),
		X33: 1,
	}
	return trs.New(f.Origin, quaternion.FromMatrix(basis), vector3.One[float64]())
}

func (f PathFrame) Rolled(angle float64) PathFrame {
	sin, cos := math.Sincos(angle)
	f.X, f.Y = f.X.Scale(cos).Sub(f.Y.Scale(sin)), f.Y.Scale(cos).Add(f.X.Scale(sin))
	return f
}

// A zero-length segment takes its neighbour's direction, so the frames
// either side of it share an orientation instead of losing one.
func SegmentDirections(path []vector3.Float64, closed bool) []vector3.Float64 {
	count := len(path) - 1
	if closed {
		count = len(path)
	}
	directions := make([]vector3.Float64, max(count, 0))
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

// PathFrames carries one frame along the path without twisting it, rolled as
// a whole so Y points up wherever the path runs level.
func PathFrames(path []vector3.Float64, closed bool) []PathFrame {
	if len(path) < 2 {
		return nil
	}
	segments := SegmentDirections(path, closed)
	frames := make([]PathFrame, 0, len(path))
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
			frames = append(frames, PathFrame{Point: i, Origin: origin, Tangent: outgoing, Outgoing: outgoing})
		case outgoing == (vector3.Float64{}):
			frames = append(frames, PathFrame{Point: i, Origin: origin, Tangent: incoming, Incoming: incoming})
		default:
			// |incoming + outgoing| is 2cos(half the turn), and its inverse
			// is the stretch that keeps the sweep's width through the turn.
			bisector := incoming.Add(outgoing)
			if bisector.Length() < 2/MaxMiterStretch {
				frames = append(frames,
					PathFrame{Point: i, Origin: origin, Tangent: incoming, Incoming: incoming},
					PathFrame{Point: i, Origin: origin, Tangent: outgoing, Outgoing: outgoing},
				)
				continue
			}
			f := PathFrame{Point: i, Origin: origin, Tangent: bisector.Normalized(), Incoming: incoming, Outgoing: outgoing}
			f.Miter, f.Stretch = Miter(incoming, outgoing)
			frames = append(frames, f)
		}
	}

	frames[0].X = plumbX(frames[0].Tangent)
	frames[0].Y = frames[0].X.Cross(frames[0].Tangent)
	for i := 1; i < len(frames); i++ {
		frames[i].X = carried(frames[i-1], frames[i])
		frames[i].Y = frames[i].X.Cross(frames[i].Tangent)
	}

	if closed {
		spreadSeamTwist(frames)
	}
	return levelled(frames)
}

// Tilted over from whichever plumb direction is nearer: tilting a falling
// tangent from straight up would spin X with the way it leans.
func plumbX(tangent vector3.Float64) vector3.Float64 {
	if tangent.Y() < 0 {
		return quaternion.RotationTo(vector3.Down[float64](), tangent).Rotate(vector3.Left[float64]())
	}
	return quaternion.RotationTo(vector3.Up[float64](), tangent).Rotate(vector3.Right[float64]())
}

// Miter is the direction a section through a corner is stretched along, and
// by how much, to keep the sweep's width through the turn.
func Miter(incoming, outgoing vector3.Float64) (vector3.Float64, float64) {
	bisector := incoming.Add(outgoing)
	across := outgoing.Sub(incoming)
	if across.Length() <= 1e-9 || bisector.Length() < 2/MaxMiterStretch {
		return vector3.Float64{}, 0
	}
	return across.Normalized(), 2 / bisector.Length()
}

// carried turns from's X the way the path turns between the two frames, so a
// straight segment adds no twist and a corner only the turn itself.
func carried(from, to PathFrame) vector3.Float64 {
	x := from.X
	if chord := to.Origin.Sub(from.Origin); chord.Length() > 0 {
		along := chord.Normalized()
		x = quaternion.RotationTo(from.Tangent, along).Rotate(x)
		x = quaternion.RotationTo(along, to.Tangent).Rotate(x)
	} else {
		x = quaternion.RotationTo(from.Tangent, to.Tangent).Rotate(x)
	}

	// Rotations accumulate error, so X is squared back up against the
	// tangent rather than left to shear over a long path.
	x = x.Sub(to.Tangent.Scale(x.Dot(to.Tangent)))
	if x.Length() == 0 {
		return to.Tangent.Perpendicular().Normalized()
	}
	return x.Normalized()
}

// A frame carried round a loop does not come back to where it started. The
// shortfall is shared out by distance travelled, rather than wrung out on the
// segment that closes the loop.
func spreadSeamTwist(frames []PathFrame) {
	last := frames[len(frames)-1]
	returned := carried(last, frames[0])
	first := frames[0]
	mismatch := math.Atan2(first.X.Cross(returned).Dot(first.Tangent), first.X.Dot(returned))
	if mismatch == 0 {
		return
	}

	walked := FrameDistances(frames)
	total := walked[len(walked)-1] + last.Origin.Distance(first.Origin)
	if total == 0 {
		return
	}
	for i := range frames {
		frames[i] = frames[i].Rolled(-mismatch * walked[i] / total)
	}
}

func FrameDistances(frames []PathFrame) []float64 {
	walked := make([]float64, len(frames))
	for i := 1; i < len(frames); i++ {
		walked[i] = walked[i-1] + frames[i].Origin.Distance(frames[i-1].Origin)
	}
	return walked
}

// A diagonal segment has to land on the level side whichever way it rounds.
const steepestLevelClimb = math.Sqrt2/2 + 1e-9

func runsLevel(frames []PathFrame) bool {
	for _, f := range frames {
		for _, along := range [2]vector3.Float64{f.Incoming, f.Outgoing} {
			if along != (vector3.Float64{}) && math.Abs(along.Y()) <= steepestLevelClimb {
				return true
			}
		}
	}
	return false
}

// levelled rolls every frame by the one angle that best points Y up, each
// frame weighted by how much path it stands for and how level the path runs
// there. Frames on a vertical run carry no weight, so they never twist to
// chase an up they cannot reach.
//
// A path steeper than a diagonal throughout is left as it was carried: the
// roll would hang on which way it leans, however slightly.
func levelled(frames []PathFrame) []PathFrame {
	if !runsLevel(frames) {
		return frames
	}
	up := vector3.Up[float64]()
	walked := FrameDistances(frames)

	var alongY, alongX float64
	for i, f := range frames {
		reach := 0.
		if i > 0 {
			reach += walked[i] - walked[i-1]
		}
		if i < len(frames)-1 {
			reach += walked[i+1] - walked[i]
		}
		level := up.Sub(f.Tangent.Scale(up.Dot(f.Tangent))).Length()
		alongY += reach * level * f.Y.Dot(up)
		alongX += reach * level * f.X.Dot(up)
	}
	if math.Hypot(alongX, alongY) < 1e-12 {
		return frames
	}

	angle := math.Atan2(alongX, alongY)
	for i := range frames {
		frames[i] = frames[i].Rolled(angle)
	}
	return frames
}

// Along is how far each frame lies along the path, from 0 at the first to 1
// at the end; a closed path's last frame stops short of 1.
func Along(frames []PathFrame, closed bool) []float64 {
	walked := FrameDistances(frames)
	if len(walked) == 0 {
		return walked
	}
	total := walked[len(walked)-1]
	if closed {
		total += frames[len(frames)-1].Origin.Distance(frames[0].Origin)
	}
	if total > 0 {
		for i := range walked {
			walked[i] /= total
		}
	}
	return walked
}

func FrameTRSs(frames []PathFrame) []trs.TRS {
	out := make([]trs.TRS, len(frames))
	for i, f := range frames {
		out[i] = f.TRS()
	}
	return out
}

type PathFramesNode struct {
	Path   nodes.LiftedPort[[]vector3.Float64]
	Closed nodes.Output[bool] `description:"Joins the last point back to the first."`
}

func (PathFramesNode) Description() string {
	return "Transforms along a path that turn with it without twisting."
}

func (PathFramesNode) FramesDescription() string {
	return "Travel along +Z, up along +Y; sharp corners get two."
}

func (PathFramesNode) AlongDescription() string {
	return "Each frame's fraction of the way along, 0 to 1."
}

func (n PathFramesNode) Frames(out *nodes.Lifted[[]trs.TRS]) {
	closed := nodes.GetOutputValue(out, n.Closed)
	nodes.Zip1(out, n.Path, func(path []vector3.Float64) []trs.TRS {
		return FrameTRSs(PathFrames(withoutRepeats(path, closed), closed))
	})
}

func (n PathFramesNode) Along(out *nodes.Lifted[[]float64]) {
	closed := nodes.GetOutputValue(out, n.Closed)
	nodes.Zip1(out, n.Path, func(path []vector3.Float64) []float64 {
		return Along(PathFrames(withoutRepeats(path, closed), closed), closed)
	})
}

// A point repeating its predecessor, or in a loop the first point repeated at
// the end, adds a segment of no length.
func withoutRepeats(points []vector3.Float64, loop bool) []vector3.Float64 {
	kept := make([]vector3.Float64, 0, len(points))
	for _, p := range points {
		if len(kept) > 0 && p == kept[len(kept)-1] {
			continue
		}
		kept = append(kept, p)
	}
	for loop && len(kept) > 1 && kept[len(kept)-1] == kept[0] {
		kept = kept[:len(kept)-1]
	}
	return kept
}
