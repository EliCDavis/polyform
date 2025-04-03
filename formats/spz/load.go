package spz

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/EliCDavis/vector/vector4"
)

func degreeForDim(dim int) int {
	if dim < 3 {
		return 0
	}

	if dim < 8 {
		return 1
	}

	if dim < 15 {
		return 2
	}

	return 3
}

// https://github.com/aras-p/UnityGaussianSplatting/blob/main/package/Shaders/GaussianSplatting.hlsl#L139
/*
half3 ShadeSH(SplatSHData splat, half3 dir, int shOrder, bool onlySH)
{
    dir *= -1;

    half x = dir.x, y = dir.y, z = dir.z;

    // ambient band
    half3 res = splat.col; // col = sh0 * SH_C0 + 0.5 is already precomputed
    if (onlySH)
        res = 0.5;
    // 1st degree
    if (shOrder >= 1)
    {
        res += SH_C1 * (-splat.sh1 * y + splat.sh2 * z - splat.sh3 * x);
        // 2nd degree
        if (shOrder >= 2)
        {
            half xx = x * x, yy = y * y, zz = z * z;
            half xy = x * y, yz = y * z, xz = x * z;
            res +=
                (SH_C2[0] * xy) * splat.sh4 +
                (SH_C2[1] * yz) * splat.sh5 +
                (SH_C2[2] * (2 * zz - xx - yy)) * splat.sh6 +
                (SH_C2[3] * xz) * splat.sh7 +
                (SH_C2[4] * (xx - yy)) * splat.sh8;
            // 3rd degree
            if (shOrder >= 3)
            {
                res +=
                    (SH_C3[0] * y * (3 * xx - yy)) * splat.sh9 +
                    (SH_C3[1] * xy * z) * splat.sh10 +
                    (SH_C3[2] * y * (4 * zz - xx - yy)) * splat.sh11 +
                    (SH_C3[3] * z * (2 * zz - 3 * xx - 3 * yy)) * splat.sh12 +
                    (SH_C3[4] * x * (4 * zz - xx - yy)) * splat.sh13 +
                    (SH_C3[5] * z * (xx - yy)) * splat.sh14 +
                    (SH_C3[6] * x * (xx - 3 * yy)) * splat.sh15;
            }
        }
    }
    return max(res, 0);
}
*/

const sh_C1 = 0.4886025

var sh_C2 = []float64{1.0925484, -1.0925484, 0.3153916, -1.0925484, 0.5462742}
var sh_C3 = []float64{-0.5900436, 2.8906114, -0.4570458, 0.3731763, -0.4570458, 1.4453057, -0.5900436}

func shadeSH(splat []vector3.Float64, dir vector3.Float64, shOrder int, col float64, c int) float64 {
	x := dir.X() * -1
	y := dir.Y() * -1
	z := dir.Z() * -1

	res := col

	if shOrder >= 1 {
		res += sh_C1 * (-splat[0].Component(c)*y + splat[1].Component(c)*z - splat[2].Component(c)*x)
		return math.Max(res, 0)
	}

	if shOrder >= 2 {
		xx := x * x
		yy := y * y
		zz := z * z
		xy := x * y
		yz := y * z
		xz := x * z
		res += (sh_C2[0]*xy)*splat[3].Component(c) +
			(sh_C2[1]*yz)*splat[4].Component(c) +
			(sh_C2[2]*(2*zz-xx-yy))*splat[5].Component(c) +
			(sh_C2[3]*xz)*splat[6].Component(c) +
			(sh_C2[4]*(xx-yy))*splat[7].Component(c)

		if shOrder >= 3 {
			res +=
				(sh_C3[0]*y*(3*xx-yy))*splat[8].Component(c) +
					(sh_C3[1]*xy*z)*splat[9].Component(c) +
					(sh_C3[2]*y*(4*zz-xx-yy))*splat[10].Component(c) +
					(sh_C3[3]*z*(2*zz-3*xx-3*yy))*splat[11].Component(c) +
					(sh_C3[4]*x*(4*zz-xx-yy))*splat[12].Component(c) +
					(sh_C3[5]*z*(xx-yy))*splat[13].Component(c) +
					(sh_C3[6]*x*(xx-3*yy))*splat[14].Component(c)
		}
	}

	return math.Max(res, 0)
}

