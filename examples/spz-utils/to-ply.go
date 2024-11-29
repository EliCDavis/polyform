package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strings"

	"github.com/EliCDavis/iter"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/formats/spz"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
	"github.com/urfave/cli/v2"
)

func lowerNoSpace(s string) ColorPropertyFormat {
	return ColorPropertyFormat(strings.TrimSpace(strings.ToLower(s)))
}

func plyFormatEnum(s string) PropertyFormat {
	return PropertyFormat(strings.TrimSpace(strings.ToLower(s)))
}

type ColorPropertyFormat string

const (
	RGB ColorPropertyFormat = "rgb"
	FDC ColorPropertyFormat = "f_dc"
)

type PropertyFormat string

const (
	Splat PropertyFormat = "splat"
	SPZ   PropertyFormat = "spz"
)

var ToPlyCommand = &cli.Command{
	Name: "to-ply",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:     "out",
			Usage:    "path to write ply data to",
			Aliases:  []string{"o"},
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "spherical-harmonics",
			Usage: "whether or not to include spherical harmonics",
			Value: true,
		},
		&cli.StringFlag{
			Name:  "color",
			Usage: "PLY Color Proptery format: [f_dc, rgb]",
			Value: "f_dc",
			Action: func(ctx *cli.Context, s string) error {
				v := lowerNoSpace(s)
				if v == RGB || v == FDC {
					return nil
				}
				return fmt.Errorf("unrecognized color format: '%s'", v)
			},
		},
		&cli.StringFlag{
			Name:  "format",
			Usage: "PLY Proptery format: [splat, spz]",
			Value: "splat",
			Action: func(ctx *cli.Context, s string) error {
				v := plyFormatEnum(s)
				if v == SPZ || v == Splat {
					return nil
				}
				return fmt.Errorf("unrecognized property format: '%s'", v)
			},
		},
	},
	Action: func(ctx *cli.Context) error {

		includeHarmonics := ctx.Bool("spherical-harmonics")
		propertyFormat := plyFormatEnum(ctx.String("format"))

		cloud, err := spz.Load(inFilePath)
		if err != nil {
			return err
		}

		plyFile, err := os.Create(ctx.String("out"))
		if err != nil {
			return err
		}
		defer plyFile.Close()
		out := bufio.NewWriter(plyFile)

		props := []ply.Property{
			ply.ScalarProperty{PropertyName: "x", Type: ply.Float},
			ply.ScalarProperty{PropertyName: "y", Type: ply.Float},
			ply.ScalarProperty{PropertyName: "z", Type: ply.Float},
		}

		var scalarType ply.ScalarPropertyType
		switch propertyFormat {
		case SPZ:
			scalarType = ply.UChar

		case Splat:
			scalarType = ply.Float
		}

		switch lowerNoSpace(ctx.String("color")) {
		case RGB:
			props = append(
				props,
				ply.ScalarProperty{PropertyName: "r", Type: scalarType},
				ply.ScalarProperty{PropertyName: "g", Type: scalarType},
				ply.ScalarProperty{PropertyName: "b", Type: scalarType},
			)

		case FDC:
			props = append(
				props,
				ply.ScalarProperty{PropertyName: "f_dc_0", Type: scalarType},
				ply.ScalarProperty{PropertyName: "f_dc_1", Type: scalarType},
				ply.ScalarProperty{PropertyName: "f_dc_2", Type: scalarType},
			)
		}

		props = append(
			props,
			ply.ScalarProperty{PropertyName: "scale_0", Type: scalarType},
			ply.ScalarProperty{PropertyName: "scale_1", Type: scalarType},
			ply.ScalarProperty{PropertyName: "scale_2", Type: scalarType},
			ply.ScalarProperty{PropertyName: "opacity", Type: scalarType},
			ply.ScalarProperty{PropertyName: "rot_0", Type: scalarType},
			ply.ScalarProperty{PropertyName: "rot_1", Type: scalarType},
			ply.ScalarProperty{PropertyName: "rot_2", Type: scalarType},
		)

		if propertyFormat == Splat {
			props = append(props, ply.ScalarProperty{PropertyName: "rot_3", Type: scalarType})
		}

		cloudDimensions, err := cloud.Header.ShDimensions()
		if err != nil {
			return err
		}

		shArrays := make([]*iter.ArrayIterator[vector3.Float64], cloudDimensions)
		if includeHarmonics {
			for i := 0; i < cloudDimensions; i++ {
				i3 := i * 3
				props = append(props, ply.ScalarProperty{
					Type:         scalarType,
					PropertyName: fmt.Sprintf("f_rest_%d", i3),
				})
				props = append(props, ply.ScalarProperty{
					Type:         scalarType,
					PropertyName: fmt.Sprintf("f_rest_%d", i3+1),
				})
				props = append(props, ply.ScalarProperty{
					Type:         scalarType,
					PropertyName: fmt.Sprintf("f_rest_%d", i3+2),
				})
				shArrays[i] = cloud.Mesh.Float3Attribute(fmt.Sprintf("SH_%d", i))
			}
		}
		endian := binary.LittleEndian
		header := ply.Header{
			Format: ply.BinaryLittleEndian,
			Elements: []ply.Element{
				{
					Name:       "vertex",
					Count:      int64(cloud.Header.NumPoints),
					Properties: props,
				},
			},
		}

		err = header.Write(out)
		if err != nil {
			return err
		}

		scales := cloud.Mesh.Float3Attribute(modeling.ScaleAttribute)
		colors := cloud.Mesh.Float3Attribute(modeling.FDCAttribute)
		positions := cloud.Mesh.Float3Attribute(modeling.PositionAttribute)
		rotations := cloud.Mesh.Float4Attribute(modeling.RotationAttribute)
		alphas := cloud.Mesh.Float1Attribute(modeling.OpacityAttribute)

		switch propertyFormat {
		case SPZ:
			for i := 0; i < int(cloud.Header.NumPoints); i++ {
				positions.
					At(i).
					ToFloat32().
					Write(out, endian)

				color := colors.At(i).Clamp(0, 1).Scale(255)
				binary.Write(out, endian, byte(color.X()))
				binary.Write(out, endian, byte(color.Y()))
				binary.Write(out, endian, byte(color.Z()))

				scale := scales.At(i).Clamp(0, 1).Scale(255)
				binary.Write(out, endian, byte(scale.X()))
				binary.Write(out, endian, byte(scale.Y()))
				binary.Write(out, endian, byte(scale.Z()))

				binary.Write(out, endian, byte(alphas.At(i)*255))

				rotation := rotations.At(i).Clamp(0, 1).Scale(255)
				binary.Write(out, endian, byte(rotation.X()))
				binary.Write(out, endian, byte(rotation.Y()))
				binary.Write(out, endian, byte(rotation.Z()))

				if includeHarmonics {
					for _, arr := range shArrays {
						sh := arr.At(i).Clamp(0, 1).Scale(255)
						binary.Write(out, endian, byte(sh.X()))
						binary.Write(out, endian, byte(sh.Y()))
						binary.Write(out, endian, byte(sh.Z()))
					}
				}
			}

		case Splat:
			for i := 0; i < int(cloud.Header.NumPoints); i++ {
				positions.At(i).ToFloat32().Write(out, endian)
				colors.At(i).Clamp(0, 1).ToFloat32().Write(out, endian)
				scales.At(i).Clamp(0, 1).ToFloat32().Write(out, endian)
				binary.Write(out, endian, float32(alphas.At(i)))
				rotations.At(i).Clamp(0, 1).ToFloat32().Write(out, endian)

				if includeHarmonics {
					for _, arr := range shArrays {
						arr.At(i).ToFloat32().Write(out, endian)
					}
				}
			}
		}

		return out.Flush()
	},
}
