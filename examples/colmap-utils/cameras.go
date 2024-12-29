package main

import (
	"fmt"

	"github.com/EliCDavis/sfm/colmap"
	"github.com/urfave/cli/v2"
)

var CamerasCommand = &cli.Command{
	Name:  "camera-info",
	Usage: "print information about cameras",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "cameras",
			Usage: "path to cameras file",
			Value: "cameras.bin",
		},
	},
	Action: func(ctx *cli.Context) error {
		cameras, err := colmap.LoadCamerasBinary(ctx.String("cameras"))
		if err != nil {
			return err
		}

		for _, cam := range cameras {
			fmt.Fprintf(ctx.App.Writer, "Camera [%d] - %s (%dx%d)\n", cam.ID, cam.Model.String(), cam.Width, cam.Height)

			switch cam.Model {
			case colmap.SIMPLE_PINHOLE:
				model := colmap.SimplePinholeCamera(cam)
				fmt.Fprintf(ctx.App.Writer, "\tFocal Length %f\n", model.FocalLength())
				fmt.Fprintf(ctx.App.Writer, "\tCx %f\n", model.Cx())
				fmt.Fprintf(ctx.App.Writer, "\tCy %f\n", model.Cy())

			case colmap.PINHOLE:
				model := colmap.PinholeCamera(cam)
				fmt.Fprintf(ctx.App.Writer, "\tFx %f\n", model.Fx())
				fmt.Fprintf(ctx.App.Writer, "\tFy %f\n", model.Fy())
				fmt.Fprintf(ctx.App.Writer, "\tCx %f\n", model.Cx())
				fmt.Fprintf(ctx.App.Writer, "\tCy %f\n", model.Cy())

			case colmap.SIMPLE_RADIAL:
				model := colmap.SimpleRadialCamera(cam)
				fmt.Fprintf(ctx.App.Writer, "\tFocal Length %f\n", model.FocalLength())
				fmt.Fprintf(ctx.App.Writer, "\tCx %f\n", model.Cx())
				fmt.Fprintf(ctx.App.Writer, "\tCy %f\n", model.Cy())
				fmt.Fprintf(ctx.App.Writer, "\tK  %f\n", model.K())

			case colmap.RADIAL:
				model := colmap.RadialCamera(cam)
				fmt.Fprintf(ctx.App.Writer, "\tFocal Length %f\n", model.FocalLength())
				fmt.Fprintf(ctx.App.Writer, "\tCx %f\n", model.Cx())
				fmt.Fprintf(ctx.App.Writer, "\tCy %f\n", model.Cy())
				fmt.Fprintf(ctx.App.Writer, "\tK1 %f\n", model.K1())
				fmt.Fprintf(ctx.App.Writer, "\tK2 %f\n", model.K2())

			case colmap.OPENCV:
				model := colmap.OpenCVCamera(cam)
				fmt.Fprintf(ctx.App.Writer, "\tFx %f\n", model.Fx())
				fmt.Fprintf(ctx.App.Writer, "\tFy %f\n", model.Fy())
				fmt.Fprintf(ctx.App.Writer, "\tCx %f\n", model.Cx())
				fmt.Fprintf(ctx.App.Writer, "\tCy %f\n", model.Cy())
				fmt.Fprintf(ctx.App.Writer, "\tK1 %f\n", model.K1())
				fmt.Fprintf(ctx.App.Writer, "\tK2 %f\n", model.K2())
				fmt.Fprintf(ctx.App.Writer, "\tP1 %f\n", model.P1())
				fmt.Fprintf(ctx.App.Writer, "\tP2 %f\n", model.P2())

			default:
				fmt.Fprintf(ctx.App.Writer, "\t%v\n", cam.Params)
			}
		}
		return nil
	},
}