func ReadHeader(in io.Reader) (*Header, error) {
	reader, err := gzip.NewReader(in)
	if err != nil {
		return nil, err
	}

	var header Header
	if err := binary.Read(reader, binary.LittleEndian, &header); err != nil {
		return nil, err
	}

	return &header, header.Validate()
}

// Deserialize a gaussian splat from the input reader.
func Read(inUncompressed io.Reader) (*Cloud, error) {
	in, err := gzip.NewReader(inUncompressed)
	if err != nil {
		return nil, err
	}

	var header Header
	if err := binary.Read(in, binary.LittleEndian, &header); err != nil {
		return nil, err
	}
	// panic(fmt.Errorf("%+v\n", header))

	if err := header.Validate(); err != nil {
		return nil, err
	}

	positions, err := header.readPositions(in)
	if err != nil {
		return nil, err
	}

	alphas, err := header.readAlphas(in)
	if err != nil {
		return nil, err
	}

	colors, err := header.readColors(in)
	if err != nil {
		return nil, err
	}

	scales, err := header.readScale(in)
	if err != nil {
		return nil, err
	}

	rotations, err := header.readRotations(in)
	if err != nil {
		return nil, err
	}

	sh, err := header.readSh(in)
	if err != nil {
		return nil, err
	}

	// dir := vector3.Forward[float64]()
	// for i := 0; i < len(colors); i++ {

	// 	harmonics := make([]vector3.Float64, 0)
	// 	for _, h := range sh {
	// 		harmonics = append(harmonics, h[i])
	// 	}

	// 	r := ShadeSH(harmonics, dir, int(header.ShDegree), colors[i].X(), 0)
	// 	g := ShadeSH(harmonics, dir, int(header.ShDegree), colors[i].Y(), 1)
	// 	b := ShadeSH(harmonics, dir, int(header.ShDegree), colors[i].Z(), 2)
	// 	colors[i] = vector3.New(r, g, b)
	// }

	v3Data := map[string][]vector3.Vector[float64]{
		modeling.PositionAttribute: positions,
		modeling.ScaleAttribute:    scales,
		modeling.FDCAttribute:      colors,
	}

	for i, h := range sh {
		v3Data[fmt.Sprintf("SH_%d", i)] = h
	}

	/*
	  const int numPoints = header.numPoints;
	  PackedGaussians result = {
	    .numPoints = numPoints,
	    .shDegree = header.shDegree,
	    .fractionalBits = header.fractionalBits,
	    .antialiased = (header.flags & FlagAntialiased) != 0};
	  result.positions.resize(numPoints * 3 * (usesFloat16 ? 2 : 3));
	  result.scales.resize(numPoints * 3);
	  result.rotations.resize(numPoints * 3);
	  result.alphas.resize(numPoints);
	  result.colors.resize(numPoints * 3);
	  result.sh.resize(numPoints * shDim * 3);
	  in.read(reinterpret_cast<char *>(result.positions.data()), countBytes(result.positions));
	  in.read(reinterpret_cast<char *>(result.alphas.data()), countBytes(result.alphas));
	  in.read(reinterpret_cast<char *>(result.colors.data()), countBytes(result.colors));
	  in.read(reinterpret_cast<char *>(result.scales.data()), countBytes(result.scales));
	  in.read(reinterpret_cast<char *>(result.rotations.data()), countBytes(result.rotations));
	  in.read(reinterpret_cast<char *>(result.sh.data()), countBytes(result.sh));
	  if (!in) {
	    SpzLog("[SPZ ERROR] deserializePackedGaussians: read error");
	    return {};
	  }
	  return result;
	*/

	return &Cloud{
		Header: header,
		Mesh: modeling.NewPointCloud(
			map[string][]vector4.Vector[float64]{
				modeling.RotationAttribute: rotations,
			},
			v3Data,
			nil,
			map[string][]float64{
				modeling.OpacityAttribute: alphas,
			},
		),
	}, nil
}

// Opens the file located at the filePath and deserializes a gaussian splat.
func Load(filePath string) (*Cloud, error) {
	// https://github.com/nianticlabs/spz/blob/main/src/cc/load-spz.cc#L458
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)
	a, err := Read(r)
	if err != nil {
		panic(err)
	}

	return a, err
}
