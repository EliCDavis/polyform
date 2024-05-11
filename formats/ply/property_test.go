package ply_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/stretchr/testify/assert"
)

func TestScalarPropertySize(t *testing.T) {
	tests := map[string]struct {
		input ply.ScalarProperty
		want  int
	}{
		"char": {
			input: ply.ScalarProperty{Type: ply.Char},
			want:  1,
		},
		"uchar": {
			input: ply.ScalarProperty{Type: ply.UChar},
			want:  1,
		},
		"short": {
			input: ply.ScalarProperty{Type: ply.Short},
			want:  2,
		},
		"ushort": {
			input: ply.ScalarProperty{Type: ply.UShort},
			want:  2,
		},
		"int": {
			input: ply.ScalarProperty{Type: ply.Int},
			want:  4,
		},
		"uint": {
			input: ply.ScalarProperty{Type: ply.UInt},
			want:  4,
		},
		"float": {
			input: ply.ScalarProperty{Type: ply.Float},
			want:  4,
		},
		"double": {
			input: ply.ScalarProperty{Type: ply.Double},
			want:  8,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.input.Size())
		})
	}
}
