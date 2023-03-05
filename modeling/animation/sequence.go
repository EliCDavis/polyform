package animation

import "github.com/EliCDavis/vector/vector3"

type Frame struct {
	time float64
	val  vector3.Float64
}

func NewFrame(time float64, val vector3.Float64) Frame {
	return Frame{
		time: time,
		val:  val,
	}
}

func (s Frame) Time() float64 {
	return s.time
}

func (s Frame) Val() vector3.Float64 {
	return s.val
}

type Sequence struct {
	joint  string
	frames []Frame
}

func (s Sequence) Frames() []Frame {
	return s.frames
}

func (s Sequence) Joint() string {
	return s.joint
}

func NewSequence(joint string, frames []Frame) Sequence {
	return Sequence{
		joint:  joint,
		frames: frames,
	}
}
