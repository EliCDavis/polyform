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

// https://stackoverflow.com/questions/5281261/generating-a-normal-map-from-a-height-map
func ImageFromHeightmap(heightmap image.Image, scale float64) *image.RGBA {
	dst := image.NewRGBA(heightmap.Bounds())

	texturing.ConvolveImage(heightmap, func(x, y int, values []color.Color) {
		// s11 := float64(colors.Red(values[4])) / 255

		s01 := float64(coloring.Red(values[3])) / 255.
		s21 := float64(coloring.Red(values[5])) / 255.
		s10 := float64(coloring.Red(values[1])) / 255.
		s12 := float64(coloring.Red(values[7])) / 255.

		va := vector3.New(2, 0, (s21-s01)*scale).Normalized()
		vb := vector3.New(0, 2, (s12-s10)*scale).Normalized()
		n := va.Cross(vb)

		dst.Set(x, y, color.RGBA{
			R: uint8((0.5 + (n.X() / 2.)) * 255),
			G: uint8((0.5 + (n.Y() / 2.)) * 255),
			B: uint8((0.5 + (n.Z() / 2.)) * 255),
			A: 255,
		})
	})

	return dst
}

func FromHeightmap(heightmap texturing.Texture[float64], scale float64) texturing.Texture[vector3.Float64] {
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

func ToNormalmap(normals texturing.Texture[vector3.Float64]) image.Image {
	return normals.ToImage(func(n vector3.Float64) color.Color {
		return color.RGBA{
			R: uint8((0.5 + (n.X() / 2.)) * 255),
			G: uint8((0.5 + (n.Y() / 2.)) * 255),
			B: uint8((0.5 + (n.Z() / 2.)) * 255),
			A: 255,
		}
	})
}
