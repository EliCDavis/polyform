package animation

import "github.com/EliCDavis/vector/vector3"

type Frame[T any] struct {
	time float64
	val  T
}

func NewFrame[T any](time float64, val T) Frame[T] {
	return Frame[T]{
		time: time,
		val:  val,
	}
}

func (s Frame[T]) Time() float64 {
	return s.time
}

func (s Frame[T]) Val() T {
	return s.val
}

func UniformFrames[T any](data []T, time float64) []Frame[T] {
	if len(data) == 0 {
		return nil
	}

	if len(data) == 1 {
		return []Frame[T]{NewFrame(0, data[0])}
	}

	frames := make([]Frame[T], len(data))
	timeStep := time / float64(len(data)-1)
	for i, v := range data {
		frames[i] = NewFrame(timeStep*float64(i), v)
	}

	return frames
}

// type Sequence interface{}

type Sequence struct {
	joint  string
	frames []Frame[vector3.Float64]
}

func (s Sequence) Frames() []Frame[vector3.Float64] {
	return s.frames
}

func (s Sequence) Joint() string {
	return s.joint
}

func NewSequence(joint string, frames []Frame[vector3.Float64]) Sequence {
	return Sequence{
		joint:  joint,
		frames: frames,
	}
}
