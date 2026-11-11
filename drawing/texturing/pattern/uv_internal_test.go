package pattern

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCellIndexWrapsAcrossTheUVSeam(t *testing.T) {
	assert.Equal(t, cellIndex(0.05, 4), cellIndex(1.05, 4))
	assert.Equal(t, 3, cellIndex(-0.05, 4))
}

func TestHashIsStableAndSpreadsAcrossCells(t *testing.T) {
	assert.Equal(t, hash01(3, 7, 1), hash01(3, 7, 1))
	assert.NotEqual(t, hash01(3, 7, 1), hash01(3, 7, 2))

	sum, distinct := 0., map[float64]bool{}
	for column := range 32 {
		for row := range 32 {
			v := hash01(column, row, 0)
			assert.GreaterOrEqual(t, v, 0.)
			assert.Less(t, v, 1.)
			sum += v
			distinct[v] = true
		}
	}
	assert.InDelta(t, 0.5, sum/1024, 0.05)
	assert.Len(t, distinct, 1024)
}

func TestStripeCoversItsWidth(t *testing.T) {
	assert.Equal(t, 1., stripe(0.05, 4, 0.5, 0))
	assert.Equal(t, 0., stripe(0.2, 4, 0.5, 0))
	assert.InDelta(t, 0.5, stripe(0.125, 1, 0.125, 0.1), 1e-9)
}

func TestPlainWeaveAlternatesWhichStrandIsOnTop(t *testing.T) {
	w := weave{columns: 4, rows: 4, over: 1, under: 1}
	assert.True(t, w.sample(0.125, 0.125).alongU)
	assert.False(t, w.sample(0.375, 0.125).alongU)
	assert.False(t, w.sample(0.125, 0.375).alongU)
	assert.True(t, w.sample(0.375, 0.375).alongU)
}

func TestWeaveTopStrandPeaksMidFloatAndStaysInRange(t *testing.T) {
	w := weave{columns: 8, rows: 8, over: 2, under: 2, gap: 0.1}
	for i := range 256 {
		for j := range 256 {
			h := w.sample((float64(i)+0.5)/256, (float64(j)+0.5)/256).height
			assert.GreaterOrEqual(t, h, 0.)
			assert.LessOrEqual(t, h, 1.)
		}
	}
	mid := w.sample(1./8, 0.5/8)
	end := w.sample(0.02/8, 0.5/8)
	assert.Greater(t, mid.height, end.height)
}

func TestWeaveGapShowsNothingWhereNoStrandRuns(t *testing.T) {
	w := weave{columns: 4, rows: 4, over: 1, under: 1, gap: 0.5}
	s := w.sample(0.01, 0.01)
	assert.Equal(t, 0., s.height)
	assert.Equal(t, -1, s.strand)
}

func TestHerringboneReversesTheDiagonal(t *testing.T) {
	twill := weave{columns: 8, rows: 8, over: 2, under: 2}
	zigzag := weave{columns: 8, rows: 8, over: 2, under: 2, herringbone: 2}
	at := func(w weave, column, row int) bool {
		return w.sample((float64(column)+0.5)/8, (float64(row)+0.5)/8).alongU
	}
	for row := range 8 {
		assert.Equal(t, at(twill, 0, row), at(zigzag, 0, row))
		assert.Equal(t, at(twill, 1, row), at(zigzag, 1, row))
	}
	differs := false
	for row := range 8 {
		differs = differs || at(twill, 2, row) != at(zigzag, 2, row)
	}
	assert.True(t, differs)
}

func TestWeaveTilesAcrossTheUVSeam(t *testing.T) {
	w := weave{columns: 8, rows: 8, over: 2, under: 2, herringbone: 2}
	inside, wrapped := w.sample(0.03, 0.4), w.sample(1.03, 0.4)
	assert.InDelta(t, inside.height, wrapped.height, 1e-9)
	assert.Equal(t, inside.alongU, wrapped.alongU)
	assert.Equal(t, inside.strand, wrapped.strand)
}
