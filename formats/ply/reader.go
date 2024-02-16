package ply

import "github.com/EliCDavis/polyform/modeling"

type BodyReader interface {
	ReadMesh(vertexAttributes map[string]bool) (*modeling.Mesh, error)
}

var GuassianSplatVertexAttributes map[string]bool = map[string]bool{
	"x":       true,
	"y":       true,
	"z":       true,
	"scale_0": true,
	"scale_1": true,
	"scale_2": true,
	"rot_0":   true,
	"rot_1":   true,
	"rot_2":   true,
	"rot_3":   true,
	"f_dc_0":  true,
	"f_dc_1":  true,
	"f_dc_2":  true,
	"opacity": true,
}
