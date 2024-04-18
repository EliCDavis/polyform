package sdf_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/EliCDavis/polyform/math/sdf"
	"github.com/EliCDavis/vector/vector3"
	"github.com/stretchr/testify/assert"
)

func TestTranslate(t *testing.T) {
	sphere := sdf.Sphere(vector3.Zero[float64](), 1)

	tests := map[string]struct {
		pos   vector3.Float64
		field sample.Vec3ToFloat
		want  float64
	}{
		"center": {
			pos:   vector3.Zero[float64](),
			field: sphere,
			want:  -1.,
		},
		"on bounds": {
			pos:   vector3.Zero[float64](),
			field: sdf.Translate(sphere, vector3.Up[float64]()),
			want:  0.,
		},
		"outside": {
			pos:   vector3.Zero[float64](),
			field: sdf.Translate(sphere, vector3.Up[float64]().Scale(2)),
			want:  1.,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.field(tc.pos))
		})
	}
}
