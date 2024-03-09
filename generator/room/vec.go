package room

import "github.com/EliCDavis/vector"

// structs with exposed fields for sake of binary serialization

type Vec3[T vector.Number] struct {
	X T
	Y T
	Z T
}

type Vec4[T vector.Number] struct {
	X T
	Y T
	Z T
	W T
}
