package normals

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
	"github.com/EliCDavis/polyform/utils"
	"github.com/EliCDavis/vector/vector3"
)

/*
uniform sampler2D unit_wave
noperspective in vec2 tex_coord;
const vec2 size = vec2(2.0,0.0);
const ivec3 off = ivec3(-1,0,1);

    vec4 wave = texture(unit_wave, tex_coord);
    float s11 = wave.x;
    float s01 = textureOffset(unit_wave, tex_coord, off.xy).x;
    float s21 = textureOffset(unit_wave, tex_coord, off.zy).x;
    float s10 = textureOffset(unit_wave, tex_coord, off.yx).x;
    float s12 = textureOffset(unit_wave, tex_coord, off.yz).x;
    vec3 va = normalize(vec3(size.xy,s21-s01));
    vec3 vb = normalize(vec3(size.yx,s12-s10));
    vec4 bump = vec4( cross(va,vb), s11 );
*/

func ImageToHeightmap(img image.Image, scale float64) HeightMap {
	bounds := img.Bounds()
	heightmap := texturing.Empty[float64](bounds.Dx(), bounds.Dy())

	heightmap.MutateParallel(func(x, y int, v float64) float64 {
		return coloring.Greyscale(img.At(x, y)) * scale
	})

	return heightmap
}

// https://stackoverflow.com/questions/5281261/generating-a-normal-map-from-a-height-map
func FromHeightmap(heightmap HeightMap, scale float64) NormalMap {
	width, height := heightmap.Width(), heightmap.Height()
	dst := texturing.Empty[vector3.Float64](width, height)

	utils.ParallelRange(height, func(start, end int) {
		for y := start; y < end; y++ {
			above, below := mirrored(y+1, height), mirrored(y-1, height)
			for x := range width {
				left, right := mirrored(x-1, width), mirrored(x+1, width)

				va := vector3.New(2, 0, (heightmap.Get(right, y)-heightmap.Get(left, y))*scale).Normalized()
				vb := vector3.New(0, 2, (heightmap.Get(x, below)-heightmap.Get(x, above))*scale).Normalized()

				dst.Set(x, y, va.Cross(vb).Normalized())
			}
		}
	})

	return dst
}

func mirrored(i, n int) int {
	if n == 1 {
		return 0
	}
	if i < 0 {
		return 1
	}
	if i >= n {
		return n - 2
	}
	return i
}

func RasterizeNormalmap(normals NormalMap) image.Image {
	return normals.ToImage(func(n vector3.Float64) color.Color {
		return color.RGBA{
			R: uint8((0.5 + (n.X() / 2.)) * 255),
			G: uint8((0.5 + (n.Y() / 2.)) * 255),
			B: uint8((0.5 + (n.Z() / 2.)) * 255),
			A: 255,
		}
	})
}
