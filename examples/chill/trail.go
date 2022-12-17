package main

type Trail struct {
	Segments []TrailSegment `json:"segments"`
}

type TrailSegment struct {
	Width float64 `json:"width"`
	Depth float64 `json:"depth"`

	StartX float64 `json:"startX"`
	StartY float64 `json:"startY"`

	EndX float64 `json:"endX"`
	EndY float64 `json:"endY"`
}
