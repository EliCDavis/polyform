package pipeline

import (
	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector2"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type ReadArrayPermission[T any] struct {
	data []T
}

func (rdep ReadArrayPermission[T]) Data() *iter.ArrayIterator[T] {
	return iter.Array[T](rdep.data)
}

type ReadIndicesPermission struct {
	ReadArrayPermission[int]
	m *modeling.Mesh
}

func (ip ReadIndicesPermission) VertexNeighborTable() modeling.VertexLUT {
	return ip.m.VertexNeighborTable()
}

type ReadPermission[T any] struct {
	data T
}

func (rdep ReadPermission[T]) Data() T {
	return rdep.data
}

type MeshReadPermission struct {
	Everything    *ReadPermission[modeling.Mesh]
	Indices       *ReadIndicesPermission
	V1Permissions map[string]ReadArrayPermission[float64]
	V2Permissions map[string]ReadArrayPermission[vector2.Float64]
	V3Permissions map[string]ReadArrayPermission[vector3.Float64]
	V4Permissions map[string]ReadArrayPermission[vector4.Float64]
}

type WriteArrayPermission[T any] struct {
	data []T
}

func (wap WriteArrayPermission[T]) Data() []T {
	return wap.data
}

type WritePermission[T any] struct {
	data    T
	written bool
}

func (wp WritePermission[T]) Data() T {
	return wp.data
}

func (wp *WritePermission[T]) Write(val T) {
	wp.data = val
	wp.written = true
}

type MeshWritePermission struct {
	Everything    *WritePermission[modeling.Mesh]
	Indices       *WriteArrayPermission[int]
	V1Permissions map[string]WriteArrayPermission[float64]
	V2Permissions map[string]WriteArrayPermission[vector2.Float64]
	V3Permissions map[string]WriteArrayPermission[vector3.Float64]
	V4Permissions map[string]WriteArrayPermission[vector4.Float64]
}
