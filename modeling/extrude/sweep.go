package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/triangulation"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

// sweep builds a profile's shell along a set of frames. The profile has to be
// counter-clockwise: every face then points out without being checked.
type sweep struct {
	profile []vector2.Float64
	frames  []frame
	closed  bool

	// Indexed by frame.point; nil leaves the profile at its own size.
	scales []float64

	// Outline and path corners turning at least this many radians keep a
	// hard edge.
	smoothing float64

	// Nil writes no texture coordinates.
	uv func(point int, across, along float64) vector2.Float64
}

var allSmooth = math.Inf(1)

func (s sweep) scale(point int) float64 {
	if s.scales == nil {
		return 1
	}
	return s.scales[point]
}

func (s sweep) hard(f frame) bool {
	if f.incoming == (vector3.Float64{}) || f.outgoing == (vector3.Float64{}) {
		return false
	}
	return f.incoming.Angle(f.outgoing) >= s.smoothing-1e-9
}

func (s sweep) along() []float64 {
	walked := distances(s.frames)
	total := walked[len(walked)-1]
	if s.closed {
		total += s.frames[len(s.frames)-1].origin.Distance(s.frames[0].origin)
	}
	if total > 0 {
		for i := range walked {
			walked[i] /= total
		}
	}
	return walked
}

func (s sweep) shell() modeling.Mesh {
	slots, start, end := outlineSlots(s.profile, s.smoothing, s.uv != nil)
	var across []float64
	if s.uv != nil {
		across = perimeterFractions(s.profile, slots, end[len(s.profile)-1])
	}
	along := s.along()

	var vertices, normals []vector3.Float64
	var uvs []vector2.Float64
	ring := func(f, shading frame, along float64) int {
		base := len(vertices)
		scale := s.scale(f.point)
		for k, slot := range slots {
			vertices = append(vertices, f.place(s.profile[slot.point].Scale(scale)))
			normals = append(normals, shading.direction(slot.normal).Normalized())
			if s.uv != nil {
				uvs = append(uvs, s.uv(f.point, across[k], along))
			}
		}
		return base
	}

	// A hard corner gets a ring per side, shaded square to that side's
	// segment.
	before := make([]int, len(s.frames))
	after := make([]int, len(s.frames))
	for i, f := range s.frames {
		if s.hard(f) {
			before[i] = ring(f, f.facing(f.incoming), along[i])
			after[i] = ring(f, f.facing(f.outgoing), along[i])
			continue
		}
		before[i] = ring(f, f, along[i])
		after[i] = before[i]
	}

	segments := len(s.frames) - 1
	if s.closed {
		segments = len(s.frames)
		if s.uv != nil {
			first := s.frames[0]
			shading := first
			if s.hard(first) {
				shading = first.facing(first.incoming)
			}
			before[0] = ring(first, shading, 1)
		}
	}

	tris := make([]int, 0, segments*len(s.profile)*6)
	for i := 0; i < segments; i++ {
		lower, upper := after[i], before[(i+1)%len(s.frames)]
		for j := range s.profile {
			a, b := lower+start[j], lower+end[j]
			c, d := upper+start[j], upper+end[j]
			tris = append(tris, a, c, d, a, d, b)
		}
	}

	mesh := modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		})
	if s.uv == nil {
		return mesh
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

// cap closes the sweep at f. A counter-clockwise triangle placed by a frame
// faces back along its tangent, so the start cap keeps its winding and the
// end cap reverses it.
func (s sweep) cap(region triangulation.Region, f frame, end bool, uv func(vector2.Float64) vector2.Float64) modeling.Mesh {
	normal := f.tangent.Scale(-1)
	if end {
		normal = f.tangent
	}

	scale := s.scale(f.point)
	vertices := make([]vector3.Float64, len(region.Points))
	normals := make([]vector3.Float64, len(region.Points))
	for i, p := range region.Points {
		vertices[i] = f.place(p.Scale(scale))
		normals[i] = normal
	}

	tris := make([]int, 0, len(region.Triangles)*3)
	for _, tri := range region.Triangles {
		if end {
			tris = append(tris, tri[0], tri[2], tri[1])
		} else {
			tris = append(tris, tri[0], tri[1], tri[2])
		}
	}

	mesh := modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		})
	if uv == nil {
		return mesh
	}
	uvs := make([]vector2.Float64, len(region.Points))
	for i, p := range region.Points {
		uvs[i] = uv(p)
	}
	return mesh.SetFloat2Attribute(modeling.TexCoordAttribute, uvs)
}

