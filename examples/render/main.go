package main

import (
	"errors"
	"fmt"
	"image/png"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/formats/obj"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/rendering"
	"github.com/EliCDavis/polyform/rendering/materials"
	"github.com/EliCDavis/polyform/rendering/textures"
	"github.com/EliCDavis/vector/vector3"
	"github.com/schollz/progressbar/v3"
	"github.com/urfave/cli/v2"
)

func readMesh(meshPath string) (*modeling.Mesh, error) {
	ext := filepath.Ext(meshPath)

	inFile, err := os.Open(meshPath)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	switch strings.ToLower(ext) {

	case ".obj":
		scene, err := obj.Load(meshPath)
		mesh := scene.ToMesh()
		return &mesh, err

	case ".ply":
		return ply.ReadMesh(inFile)

	default:
		return nil, fmt.Errorf("unimplemented format: %s", ext)
	}
}

func fileExits(filePath string) bool {
	if _, err := os.Stat(filePath); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func inferRenderingMaterial(originalPath string, mat *obj.Material) (rendering.Material, error) {
	defaultMat := materials.NewLambertian(textures.NewSolidColorTexture(vector3.Fill(0.7)))
	if mat == nil {
		return defaultMat, nil
	}

	if mat.ColorTextureURI == nil {
		r, g, b, _ := mat.DiffuseColor.RGBA()

		return materials.NewLambertian(textures.NewSolidColorTexture(vector3.New(
			float64(r)/float64(0xffff),
			float64(g)/float64(0xffff),
			float64(b)/float64(0xffff),
		))), nil
	}

	imagePath := path.Join(path.Dir(originalPath), *mat.ColorTextureURI)
	if !fileExits(imagePath) {
		return nil, fmt.Errorf("(%s) references a image at (%s), but does not exist", originalPath, imagePath)
	}

	imgFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}

	img, err := png.Decode(imgFile)
	if err != nil {
		return nil, err
	}
	return materials.NewLambertian(textures.NewImage(img)), nil
}

func makeRenderFunction(name string, maxRayBounce, samplesPerPixel, imageWidth int) *cli.Command {
	return &cli.Command{
		Name: name,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Aliases:  []string{"i"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "out",
				Aliases: []string{"o"},
				Value:   fmt.Sprintf("%s-render.png", name),
			},
			&cli.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Value:   false,
			},
			&cli.BoolFlag{
				Name:    "progress",
				Usage:   "Whether or not to show a bar to indicate what percentage of the image has been rendered",
				Aliases: []string{"p"},
				Value:   true,
			},
			&cli.IntFlag{
				Name:  "max-ray-bounce",
				Usage: "Max number of times a single ray can bounce before it's terminated",
				Value: maxRayBounce,
			},
			&cli.IntFlag{
				Name:  "samples-per-pixel",
				Usage: "Number of rays to send per pixel to average for a final color value",
				Value: samplesPerPixel,
			},
			&cli.IntFlag{
				Name:  "image-size",
				Usage: "Number of pixels for image width/height",
				Value: imageWidth,
			},
		},
		Action: func(ctx *cli.Context) error {
			modelPath := ctx.String("in")

			mesh, err := readMesh(modelPath)
			if err != nil {
				return err
			}

			if fileExits(ctx.String("out")) {
				if ctx.Bool("force") {
					err := os.Remove(ctx.String("out"))
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("the file (%s) already exists, use -f to overwrite", ctx.String("out"))
				}
			}

			var mats rendering.Material = materials.NewLambertian(textures.NewSolidColorTexture(vector3.Fill(0.7)))

			box := mesh.BoundingBox(modeling.PositionAttribute)
			centeredMesh := mesh.Translate(box.Center().Scale(-1))

			scene := []rendering.Hittable{
				rendering.NewMesh(centeredMesh, mats),
			}

			size := box.Size()

			origin := vector3.New(-size.X()*.61, size.Y()*0.25, -size.Z()*.61)
			lookat := vector3.Zero[float64]()
			camera := rendering.NewDefaultCamera(1, origin, lookat, 0, 0)

			if ctx.Bool("progress") {
				bar := progressbar.Default(100, "Rendering Image")
				completion := make(chan float64, 1)
				go func() {
					err = rendering.RenderToFile(
						ctx.Int("max-ray-bounce"),
						ctx.Int("samples-per-pixel"),
						ctx.Int("image-size"),
						scene,
						camera,
						ctx.String("out"),
						completion,
					)
				}()

				lastUpdate := time.Now()
				for progress := range completion {
					if time.Since(lastUpdate) > time.Second {
						bar.Set(int(progress * 100.))
						lastUpdate = time.Now()
					}
				}
				bar.Set(100)
				bar.Finish()

				return err
			}

			return rendering.RenderToFile(
				ctx.Int("max-ray-bounce"),
				ctx.Int("samples-per-pixel"),
				ctx.Int("image-size"),
				scene,
				camera,
				ctx.String("out"),
				nil,
			)
		},
	}
}

func main() {
	app := &cli.App{
		Name: "render",
		Authors: []*cli.Author{
			{
				Name:  "Eli Davis",
				Email: "eli@recolude.com",
			},
		},
		Commands: []*cli.Command{
			makeRenderFunction("preview", 10, 25, 250),
			makeRenderFunction("basic", 50, 100, 500),
			makeRenderFunction("high-def", 100, 400, 2000),
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
