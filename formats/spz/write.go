package spz

import (
	"encoding/binary"
	"io"

	"github.com/EliCDavis/polyform/modeling"
)

func Write(cloud modeling.Mesh, out io.Writer) error {
	header := Header{
		Magic:          magicNum,
		Version:        2,
		NumPoints:      uint32(cloud.PrimitiveCount()),
		ShDegree:       0,
		FractionalBits: 3,
	}

	return binary.Write(out, binary.LittleEndian, header)
}
