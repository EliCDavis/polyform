package spz

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"

	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

type Header struct {
	Magic          uint32 `json:"magic"`     // Must be 0x5053474e (NGSP = Niantic gaussian splat)
	Version        uint32 `json:"version"`   // Must be 2
	NumPoints      uint32 `json:"numPoints"` // Cant be greater than 10000000
	ShDegree       uint8  `json:"shDegree"`  // Must be between 0 and 3
	FractionalBits uint8  `json:"fractionalBits"`
	Flags          uint8  `json:"flags"`
	Reserved       uint8  `json:"reserved"` // Must be 0
}

func (pgh Header) Validate() error {
	if pgh.Magic != 0x5053474e {
		return fmt.Errorf("invalid magic number in header: %d", pgh.Magic)
	}

	if pgh.Version < 1 || pgh.Version > 2 {
		return fmt.Errorf("unsupported version: %d", pgh.Version)
	}

	const maxPointsToRead uint32 = 10000000
	if pgh.NumPoints > maxPointsToRead {
		return fmt.Errorf("header defines too many points: %d", pgh.NumPoints)
	}

	if pgh.ShDegree > 3 {
		return fmt.Errorf("unsupported SH degree: %d", pgh.ShDegree)
	}

	return nil
}

func (pgh Header) Float16Positions() bool {
	return pgh.Version == 1
}

func (pgh Header) readPositionsFloat16(in io.Reader) ([]vector3.Float64, error) {
	positionData := make([]uint16, pgh.NumPoints*3)
	if err := binary.Read(in, binary.LittleEndian, &positionData); err != nil {
		return nil, err
	}

	positions := make([]vector3.Float64, pgh.NumPoints)
	for i := 0; i < len(positions); i++ {
		i3 := i * 3
		positions[i] = vector3.New(
			halfToFloat(positionData[i3]),
			halfToFloat(positionData[i3+1]),
			halfToFloat(positionData[i3+2]),
		)
	}
	return positions, nil
}

func unquantizeSH(x uint8) float64 {
	return (float64(x) - 128.0) / 128.0
}

func (pgh Header) readSh(in io.Reader) ([][]vector3.Float64, error) {
	shDim, err := dimForDegree(int(pgh.ShDegree))
	if err != nil {
		return nil, err
	}

	if shDim == 0 {
		return nil, nil
	}

	shData := make([]byte, pgh.NumPoints*3*uint32(shDim))
	if _, err := io.ReadFull(in, shData); err != nil {
		return nil, err
	}

	sh := make([][]vector3.Float64, uint32(shDim))

	for i := 0; i < len(sh); i++ {
		sh[i] = make([]vector3.Vector[float64], pgh.NumPoints)
	}

	for i := 0; i < int(pgh.NumPoints); i++ {
		for d := 0; d < shDim; d++ {
			i3 := d*3 + (i * 3 * int(shDim))
			sh[d][i] = vector3.New(
				unquantizeSH(shData[i3+0]),
				unquantizeSH(shData[i3+1]),
				unquantizeSH(shData[i3+2]),
			)
		}
	}

	return sh, nil
}

func (pgh Header) readRotations(in io.Reader) ([]vector4.Float64, error) {

	/*
	   const uint8_t *r = &rotation[0];
	     Vec3f xyz = plus(
	       times(
	         Vec3f{static_cast<float>(r[0]), static_cast<float>(r[1]), static_cast<float>(r[2])},
	         1.0f / 127.5f),
	       Vec3f{-1, -1, -1});
	     std::copy(xyz.data(), xyz.data() + 3, &result.rotation[0]);
	     // Compute the real component - we know the quaternion is normalized and w is non-negative
	     result.rotation[3] = std::sqrt(std::max(0.0f, 1.0f - squaredNorm(xyz)));
	*/

	rotationData := make([]byte, pgh.NumPoints*3)
	if _, err := io.ReadFull(in, rotationData); err != nil {
		return nil, err
	}

	// Decode 24-bit fixed point coordinates
	rotations := make([]vector4.Float64, pgh.NumPoints)
	const scale = 1. / 127.5
	for i := 0; i < len(rotations); i++ {
		i3 := i * 3

		v := vector3.New(
			(float64(rotationData[i3+0])*scale)-1,
			(float64(rotationData[i3+1])*scale)-1,
			(float64(rotationData[i3+2])*scale)-1,
		)
		rotations[i] = vector4.New(v.X(), v.Y(), v.Z(), math.Sqrt(math.Max(0, 1-v.Dot(v))))
	}
	return rotations, nil

}