type outlineSlot struct {
	point  int
	normal vector2.Float64
}

// A ring only needs two vertices at an outline corner where the outline
// actually turns one. Splitting every vertex leaves a spline sampled outline
// carrying twice the vertices it needs and a seam of split normals at every
// sample, which reads as faceting rather than a smooth sweep.
//
// splitSeam gives the loop's closing point two slots even where the outline
// runs smoothly through it, so one can carry perimeter 0 and the other 1.
// Both keep the averaged normal, so the split shows in the UVs and not in
// the shading.
func outlineSlots(shape []vector2.Float64, smoothing float64, splitSeam bool) (slots []outlineSlot, start, end []int) {
	count := len(shape)
	edges := make([]vector2.Float64, count)
	for j := range shape {
		along := shape[(j+1)%count].Sub(shape[j])
		if along.Length() == 0 {
			continue
		}
		edges[j] = vector2.New(along.Y(), -along.X()).Normalized()
	}

	start = make([]int, count)
	end = make([]int, count)

	for j := range shape {
		previous := (j + count - 1) % count
		incoming, outgoing := edges[previous], edges[j]

		switch {
		case incoming.Length() == 0:
			incoming = outgoing
		case outgoing.Length() == 0:
			outgoing = incoming
		}

		// Nudged so a shape whose corners all sit exactly on the threshold,
		// a regular octagon against the 45 degree default, resolves the same
		// way at every one of them instead of splitting half of them.
		if incoming.Length() > 0 && incoming.Angle(outgoing) < smoothing-1e-9 {
			averaged := incoming.Add(outgoing).Normalized()
			if splitSeam && j == 0 {
				slots = append(slots, outlineSlot{j, averaged})
				end[previous] = len(slots) - 1
				slots = append(slots, outlineSlot{j, averaged})
				start[j] = len(slots) - 1
				continue
			}
			slots = append(slots, outlineSlot{j, averaged})
			start[j] = len(slots) - 1
			end[previous] = len(slots) - 1
			continue
		}

		slots = append(slots, outlineSlot{j, incoming})
		end[previous] = len(slots) - 1
		slots = append(slots, outlineSlot{j, outgoing})
		start[j] = len(slots) - 1
	}

	return
}

func perimeterFractions(shape []vector2.Float64, slots []outlineSlot, seam int) []float64 {
	count := len(shape)
	walked := make([]float64, count+1)
	for j := 0; j < count; j++ {
		walked[j+1] = walked[j] + shape[(j+1)%count].Sub(shape[j]).Length()
	}

	fractions := make([]float64, len(slots))
	if perimeter := walked[count]; perimeter > 0 {
		for i, slot := range slots {
			fractions[i] = walked[slot.point] / perimeter
		}
		if seam >= 0 && seam < len(fractions) {
			fractions[seam] = 1
		}
	}
	return fractions
}

func counterClockwise(shape []vector2.Float64) []vector2.Float64 {
	if geometry.Shape(shape).SignedArea() >= 0 {
		return shape
	}

	flipped := make([]vector2.Float64, len(shape))
	for i, p := range shape {
		flipped[len(shape)-1-i] = p
	}
	return flipped
}

// A point repeating its predecessor, or in a loop the first point repeated at
// the end to close it, adds an edge of no length: a zero-area quad in the
// shell, and a NaN normal where two such edges meet.
func withoutRepeats[T comparable](points []T, loop bool) []T {
	kept := make([]T, 0, len(points))
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
