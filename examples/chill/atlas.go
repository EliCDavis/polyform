package main

import (
	"math"
	"math/rand"

	"github.com/EliCDavis/vector/vector2"
)

type Atlas struct {
	Name       string
	BottomLeft vector2.Float64
	TopRight   vector2.Float64
	SubAtlas   []*Atlas
	Entries    []AtlasEntry
}

func (atlas Atlas) RandomEntry() AtlasEntry {
	return atlas.Entries[int(math.Floor(rand.Float64()*float64(len(atlas.Entries))))]
}

type AtlasEntry struct {
	BottomLeft vector2.Float64
	TopRight   vector2.Float64
}

func (ae AtlasEntry) Height() float64 {
	return ae.TopRight.Y() - ae.BottomLeft.Y()
}

func (ae AtlasEntry) Width() float64 {
	return ae.TopRight.X() - ae.BottomLeft.X()
}

func (ae AtlasEntry) Area() float64 {
	return ae.TopRight.Sub(ae.BottomLeft).Y()
}

func (ae AtlasEntry) MinX() float64 {
	return ae.BottomLeft.X()
}

func (ae AtlasEntry) MaxX() float64 {
	return ae.TopRight.X()
}

func (ae AtlasEntry) MinY() float64 {
	return ae.BottomLeft.Y()
}

func (ae AtlasEntry) MaxY() float64 {
	return ae.TopRight.Y()
}
