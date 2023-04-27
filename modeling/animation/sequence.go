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
