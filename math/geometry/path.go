package geometry

import (
	"github.com/EliCDavis/polyform/math/quaternion"
	"github.com/EliCDavis/polyform/math/trs"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
)

func PathDistances(path []vector3.Float64) []float64 {
	distances := make([]float64, len(path))
	for i := 1; i < len(path); i++ {
		distances[i] = distances[i-1] + path[i].Distance(path[i-1])
	}
	return distances
}

type PathLengthNode struct {
	Path nodes.LiftedPort[[]vector3.Float64]
}

func (PathLengthNode) Description() string {
	return "Length of a path of straight segments."
}

func (PathLengthNode) DistancesDescription() string {
	return "How far along the path each point lies."
}

func (n PathLengthNode) Length(out *nodes.Lifted[float64]) {
	nodes.Zip1(out, n.Path, func(path []vector3.Float64) float64 {
		distances := PathDistances(path)
		if len(distances) == 0 {
			return 0
		}
		return distances[len(distances)-1]
	})
}

func (n PathLengthNode) Distances(out *nodes.Lifted[[]float64]) {
	nodes.Zip1(out, n.Path, PathDistances)
}

type PathSample struct {
	Position vector3.Float64
	Tangent  vector3.Float64
	Segment  int
}

func SamplePath(path []vector3.Float64, distance float64) PathSample {
	switch len(path) {
	case 0:
		return PathSample{}
	case 1:
		return PathSample{Position: path[0]}
	}

	last := len(path) - 2
	travelled := 0.
	var tangent vector3.Float64
	for i := 0; i <= last; i++ {
		start, end := path[i], path[i+1]
		length := end.Distance(start)
		if length == 0 && i < last {
			continue
		}
		if length > 0 {
			tangent = end.Sub(start).Normalized()
		}
		if distance < travelled+length || i == last {
			t := 0.
			if length > 0 {
				t = max(0, min(1, (distance-travelled)/length))
			}
			return PathSample{
				Position: start.Add(end.Sub(start).Scale(t)),
				Tangent:  tangent,
				Segment:  i,
			}
		}
		travelled += length
	}
	return PathSample{}
}

type SamplePathNode struct {
	Path     nodes.LiftedPort[[]vector3.Float64]
	Distance nodes.LiftedPort[float64] `description:"How far along the path, clamped to its ends."`
}

func (SamplePathNode) Description() string {
	return "The point a distance along a path of straight segments."
}

func (SamplePathNode) TangentDescription() string {
	return "Direction of travel at the point."
}

func (SamplePathNode) TransformDescription() string {
	return "At the point, facing along the path."
}

func (SamplePathNode) SegmentDescription() string {
	return "Index of the segment the point lies on."
}

func (n SamplePathNode) Position(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(out, n.Path, n.Distance, func(path []vector3.Float64, d float64) vector3.Float64 {
		return SamplePath(path, d).Position
	})
}

func (n SamplePathNode) Tangent(out *nodes.Lifted[vector3.Float64]) {
	nodes.Zip2(out, n.Path, n.Distance, func(path []vector3.Float64, d float64) vector3.Float64 {
		return SamplePath(path, d).Tangent
	})
}

func (n SamplePathNode) Transform(out *nodes.Lifted[trs.TRS]) {
	nodes.Zip2(out, n.Path, n.Distance, func(path []vector3.Float64, d float64) trs.TRS {
		sample := SamplePath(path, d)
		rotation := quaternion.Identity()
		if sample.Tangent != (vector3.Float64{}) {
			rotation = quaternion.RotationTo(vector3.Forward[float64](), sample.Tangent)
		}
		return trs.New(sample.Position, rotation, vector3.One[float64]())
	})
}

func (n SamplePathNode) Segment(out *nodes.Lifted[int]) {
	nodes.Zip2(out, n.Path, n.Distance, func(path []vector3.Float64, d float64) int {
		return SamplePath(path, d).Segment
	})
}
