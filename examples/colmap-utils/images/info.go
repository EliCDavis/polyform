package images

import (
	"fmt"

	"github.com/EliCDavis/sfm/colmap"
	"github.com/urfave/cli/v2"
)

var infoCommand = &cli.Command{
	Name:  "info",
	Usage: "Print info pertaining to the reconstructed image data",
	Action: func(ctx *cli.Context) error {
		images, err := colmap.LoadImagesBinary(ctx.String(imagesPathFlagName))
		if err != nil {
			return err
		}

		for _, img := range images {
			fmt.Fprintf(ctx.App.Writer, "[%d] %s\n", img.Id, img.Name)
			fmt.Fprintf(ctx.App.Writer, "\tCamera Id:   %d\n", img.CameraId)
			fmt.Fprintf(ctx.App.Writer, "\tTranslation: %s\n", img.Translation.Format("%f, %f, %f"))
			fmt.Fprintf(ctx.App.Writer, "\tRotation:    %s\n", img.Rotation.Format("%f, %f, %f, %f"))
			fmt.Fprintf(ctx.App.Writer, "\tPoint Count: %d\n\n", len(img.Points))
		}

		return nil
	},
}
