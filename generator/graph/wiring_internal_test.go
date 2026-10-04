package graph

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitElement(t *testing.T) {
	for name, want := range map[string]struct {
		port    string
		element int
	}{
		"In":        {"In", nextElement},
		"Meshes.3":  {"Meshes", 3},
		"Meshes.10": {"Meshes", 10},
		"Scale.X":   {"Scale.X", nextElement},
		"Scale.X.2": {"Scale.X", 2},
		"Items.-1":  {"Items.-1", nextElement},
	} {
		port, element := splitElement(name)
		assert.Equal(t, want.port, port, name)
		assert.Equal(t, want.element, element, name)
	}
}
