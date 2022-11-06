package mesh

import "image/color"

// Material is just a clone of obj's MTL format at the moment cause man this
// problem scares me.
type Material struct {
	Name string

	// Account for light that is scattered about the entire scene
	AmbientColor color.Color

	// The main color
	DiffuseColor color.Color

	// Color seen where the surface is shiny and mirror like
	SpecularColor color.Color

	// Typically between 0 - 1000, with a high value resulting in a tight,
	// concentrated highlight
	SpecularHighlight float64

	// Index of refraction, between 0.001 to 10, 1.0 means light does not bend
	// as it passes through the object
	OpticalDensity float64

	// Specifies how much this material dissolves into the background. A factor
	// of 0.0 is fully opaque. A factor of 1.0 is completely transparent.
	Transparency float64

	ColorTextureURI *string
}

func DefaultMaterial() Material {
	return Material{
		Name:              "Default Diffuse",
		AmbientColor:      color.Black,
		DiffuseColor:      color.White,
		SpecularColor:     color.Black,
		SpecularHighlight: 100,
		OpticalDensity:    1,
		Transparency:      0,
		ColorTextureURI:   nil,
	}
}

func DefaultColorMaterial(c color.Color) Material {
	return Material{
		Name:              "DefaultDiffuse",
		AmbientColor:      color.Black,
		DiffuseColor:      c,
		SpecularColor:     color.Black,
		SpecularHighlight: 100,
		OpticalDensity:    1,
		Transparency:      0,
		ColorTextureURI:   nil,
	}
}
