package ply_test

import (
	"testing"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/stretchr/testify/assert"
)

func TestElementPointSize(t *testing.T) {
	element := ply.Element{
		Properties: []ply.Property{
			ply.ScalarProperty{Type: ply.Double},
			ply.ScalarProperty{Type: ply.Double},
			ply.ScalarProperty{Type: ply.Double},
		},
	}

	size, err := element.PointSize()

	assert.Equal(t, 24, size)
	assert.NoError(t, err)
	assert.True(t, element.DeterministicPointSize())
}

func TestElementPointSize_ListProp(t *testing.T) {
	element := ply.Element{
		Properties: []ply.Property{
			ply.ScalarProperty{Type: ply.Double},
			ply.ListProperty{PropertyName: "Ye"},
			ply.ScalarProperty{Type: ply.Double},
		},
	}

	size, err := element.PointSize()

	assert.Equal(t, 0, size)
	assert.EqualError(t, err, `property "Ye" is not scalar, point size is variable`)
	assert.False(t, element.DeterministicPointSize())
}
