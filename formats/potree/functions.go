package potree

import (
	"encoding/binary"

	"github.com/EliCDavis/vector/vector3"
)

func LoadNodePositionDataIntoArray(m *Metadata, buf []byte, positions []vector3.Float64) {
	attributeOffset := 0

	for _, attribute := range m.Attributes {
		if attribute.IsPosition() {

			endian := binary.LittleEndian
			bytesPerPoint := m.BytesPerPoint()
			scale := vector3.New(m.Scale[0], m.Scale[1], m.Scale[2])
			shift := vector3.New(m.Offset[0], m.Offset[1], m.Offset[2])

			pointOffset := attributeOffset
			for i := 0; i < len(positions); i++ {
				positions[i] = vector3.
					New(
						int(endian.Uint32(buf[pointOffset:])),
						int(endian.Uint32(buf[pointOffset+4:])),
						int(endian.Uint32(buf[pointOffset+8:])),
					).
					ToFloat64().
					MultByVector(scale).
					Add(shift)
				pointOffset += bytesPerPoint
			}
			return
		}
		attributeOffset += attribute.Size
	}
}

func LoadNodeColorDataIntoArray(m *Metadata, buf []byte, colors []vector3.Float64) {
	attributeOffset := 0

	for _, attribute := range m.Attributes {
		if attribute.IsColor() {
			endian := binary.LittleEndian
			bytesPerPoint := m.BytesPerPoint()

			pointOffset := attributeOffset
			for i := 0; i < len(colors); i++ {
				col := vector3.
					New(
						int(endian.Uint16(buf[pointOffset:])),
						int(endian.Uint16(buf[pointOffset+2:])),
						int(endian.Uint16(buf[pointOffset+4:])),
					).
					ToFloat64()

				if col.X() > 255 || col.Y() > 255 || col.Z() > 255 {
					col = col.DivByConstant(256)
				}

				colors[i] = col.DivByConstant(255)
				pointOffset += bytesPerPoint
			}
			return
		}
		attributeOffset += attribute.Size
	}
}
