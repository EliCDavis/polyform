package sample

import "github.com/EliCDavis/vector"

func ComposeFloat(f2fs ...FloatToFloat) FloatToFloat {
	if len(f2fs) == 0 {
		panic("can not create a composition without any functions")
	}

	if len(f2fs) == 1 {
		return f2fs[0]
	}

	return func(f float64) float64 {
		result := f2fs[0](f)
		for i := 1; i < len(f2fs); i++ {
			result = f2fs[i](result)
		}
		return result
	}
}

func ComposeVec2(f2fs ...Vec2ToVec2) Vec2ToVec2 {
	if len(f2fs) == 0 {
		panic("can not create a composition without any functions")
	}

	if len(f2fs) == 1 {
		return f2fs[0]
	}

	return func(f vector.Vector2) vector.Vector2 {
		result := f2fs[0](f)
		for i := 1; i < len(f2fs); i++ {
			result = f2fs[i](result)
		}
		return result
	}
}

func ComposeVec3(f2fs ...Vec3ToVec3) Vec3ToVec3 {
	if len(f2fs) == 0 {
		panic("can not create a composition without any functions")
	}

	if len(f2fs) == 1 {
		return f2fs[0]
	}

	return func(f vector.Vector3) vector.Vector3 {
		result := f2fs[0](f)
		for i := 1; i < len(f2fs); i++ {
			result = f2fs[i](result)
		}
		return result
	}
}
