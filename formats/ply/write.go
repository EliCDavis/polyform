package ply

import (
	"io"

	"github.com/EliCDavis/polyform/modeling"
)

var defaultWriter = MeshWriter{
	Format: BinaryLittleEndian,
	Properties: []PropertyWriter{
		Vector3PropertyWriter{
			ModelAttribute: modeling.PositionAttribute,
			Type:           Float,
			PlyPropertyX:   "x",
			PlyPropertyY:   "y",
			PlyPropertyZ:   "z",
		},
		Vector3PropertyWriter{
			ModelAttribute: modeling.NormalAttribute,
			Type:           Float,
			PlyPropertyX:   "nx",
			PlyPropertyY:   "ny",
			PlyPropertyZ:   "nz",
		},
		Vector3PropertyWriter{
			ModelAttribute: modeling.ColorAttribute,
			Type:           UChar,
			PlyPropertyX:   "red",
			PlyPropertyY:   "green",
			PlyPropertyZ:   "blue",
		},
	},
}

func Write(out io.Writer, model modeling.Mesh, format Format) error {
	writer := defaultWriter
	writer.Format = format
	return writer.Write(model, out)
}
