package ply

import "github.com/EliCDavis/polyform/modeling"

type reader interface {
	ReadMesh() (*modeling.Mesh, error)
}
