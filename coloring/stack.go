package coloring

import (
	"fmt"
	"image/color"
)

type ColorStackEntry struct {
	Weight float64
	Color  color.Color
}

type ColorStack struct {
	entries      []ColorStackEntry
	totalWeight  float64
	endingValues []float64
}

func NewColorStack(entries []ColorStackEntry) ColorStack {
	if len(entries) == 0 {
		panic("can not create a color sampling stack without any entries")
	}

	endingValues := make([]float64, len(entries))

	totalWeight := 0.
	for i, e := range entries {
		if e.Weight <= 0 {
			panic(fmt.Errorf("invalid weight value: %f", e.Weight))
		}
		totalWeight += e.Weight
		endingValues[i] = totalWeight
	}

	return ColorStack{
		entries:      entries,
		totalWeight:  totalWeight,
		endingValues: endingValues,
	}
}

func (cs ColorStack) LinearSample(v float64) color.Color {
	if v < 0 || v > 1 {
		panic(fmt.Errorf("invalid sample value: %f", v))
	}

	if len(cs.entries) == 1 {
		return cs.entries[0].Color
	}

	currentWeightRead := 0.

	selectedIndex := -1
	for i, entry := range cs.entries {
		currentWeightRead += entry.Weight
		adjustedWeight := currentWeightRead / cs.totalWeight

		if adjustedWeight >= v {
			selectedIndex = i
			break
		}
	}

	if selectedIndex == -1 {
		panic(fmt.Errorf("unimplemented situation: linear color sample could not find appropriate colors to sample"))
	}

	indexA := selectedIndex - 1
	indexB := selectedIndex
	if selectedIndex == 0 {
		indexA = 0
		indexB = 1
	}

	colorA := cs.entries[indexA]
	colorB := cs.entries[indexB]

	endA := cs.endingValues[indexA]
	endB := cs.endingValues[indexB]

	scaledV := v * cs.totalWeight
	distA := scaledV - endA
	distB := endB - scaledV
	distA2B := distA + distB
	percentA := distA / distA2B
	percentB := distB / distA2B

	cA_R, cA_G, cA_B, cA_A := colorA.Color.RGBA()
	cB_R, cB_G, cB_B, cB_A := colorB.Color.RGBA()

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
