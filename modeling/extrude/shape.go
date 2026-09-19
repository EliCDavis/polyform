package extrude

import (
	"math"

	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
)

type ringFrame struct {
	dir, per vector3.Float64
}

// Everything else in the library lays a 2D shape out as (x, 0, y), so a path
// running straight up has to put outline x on world x and outline y on world
// z. Perpendicular answers -z for that direction, which spins the outline
// half a turn about the sweep and leaves it mirrored against the points the
// caller drew.
func seedPerpendicular(dir vector3.Float64) vector3.Float64 {
	for _, axis := range []vector3.Float64{
		vector3.Forward[float64](),
		vector3.Up[float64](),
	} {
		per := axis.Sub(dir.Scale(axis.Dot(dir)))
		if per.Length() > 1e-9 {
			return per.Normalized()
		}
	}
	return dir.Perpendicular()
}

// The frame is carried from one path point to the next by the rotation
// between their directions. Deriving per from the cross product of
// neighbouring segments instead collapses to zero the moment the path runs
// straight, and picking a fresh perpendicular per point twists the shell
// wherever a straight run meets a bend.
//
// Caps have to sit on the same frame their shell ring was built from or they
// leave a crack, so callers share this rather than recomputing it.
func pathFrames(path []vector3.Float64, closed bool) []ringFrame {
	directions := make([]vector3.Float64, len(path))
	for i := range path {
		previous := (i + len(path) - 1) % len(path)
		next := (i + 1) % len(path)

		var dir vector3.Float64
		switch {
		case !closed && i == 0:
			dir = path[next].Sub(path[i])
		case !closed && i == len(path)-1:
			dir = path[i].Sub(path[previous])
		default:
			dir = path[next].Sub(path[i]).Add(path[i].Sub(path[previous]))
		}
		directions[i] = dir.Normalized()
	}

	frames := make([]ringFrame, len(path))
	per := seedPerpendicular(directions[0])

	for i, dir := range directions {
		if i > 0 {
			per = quaternion.RotationTo(directions[i-1], dir).Rotate(per)
		}

		// Rotations accumulate error, so square per back up against dir
		// rather than letting the frame shear over a long path.
		per = per.Sub(dir.Scale(per.Dot(dir)))
		if per.Length() == 0 {
			per = dir.Perpendicular()
		}

		frames[i] = ringFrame{dir: dir, per: per.Normalized()}
	}

	if closed {
		spreadSeamTwist(frames)
	}

	return frames
}

// Transporting a frame around a closed loop does not generally bring it back
// to where it started, and the whole discrepancy piles up on the one segment
// joining the last ring to the first. Sharing it out leaves every segment
// slightly twisted instead of one segment wrung right round.
func spreadSeamTwist(frames []ringFrame) {
	last := len(frames) - 1
	if last < 1 {
		return
	}

	carried := quaternion.RotationTo(frames[last].dir, frames[0].dir).Rotate(frames[last].per)
	mismatch := math.Atan2(
		frames[0].per.Cross(carried).Dot(frames[0].dir),
		frames[0].per.Dot(carried),
	)
	if mismatch == 0 {
		return
	}

	share := mismatch / float64(len(frames))
	for i := range frames {
		frames[i].per = quaternion.
			FromTheta(-share*float64(i), frames[i].dir).
			Rotate(frames[i].per)
	}
}

func makeShape(shape []vector2.Float64, path []vector3.Float64, close bool) modeling.Mesh {
	if len(path) < 2 {
		panic("Can not extrude a path with less than 2 points")
	}

	vertices := make([]vector3.Float64, 0, len(path)*len(shape))
	normals := make([]vector3.Float64, 0, len(path)*len(shape))
	frames := pathFrames(path, close)
	for i, p := range path {
		verts, norms := ProjectFace(p, frames[i].dir, frames[i].per, shape)
		vertices = append(vertices, verts...)
		normals = append(normals, norms...)
	}

	sides := len(shape)

	tris := make([]int, 0)

	for pathIndex := range path {
		bottom := pathIndex * sides
		top := (pathIndex + 1) * sides
		if pathIndex == len(path)-1 {
			if close {
				top = 0
			} else {
				continue
			}
		}

		for sideIndex := 0; sideIndex < sides; sideIndex++ {
			topRight := top + sideIndex
			bottomRight := bottom + sideIndex

			topLeft := topRight - 1
			bottomLeft := bottomRight - 1
			if sideIndex == 0 {
				topLeft = top + sides - 1
				bottomLeft = bottom + sides - 1
			}

			tris = append(
				tris,

				bottomLeft,
				topLeft,
				topRight,

				bottomLeft,
				topRight,
				bottomRight,
			)
		}
	}

	return modeling.NewTriangleMesh(tris).
		SetFloat3Data(map[string][]vector3.Float64{
			modeling.PositionAttribute: vertices,
			modeling.NormalAttribute:   normals,
		})
}

func Shape(shape []vector2.Float64, path []vector3.Float64) modeling.Mesh {
	return makeShape(shape, path, false)
}

func ClosedShape(shape []vector2.Float64, path []vector3.Float64) modeling.Mesh {
	return makeShape(shape, path, true)
}
