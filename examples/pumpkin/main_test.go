package main

import (
	"testing"

	"github.com/EliCDavis/vector/vector2"
)

func TestNoise(t *testing.T) {
	tn := TilingNoise{}
	tn.init()
	t.Error(tn.Noise(vector2.New(0.1, 0.1), 2))
}
