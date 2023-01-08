package sample

import "github.com/EliCDavis/vector"

// ============================================================================
// Floats
// ============================================================================
type FloatToFloat func(float64) float64

func (v1tov1 FloatToFloat) Add(other FloatToFloat) FloatToFloat {
	return func(v float64) float64 {
		return v1tov1(v) + other(v)
	}
}

func (v1tov1 FloatToFloat) Sub(other FloatToFloat) FloatToFloat {
	return func(v float64) float64 {
		return v1tov1(v) - other(v)
	}
}

func (v1tov1 FloatToFloat) Multiply(other FloatToFloat) FloatToFloat {
	return func(v float64) float64 {
		return v1tov1(v) * other(v)
	}
}

func (v1tov1 FloatToFloat) Scale(other float64) FloatToFloat {
	return func(v float64) float64 {
		return v1tov1(v) * other
	}
}

type FloatToVec2 func(float64) vector.Vector2

func (v1tov2 FloatToVec2) Add(other FloatToVec2) FloatToVec2 {
	return func(v float64) vector.Vector2 {
		return v1tov2(v).Add(other(v))
	}
}

func (v1tov2 FloatToVec2) Sub(other FloatToVec2) FloatToVec2 {
	return func(v float64) vector.Vector2 {
		return v1tov2(v).Sub(other(v))
	}
}

type FloatToVec3 func(float64) vector.Vector3

func (v1tov3 FloatToVec3) Add(other FloatToVec3) FloatToVec3 {
	return func(v float64) vector.Vector3 {
		return v1tov3(v).Add(other(v))
	}
}

func (v1tov3 FloatToVec3) Sub(other FloatToVec3) FloatToVec3 {
	return func(v float64) vector.Vector3 {
		return v1tov3(v).Sub(other(v))
	}
}

// ============================================================================
// Vector 2
// ============================================================================
type Vec2ToFloat func(vector.Vector2) float64

func (v2tov1 Vec2ToFloat) Add(other Vec2ToFloat) Vec2ToFloat {
	return func(v vector.Vector2) float64 {
		return v2tov1(v) + other(v)
	}
}

func (v2tov1 Vec2ToFloat) Sub(other Vec2ToFloat) Vec2ToFloat {
	return func(v vector.Vector2) float64 {
		return v2tov1(v) - other(v)
	}
}

func (v2tov1 Vec2ToFloat) Multiply(other Vec2ToFloat) Vec2ToFloat {
	return func(v vector.Vector2) float64 {
		return v2tov1(v) * other(v)
	}
}

func (v2tov1 Vec2ToFloat) Scale(other float64) Vec2ToFloat {
	return func(v vector.Vector2) float64 {
		return v2tov1(v) * other
	}
}

type Vec2ToVec2 func(vector.Vector2) vector.Vector2

func (v2tov2 Vec2ToVec2) Add(other Vec2ToVec2) Vec2ToVec2 {
	return func(v vector.Vector2) vector.Vector2 {
		return v2tov2(v).Add(other(v))
	}
}

func (v2tov2 Vec2ToVec2) Sub(other Vec2ToVec2) Vec2ToVec2 {
	return func(v vector.Vector2) vector.Vector2 {
		return v2tov2(v).Sub(other(v))
	}
}

type Vec2ToVec3 func(vector.Vector2) vector.Vector3

func (v2tov3 Vec2ToVec3) Add(other Vec2ToVec3) Vec2ToVec3 {
	return func(v vector.Vector2) vector.Vector3 {
		return v2tov3(v).Add(other(v))
	}
}

func (v2tov3 Vec2ToVec3) Sub(other Vec2ToVec3) Vec2ToVec3 {
	return func(v vector.Vector2) vector.Vector3 {
		return v2tov3(v).Sub(other(v))
	}
}

// ============================================================================
// Vector 3
// ============================================================================
type Vec3ToFloat func(vector.Vector3) float64

func (v3tov1 Vec3ToFloat) Add(other Vec3ToFloat) Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return v3tov1(v) + other(v)
	}
}

func (v3tov1 Vec3ToFloat) Sub(other Vec3ToFloat) Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return v3tov1(v) - other(v)
	}
}

func (v3tov1 Vec3ToFloat) Multiply(other Vec3ToFloat) Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return v3tov1(v) * other(v)
	}
}

func (v3tov1 Vec3ToFloat) Scale(other float64) Vec3ToFloat {
	return func(v vector.Vector3) float64 {
		return v3tov1(v) * other
	}
}

type Vec3ToVec2 func(vector.Vector3) vector.Vector2

func (v3tov2 Vec3ToVec2) Add(other Vec3ToVec2) Vec3ToVec2 {
	return func(v vector.Vector3) vector.Vector2 {
		return v3tov2(v).Add(other(v))
	}
}

func (v3tov2 Vec3ToVec2) Sub(other Vec3ToVec2) Vec3ToVec2 {
	return func(v vector.Vector3) vector.Vector2 {
		return v3tov2(v).Sub(other(v))
	}
}

type Vec3ToVec3 func(vector.Vector3) vector.Vector3

func (v3tov3 Vec3ToVec3) Add(other Vec3ToVec3) Vec3ToVec3 {
	return func(v vector.Vector3) vector.Vector3 {
		return v3tov3(v).Add(other(v))
	}
}

func (v3tov3 Vec3ToVec3) Sub(other Vec3ToVec3) Vec3ToVec3 {
	return func(v vector.Vector3) vector.Vector3 {
		return v3tov3(v).Sub(other(v))
	}
}
