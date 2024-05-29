package ply

type Format int64

const (
	ASCII Format = iota
	BinaryBigEndian
	BinaryLittleEndian
)

const VertexElementName = "vertex"
