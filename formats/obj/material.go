package obj

import (
	"fmt"
	"image/color"
	"io"
	"strings"

	"github.com/EliCDavis/polyform/formats/txt"
)

func writeMaterialColor(colorType string, color color.Color, writer *txt.Writer) error {
	if color == nil {
		return nil
	}
	writer.StartEntry()
	writer.String(colorType)
	writer.Space()

	r, g, b, _ := color.RGBA()
	writer.Float64MaxFigs(float64(r)/0xffff, 3)
	writer.Space()
	writer.Float64MaxFigs(float64(g)/0xffff, 3)
	writer.Space()
	writer.Float64MaxFigs(float64(b)/0xffff, 3)

	writer.NewLine()
	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", colorType, err)
	}
	return nil
}

func writeMaterialFloat(floatType string, f float64, writer *txt.Writer) error {
	writer.StartEntry()
	writer.String(floatType)
	writer.Space()
	writer.Float64(f)
	writer.NewLine()

	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", floatType, err)
	}
	return nil
}

func writeMaterialTexture(texType string, tex *string, writer *txt.Writer) error {
	if tex == nil {
		return nil
	}

	writer.StartEntry()

	writer.String(texType)
	writer.Space()
	writer.String(*tex)
	writer.NewLine()

	if _, err := writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write %s: %w", texType, err)
	}
	return nil
}

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
	//
	// Defines the focus of the specular highlight
	SpecularHighlight float64

	// Index of refraction, between 0.001 to 10, 1.0 means light does not bend
	// as it passes through the object
	OpticalDensity float64

	// Specifies how much this material dissolves into the background. A factor
	// of 0.0 is fully opaque. A factor of 1.0 is completely transparent.
	Transparency float64

	ColorTextureURI *string

	NormalTextureURI *string

	SpecularTextureURI *string
}

func (mat Material) write(out io.Writer) (err error) {
	writer := txt.NewWriter(out)

	writer.StartEntry()
	writer.String("newmtl ")
	writer.String(strings.Replace(mat.Name, " ", "", -1))
	writer.NewLine()
	if _, err = writer.FinishEntry(); err != nil {
		return fmt.Errorf("failed to write newmtl: %w", err)
	}

	if err = writeMaterialColor("Kd", mat.DiffuseColor, writer); err != nil {
		return err
	}

	if err = writeMaterialColor("Ka", mat.AmbientColor, writer); err != nil {
		return err
	}

	if err = writeMaterialColor("Ks", mat.SpecularColor, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("Ns", mat.SpecularHighlight, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("Ni", mat.OpticalDensity, writer); err != nil {
		return err
	}

	if err = writeMaterialFloat("d", 1-mat.Transparency, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Kd", mat.ColorTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Ks", mat.SpecularTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("map_Bump", mat.NormalTextureURI, writer); err != nil {
		return err
	}

	if err = writeMaterialTexture("norm", mat.NormalTextureURI, writer); err != nil {
		return err
	}

	writer.StartEntry()
	writer.NewLine()
	writer.FinishEntry()

	if err = writer.Error(); err != nil {
		return fmt.Errorf("failed to write out: %w", err)
	}
	return nil
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
		Name:              "Default Diffuse",
		AmbientColor:      color.Black,
		DiffuseColor:      c,
		SpecularColor:     color.Black,
		SpecularHighlight: 100,
		OpticalDensity:    1,
		Transparency:      0,
		ColorTextureURI:   nil,
	}
}
