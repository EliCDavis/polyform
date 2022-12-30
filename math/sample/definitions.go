package sample

import "github.com/EliCDavis/vector"

// Floats
type FloatToFloat func(float64) float64
type FloatToVec2 func(float64) vector.Vector2
type FloatToVec3 func(float64) vector.Vector3

// Vector 2
type Vec2ToFloat func(vector.Vector2) float64
type Vec2ToVec2 func(vector.Vector2) vector.Vector2
type Vec2ToVec3 func(vector.Vector2) vector.Vector3

// Vector 3
type Vec3ToFloat func(vector.Vector3) float64
type Vec3ToVec2 func(vector.Vector3) vector.Vector2
type Vec3ToVec3 func(vector.Vector3) vector.Vector3
