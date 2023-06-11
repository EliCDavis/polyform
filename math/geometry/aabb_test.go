package geometry_test

import (
	"encoding/json"
	"testing"

	"github.com/EliCDavis/polyform/math/geometry"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestAABBToJSON(t *testing.T) {
	center := vector3.New(1., 2., 3.)
	size := vector3.New(4., 5., 6.)
	in := geometry.NewAABB(center, size)
	out := geometry.AABB{}

	marshalledData, marshallErr := json.Marshal(in)
	unmarshallErr := json.Unmarshal(marshalledData, &out)

	assert.NoError(t, marshallErr)
	assert.NoError(t, unmarshallErr)
	assert.Equal(t, "{\"center\":{\"x\":1,\"y\":2,\"z\":3},\"extents\":{\"x\":2,\"y\":2.5,\"z\":3}}", string(marshalledData))
	assert.Equal(t, center, out.Center())
	assert.Equal(t, size, out.Size())
}

func TestAABB(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.One[float64]())

	assert.Equal(t, center, aabb.Center())
	assert.True(t, aabb.Contains(center))
	assert.False(t, aabb.Contains(vector3.One[float64]()))
	assert.False(t, aabb.Contains(vector3.Zero[float64]()))
	assert.False(t, aabb.Contains(vector3.Right[float64]()))
}

func TestAABBClosesetPoint(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.Fill(2.))

	tests := map[string]struct {
		input vector3.Float64
		want  vector3.Float64
	}{
		"center": {input: vector3.New(2., 3., 4.), want: vector3.New(2., 3., 4.)},
		"left":   {input: vector3.New(0., 3., 4.), want: vector3.New(1., 3., 4.)},
		"right":  {input: vector3.New(4., 3., 4.), want: vector3.New(3., 3., 4.)},
		"up":     {input: vector3.New(2., 5., 4.), want: vector3.New(2., 4., 4.)},
		"down":   {input: vector3.New(2., 1., 4.), want: vector3.New(2., 2., 4.)},
		"fwd":    {input: vector3.New(2., 3., 6.), want: vector3.New(2., 3., 5.)},
		"back":   {input: vector3.New(2., 3., 2.), want: vector3.New(2., 3., 3.)},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := aabb.ClosestPoint(tc.input)
			assert.InDelta(t, tc.want.X(), got.X(), 0.00000001)
			assert.InDelta(t, tc.want.Y(), got.Y(), 0.00000001)
			assert.InDelta(t, tc.want.Z(), got.Z(), 0.00000001)
		})
	}
}

func TestAABBIntersectsRayInRange(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.Fill(2.))

	tests := map[string]struct {
		ray      geometry.Ray
		min, max float64
		want     bool
	}{
		"up at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered under aabb": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered under aabb stop short": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered under aabb skips over": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., 1., 0.)),
			min:  5,
			max:  10,
			want: false,
		},

		"down at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered over aabb": {
			ray:  geometry.NewRay(vector3.New(2., 10., 4.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered over aabb stop short": {
			ray:  geometry.NewRay(vector3.New(2., 10., 4.), vector3.New(0., -1., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered over aabb skips over": {
			ray:  geometry.NewRay(vector3.New(2., 0., 4.), vector3.New(0., -1., 0.)),
			min:  9,
			max:  10,
			want: false,
		},

		"right at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered left of origin": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered left of origin stop short": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered left of origin skips over": {
			ray:  geometry.NewRay(vector3.New(0., 3., 4.), vector3.New(1., 0., 0.)),
			min:  5,
			max:  10,
			want: false,
		},

		"forward at origin": {
			ray:  geometry.NewRay(vector3.New(0., 0., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  10,
			want: false,
		},
		"centered forward of origin": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  10,
			want: true,
		},
		"centered forward of origin stop short": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  0,
			max:  0.5,
			want: false,
		},
		"centered forward of origin skips over": {
			ray:  geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.)),
			min:  6,
			max:  10,
			want: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			got := aabb.IntersectsRayInRange(tc.ray, tc.min, tc.max)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestRayInRangeFlatBox(t *testing.T) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.New(2., 2., 0.))
	ray := geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.))

	assert.True(t, aabb.IntersectsRayInRange(ray, 0, 10))
}

func TestRayInRangeBunny(t *testing.T) {
	center := vector3.New(
		-1.2898899242281914,
		2.076879933476448,
		-0.005710022523999214,
	)
	extents := vector3.New(
		0.010519996285438538,
		0.014499947428703308,
		0.0076299998909235,
	)
	origin := vector3.New(
		7.235200983807349,
		-0.04900531242295969,
		6.757225025788709,
	)
	direction := vector3.New(
		-0.16930325604861363,
		0.8832115994797817,
		-0.43734846293968305,
	)

	aabb := geometry.NewAABB(center, extents.Scale(2))
	ray := geometry.NewRay(origin, direction)

	assert.False(t, aabb.IntersectsRayInRange(ray, 0, 10))
}

func TestRayInRangeBunny3(t *testing.T) {
	center := vector3.New(
		0.004800017923116684,
		2.354609936475754,
		0.12080997927114367,
	)
	extents := vector3.New(
		0.016750004142522812,
		0.00406995415687561,
		0.01293000066652894,
	)
	origin := vector3.New(
		7.0000217719758595,
		-0.04728545371962589,
		6.750434557639183,
	)
	direction := vector3.New(
		-0.7762940267358865,
		0.6268664062542054,
		-0.0663784058570281,
	)

	aabb := geometry.NewAABB(center, extents.Scale(2))
	ray := geometry.NewRay(origin, direction)

	assert.False(t, aabb.IntersectsRayInRange(ray, 0.001, 10000))
}