func (pgh Header) readScale(in io.Reader) ([]vector3.Float64, error) {
	/*
		for (size_t i = 0; i < 3; i++) {
			result.scale[i] = (scale[i] / 16.0f - 10.0f);
		}
	*/
	scaleData := make([]byte, pgh.NumPoints*3)
	if _, err := io.ReadFull(in, scaleData); err != nil {
		return nil, err
	}

	// Decode 24-bit fixed point coordinates
	scales := make([]vector3.Float64, pgh.NumPoints)
	for i := 0; i < len(scales); i++ {
		i3 := i * 3

		scales[i] = vector3.New(
			float64(scaleData[i3])/16.0-10.0,
			float64(scaleData[i3+1])/16.0-10.0,
			float64(scaleData[i3+2])/16.0-10.0,
		)
	}
	return scales, nil
}

func (pgh Header) readAlphas(in io.Reader) ([]float64, error) {
	/*
		float sigmoid(float x) { return 1 / (1 + std::exp(-x)); }
		float invSigmoid(float x) { return std::log(x / (1.0f - x)); }
		result.alpha = invSigmoid(alpha / 255.0f);
	*/
	alpha := make([]byte, pgh.NumPoints)
	if _, err := io.ReadFull(in, alpha); err != nil {
		return nil, err
	}

	alphas := make([]float64, pgh.NumPoints)
	for i := 0; i < len(alphas); i++ {
		x := float64(alpha[i]) / 255.

		// I feel like this is right but it looks worse when I put this in
		// alphas[i] = math.Log(x / (1.0 - x))

		alphas[i] = x
	}
	return alphas, nil
}

func (pgh Header) readColors(in io.Reader) ([]vector3.Float64, error) {
	colorData := make([]byte, pgh.NumPoints*3)
	if _, err := io.ReadFull(in, colorData); err != nil {
		return nil, err
	}

	colors := make([]vector3.Float64, pgh.NumPoints)
	for i := 0; i < len(colors); i++ {
		i3 := i * 3
		colors[i] = vector3.New(float64(colorData[i3]), float64(colorData[i3+1]), float64(colorData[i3+2])).
			DivByConstant(255.).
			Sub(vector3.Fill(0.5)).
			// Scale factor for DC color components. To convert to RGB, we
			// should multiply by 0.282, but it can be useful to represent base
			// colors that are out of range if the higher spherical harmonics
			// bands bring them back into range so we multiply by a smaller
			// value.
			DivByConstant(0.15)

		// colors[i] = vector3.New(
		// 	math.Pow(colors[i].X(), 2.4),
		// 	math.Pow(colors[i].Y(), 2.4),
		// 	math.Pow(colors[i].Z(), 2.4),
		// )
		// colors[i] = vector3.New(float64(colorData[i3]), float64(colorData[i3+1]), float64(colorData[i3+2]))

	}
	return colors, nil
}

// https://github.com/nianticlabs/spz/blob/main/src/cc/load-spz.cc#L291
func (pgh Header) readPositions(in io.Reader) ([]vector3.Float64, error) {
	if pgh.Float16Positions() {
		return pgh.readPositionsFloat16(in)
	}

	positionData := make([]byte, pgh.NumPoints*9)
	if _, err := io.ReadFull(in, positionData); err != nil {
		return nil, err
	}

	// Decode 24-bit fixed point coordinates
	b := 1 << pgh.FractionalBits
	scale := 1.0 / float64(b)

	positions := make([]vector3.Float64, pgh.NumPoints)
	for i := 0; i < len(positions); i++ {
		i9 := i * 9

		fixed32_X := uint32(positionData[i9+0])
		fixed32_X |= uint32(positionData[i9+1]) << 8
		fixed32_X |= uint32(positionData[i9+2]) << 16
		if fixed32_X&0x800000 > 0 {
			fixed32_X |= 0xff000000
		}

		fixed32_Y := uint32(positionData[i9+3])
		fixed32_Y |= uint32(positionData[i9+4]) << 8
		fixed32_Y |= uint32(positionData[i9+5]) << 16
		if fixed32_Y&0x800000 > 0 {
			fixed32_Y |= 0xff000000
		}

		fixed32_Z := uint32(positionData[i9+6])
		fixed32_Z |= uint32(positionData[i9+7]) << 8
		fixed32_Z |= uint32(positionData[i9+8]) << 16
		if fixed32_Z&0x800000 > 0 {
			fixed32_Z |= 0xff000000
		}

		positions[i] = vector3.New(int32(fixed32_X), int32(fixed32_Y), int32(fixed32_Z)).
			ToFloat64().
			Scale(scale)
	}

	/*
	   for (size_t i = 0; i < 3; i++) {
	     int32_t fixed32 = position[i * 3 + 0];
	     fixed32 |= position[i * 3 + 1] << 8;
	     fixed32 |= position[i * 3 + 2] << 16;
	     fixed32 |= (fixed32 & 0x800000) ? 0xff000000 : 0;  // sign extension
	     result.position[i] = static_cast<float>(fixed32) * scale;
	   }
	*/
	return positions, nil
}
