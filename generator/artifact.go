package generator

import "io"

type Artifact interface {
	Write(io.Writer) error
	Mime() string
}

type PolyformArtifact[T any] interface {
	Artifact
	Value() T
}