func TestRayInRangeBunny2(t *testing.T) {
	box := geometry.NewAABBFromPoints(
		vector3.New(-0.011949986219406128, 2.350539982318878, 0.12459998019039631),
		vector3.New(0.012929998338222504, 2.3586798906326294, 0.13373997993767262),
		vector3.New(0.021550022065639496, 2.3567200899124146, 0.10787997860461473),
	)
	origin := vector3.New(
		7.066985679097499,
		-0.04755005707842219,
		6.7197891752892325,
	)
	direction := vector3.New(
		-0.619907087159697,
		0.5983439767488956,
		-0.5076412993221661,
	)

	ray := geometry.NewRay(origin, direction)

	assert.False(t, box.IntersectsRayInRange(ray, 0, 10))
}

func TestAABBContains(t *testing.T) {
	tests := map[string]struct {
		center vector3.Float64
		size   vector3.Float64
		point  vector3.Float64
		result bool
	}{
		"simple": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.Zero[float64](),
			result: true,
		},
		"on corner": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0.5, 0.5, 0.5),
			result: true,
		},
		"on edge": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0.5, 0.5),
			result: true,
		},
		"on face": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., 0.5),
			result: true,
		},
		"0 volume": {
			center: vector3.Zero[float64](),
			size:   vector3.Zero[float64](),
			point:  vector3.New(0., 0., 0.),
			result: true,
		},
		"outside corner": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0.500001, 0.500001, 0.500001),
			result: false,
		},
		"outside edge": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0.500001, 0.500001),
			result: false,
		},
		"outside face": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., 0.500001),
			result: false,
		},

		"outside corner 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(-0.500001, -0.500001, -0.500001),
			result: false,
		},
		"outside edge 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., -0.500001, -0.500001),
			result: false,
		},
		"outside face 2": {
			center: vector3.Zero[float64](),
			size:   vector3.One[float64](),
			point:  vector3.New(0., 0., -0.500001),
			result: false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			aabb := geometry.NewAABB(tc.center, tc.size)
			assert.Equal(t, tc.result, aabb.Contains(tc.point))
		})
	}
}

func TestAABBExpand(t *testing.T) {
	// ARRANGE ================================================================
	aabb := geometry.NewAABB(vector3.Zero[float64](), vector3.One[float64]())

	// ACT ====================================================================
	aabb.Expand(2)

	// ASSERT =================================================================
	assert.Equal(t, vector3.New(3., 3., 3.), aabb.Size())
}

func TestAABBEncapsulatePoints(t *testing.T) {
	aabb := geometry.NewEmptyAABB()

	aabb.EncapsulatePoint(vector3.New(-0.5, 0., 0.))
	aabb.EncapsulatePoint(vector3.New(0.5, 0., 0.))

	aabb.EncapsulatePoint(vector3.New(0., 0.5, 0.))
	aabb.EncapsulatePoint(vector3.New(0., -0.5, 0.))

	aabb.EncapsulatePoint(vector3.New(0., 0., 0.5))
	aabb.EncapsulatePoint(vector3.New(0., 0., -0.5))

	assert.Equal(t, vector3.Zero[float64](), aabb.Center())
	assert.Equal(t, vector3.One[float64](), aabb.Size())
}

func TestAABBFromPoints(t *testing.T) {
	aabb := geometry.NewAABBFromPoints(
		vector3.New(-0.5, 0., 0.),
		vector3.New(0.5, 0., 0.),
		vector3.New(0., 0.5, 0.),
		vector3.New(0., -0.5, 0.),
		vector3.New(0., 0., 0.5),
		vector3.New(0., 0., -0.5),
	)

	assert.Equal(t, vector3.Zero[float64](), aabb.Center())
	assert.Equal(t, vector3.One[float64](), aabb.Size())
}

var final bool

func BenchmarkAABBIntersectsRayInRange(b *testing.B) {
	center := vector3.New(2., 3., 4.)
	aabb := geometry.NewAABB(center, vector3.Fill(2.))

	ray := geometry.NewRay(vector3.New(2., 3., 0.), vector3.New(0., 0., 1.))
	min := 0.
	max := 10.

	got := false
	for i := 0; i < b.N; i++ {
		got = aabb.IntersectsRayInRange(ray, min, max)
	}
	final = got
}

func BenchmarkAABBIntersectsRayInRangeOther(b *testing.B) {
	aabb := arrAABB{
		min: [3]float64{1, 2, 3},
		max: [3]float64{3, 4, 5},
	}

	ray := arrRay{
		dir: [3]float64{2., 3., 0.},
		org: [3]float64{0., 0., 1.},
	}
	min := 0.
	max := 10.

	got := false
	for i := 0; i < b.N; i++ {
		got = aabb.IntersectsRayInRange(ray, min, max)
	}
	final = got
}

type arrAABB struct {
	min [3]float64
	max [3]float64
}

type arrRay struct {
	dir [3]float64
	org [3]float64
}

// IntersectsRayInRange determines whether or not a ray intersects the bounding
// box with the range of the ray between min and max
//
// Intersection method by Andrew Kensler at Pixar, found in the book "Ray
// Tracing The Next Week" by Peter Shirley
func (aabb arrAABB) IntersectsRayInRange(ray arrRay, min, max float64) bool {
	for a := 0; a < 3; a++ {
		invD := 1.0 / ray.dir[a]
		t0 := (aabb.min[a] - ray.org[a]) * invD
		t1 := (aabb.max[a] - ray.org[a]) * invD
		if invD < 0.0 {
			// std::swap(t0, t1);
			t1, t0 = t0, t1
		}

		if t0 > min {
			min = t0
		}

		if t1 < max {
			max = t1
		}

		if max <= min {
			return false
		}
	}
	return true
}
