package splat

import (
	"encoding/binary"
	"io"
	"math"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func Read(in io.Reader) (modeling.Mesh, error) {

	// 12 + 12 + 4 + 4 = 32
	splatBuffer := make([]byte, 32)

	positionData := make([]vector3.Float64, 0)
	scaleData := make([]vector3.Float64, 0)
	colorData := make([]vector3.Float64, 0)
	opacityData := make([]float64, 0)
	rotationData := make([]vector4.Float64, 0)

	var err error
	for {
		_, err = io.ReadFull(in, splatBuffer)
		if err != nil {
			break
		}

		positionData = append(positionData, vector3.New(
			math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer)),
			math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer[4:])),
			math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer[8:])),
		).ToFloat64())

		scaleData = append(scaleData, vector3.New(
			math.Log(float64(math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer[12:])))),
			math.Log(float64(math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer[16:])))),
			math.Log(float64(math.Float32frombits(binary.LittleEndian.Uint32(splatBuffer[20:])))),
		).ToFloat64())

		colorData = append(colorData, vector3.New(
			((float64(splatBuffer[24])/255.)-0.5)/SH_C0,
			((float64(splatBuffer[25])/255.)-0.5)/SH_C0,
			((float64(splatBuffer[26])/255.)-0.5)/SH_C0,
		).ToFloat64())

		a := (float64(splatBuffer[27]) / 255.)
		opacityData = append(opacityData, -math.Log((1/a)-1))

		rotationData = append(rotationData, vector4.New(
			(float64(splatBuffer[28])-128)/128,
			(float64(splatBuffer[29])-128)/128,
			(float64(splatBuffer[30])-128)/128,
			(float64(splatBuffer[31])-128)/128,
		).ToFloat64())
	}

	if err == io.EOF {
		err = nil
	}

	return modeling.NewPointCloud(
		map[string][]vector4.Vector[float64]{
			modeling.RotationAttribute: rotationData,
		},
		map[string][]vector3.Vector[float64]{
			modeling.PositionAttribute: positionData,
			modeling.ScaleAttribute:    scaleData,
			modeling.FDCAttribute:      colorData,
		},
		nil,
		map[string][]float64{
			modeling.OpacityAttribute: opacityData,
		},
	), err
}
