package normals

import (
	"image"
	"image/color"

	"github.com/EliCDavis/polyform/drawing/coloring"
	"github.com/EliCDavis/polyform/drawing/texturing"
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
	dst := texturing.Empty[vector3.Float64](heightmap.Width(), heightmap.Height())

	texturing.Convolve(heightmap, func(x, y int, values []float64) {
		s01 := values[3]
		s21 := values[5]
		s10 := values[1]
		s12 := values[7]

		va := vector3.New(2, 0, (s21-s01)*scale).Normalized()
		vb := vector3.New(0, 2, (s12-s10)*scale).Normalized()

		dst.Set(x, y, va.Cross(vb).Normalized())
	})

	return dst
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
