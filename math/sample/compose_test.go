package sample_test

import (
	"testing"

	"github.com/EliCDavis/polyform/math/sample"
	"github.com/stretchr/testify/assert"
)

func TestComposeFloat_PanicOnNoFuncs(t *testing.T) {
	assert.Panics(t, func() {
		sample.ComposeFloat()
	})
}
func TestComposeVec2_PanicOnNoFuncs(t *testing.T) {
	assert.Panics(t, func() {
		sample.ComposeVec2()
	})
}

func TestComposeVec3_PanicOnNoFuncs(t *testing.T) {
	assert.Panics(t, func() {
		sample.ComposeVec3()
	})
}
