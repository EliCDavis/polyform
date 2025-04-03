package ply

import (
	"io"

	"github.com/EliCDavis/polyform/modeling"
)

var defaultWriter = MeshWriter{
	Format:                     BinaryLittleEndian,
	WriteUnspecifiedProperties: true,
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

		// Gaussian Splatting >>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>
		&Vector3PropertyWriter{
			ModelAttribute: modeling.FDCAttribute,
			Type:           Float,
			PlyPropertyX:   "f_dc_0",
			PlyPropertyY:   "f_dc_1",
			PlyPropertyZ:   "f_dc_2",
		},
		&Vector1PropertyWriter{
			ModelAttribute: modeling.OpacityAttribute,
			Type:           Float,
			PlyProperty:    "opacity",
		},
		&Vector3PropertyWriter{
			ModelAttribute: modeling.ScaleAttribute,
			Type:           Float,
			PlyPropertyX:   "scale_0",
			PlyPropertyY:   "scale_1",
			PlyPropertyZ:   "scale_2",
		},
		&Vector4PropertyWriter{
			ModelAttribute: modeling.RotationAttribute,
			Type:           Float,
			PlyPropertyX:   "rot_0",
			PlyPropertyY:   "rot_1",
			PlyPropertyZ:   "rot_2",
			PlyPropertyW:   "rot_3",
		},
		// <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
	},
}

func Write(out io.Writer, model modeling.Mesh, format Format, texture string) error {
	writer := defaultWriter
	writer.Format = format
	return writer.Write(model, texture, out)
}
