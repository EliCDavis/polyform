package coloring

import (
	"fmt"
	"image/color"
)

type ColorStackEntry struct {
	size  float64
	color color.Color

	// What percentage of the left side do you want to have faded
	fadeLeft float64

	// What percentage of the right side do you want to have faded
	fadeRight float64
}

func NewColorStackEntry(size, fadeLeft, fadeRight float64, color color.Color) ColorStackEntry {
	if size <= 0 {
		panic("color stack entry size must be greater than 0")
	}

	if fadeLeft < 0 || fadeLeft > 1 {
		panic("color stack entry fade left value must be a value between 0 and 1")
	}

	if fadeRight < 0 || fadeRight > 1 {
		panic("color stack entry fade right value must be a value between 0 and 1")
	}

	return ColorStackEntry{
		size:      size,
		color:     color,
		fadeLeft:  fadeLeft,
		fadeRight: fadeRight,
	}
}

type ColorStack struct {
	entries          []ColorStackEntry
	totalWeight      float64
	startValues      []float64
	rightBlendValues []float64
	leftBlendValues  []float64
}

func NewColorStack(entries []ColorStackEntry) ColorStack {
	if len(entries) == 0 {
		panic("can not create a color sampling stack without any entries")
	}

	rightBlendValues := make([]float64, len(entries))
	leftBlendValues := make([]float64, len(entries))
	startValues := make([]float64, len(entries))

	totalWeight := 0.
	for i, e := range entries {
		halfSize := e.size / 2.

		startValues[i] = totalWeight

		leftBlendValues[i] = totalWeight + ((e.fadeLeft) * halfSize)

		rightBlendValues[i] = totalWeight + ((1. - e.fadeRight) * halfSize) + halfSize

		totalWeight += e.size
	}

	return ColorStack{
		entries:          entries,
		totalWeight:      totalWeight,
		startValues:      startValues,
		leftBlendValues:  leftBlendValues,
		rightBlendValues: rightBlendValues,
	}
}

func (cs ColorStack) LinearSample(v float64) color.Color {
	if v < 0 || v > 1 {
		panic(fmt.Errorf("invalid sample value: %f", v))
	}

	if len(cs.entries) == 1 {
		return cs.entries[0].color
	}

	scaledV := v * cs.totalWeight

	if scaledV <= cs.rightBlendValues[0] {
		return cs.entries[0].color
	}

	if scaledV >= cs.leftBlendValues[len(cs.entries)-1] {
		return cs.entries[len(cs.entries)-1].color
	}

	indexA := -1
	indexB := -1
	for i, entry := range cs.entries {

		// Nothing to blend with
		if scaledV >= cs.leftBlendValues[i] && scaledV <= cs.rightBlendValues[i] {
			return entry.color
		}

		if scaledV < cs.leftBlendValues[i] && scaledV > cs.rightBlendValues[i-1] {
			indexA = i - 1
			indexB = i
			break
		}

		if scaledV > cs.rightBlendValues[i] && scaledV < cs.leftBlendValues[i+1] {
			indexA = i
			indexB = i + 1
			break
		}
	}

	if indexA == -1 {
		panic(fmt.Errorf("unimplemented situation: linear color sample could not find appropriate colors to sample"))
	}

	adjustedStart := scaledV - cs.rightBlendValues[indexA]

	weightA := cs.startValues[indexB] - cs.rightBlendValues[indexA]
	weightB := cs.leftBlendValues[indexB] - cs.startValues[indexB]

	var percentA float64
	var percentB float64

	if adjustedStart < weightA {
		// We're on the left side, trending towards A
		percentA = 1. - ((adjustedStart / weightA) * 0.5)
		percentB = 1. - percentA
	} else {
		// We're on the right side, trending towards B
		percentB = (((adjustedStart - weightA) / weightB) * 0.5) + 0.5
		percentA = 1. - percentB
	}

	// Debug code for helping me get math right.
	if percentA < 0 || percentB < 0 || percentA > 1 || percentB > 1 {
		return color.RGBA{
			R: 0,
			G: 0,
			B: 255,
			A: 255,
		}
	}

	cA_R, cA_G, cA_B, cA_A := cs.entries[indexA].color.RGBA()
	cB_R, cB_G, cB_B, cB_A := cs.entries[indexB].color.RGBA()

	final_R := uint32((float64(cA_R) * percentA) + (percentB * float64(cB_R)))
	final_G := uint32((float64(cA_G) * percentA) + (percentB * float64(cB_G)))
	final_B := uint32((float64(cA_B) * percentA) + (percentB * float64(cB_B)))
	final_A := uint32((float64(cA_A) * percentA) + (percentB * float64(cB_A)))

	return color.RGBA{
		R: uint8(final_R >> 8),
		G: uint8(final_G >> 8),
		B: uint8(final_B >> 8),
		A: uint8(final_A >> 8),
	}
}
