package splat

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/bitlib"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

// https://github.com/antimatter15/splat/blob/main/convert.py#L10
func Write(out io.Writer, mesh modeling.Mesh) error {

	requiredAttributes := []string{
		modeling.PositionAttribute,
		modeling.ScaleAttribute,
		modeling.FDCAttribute,
		// modeling.OpacityAttribute,
		"opacity",
		modeling.RotationAttribute,
	}

	for _, attr := range requiredAttributes {
		if !mesh.HasVertexAttribute(attr) {
			return fmt.Errorf("required attribute not present on mesh: %s", attr)
		}
	}

	if mesh.Topology() != modeling.PointTopology {
		return fmt.Errorf("mesh must be point topology, was instead %s", mesh.Topology())
	}

	count := mesh.PrimitiveCount()

	posData := mesh.Float3Attribute(modeling.PositionAttribute)
	scaleData := mesh.Float3Attribute(modeling.ScaleAttribute)
	fdcData := mesh.Float3Attribute(modeling.FDCAttribute)
	opacityData := mesh.Float1Attribute("opacity")
	rotationData := mesh.Float4Attribute(modeling.RotationAttribute)

	writer := bitlib.NewWriter(out, binary.LittleEndian)

	const SH_C0 = 0.28209479177387814

	for i := 0; i < count; i++ {
		pos := posData.At(i)
		writer.Float32(float32(pos.X()))
		writer.Float32(float32(pos.Y()))
		writer.Float32(float32(pos.Z()))

		scale := scaleData.At(i)
		writer.Float32(float32(math.Exp(scale.X())))
		writer.Float32(float32(math.Exp(scale.Y())))
		writer.Float32(float32(math.Exp(scale.Z())))

		color := fdcData.
			At(i).
			Scale(SH_C0).
			Add(vector3.Fill(0.5)).
			Clamp(0, 1)
		writer.Byte(byte(color.X() * 255))
		writer.Byte(byte(color.Y() * 255))
		writer.Byte(byte(color.Z() * 255))

		alpha := 1. / (1 + math.Exp(-opacityData.At(i)))
		writer.Byte(byte(alpha * 255))

		rot := rotationData.At(i)
		writer.Byte(byte(rot.X()*128) + 128)
		writer.Byte(byte(rot.Y()*128) + 128)
		writer.Byte(byte(rot.Z()*128) + 128)
		writer.Byte(byte(rot.W()*128) + 128)

		if writer.Error() != nil {
			return writer.Error()
		}
	}

	return nil
}
