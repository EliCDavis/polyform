package ply

import "github.com/EliCDavis/mesh"

type reader interface {
	ReadMesh() (*mesh.Mesh, error)
}
