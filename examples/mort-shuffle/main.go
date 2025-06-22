package main

import (
	"errors"
	"os"

	"github.com/EliCDavis/polyform/formats/gltf"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/modeling/meshops"
	"github.com/urfave/cli/v2"
)

var warpSizeFlag = "warp-size"

func main() {
	cli := cli.App{
		Name: "mort-shuffle",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "in",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "out",
				Value: "mort-shuffle.glb",
			},
			&cli.IntFlag{
				Name:  warpSizeFlag,
				Value: 32,
			},
			&cli.UintFlag{
				Name:  "resolution",
				Usage: "How many bits each component take up in the morton index (max 21)",
				Value: 10,
				Action: func(ctx *cli.Context, u uint) error {
					if u == 0 {
						return errors.New("Value must be atleast 1")
					}

					if u > 21 {
						return errors.New("resolution can not be greater than 21")
					}

					return nil
				},
			},
		},
		Action: func(ctx *cli.Context) error {
			in, err := ply.Load(ctx.String("in"))
			if err != nil {
				return err
			}

			// rand.Shuffle(len(newPositions), func(i, j int) {
			// 	newPositions[i], newPositions[j] = newPositions[j], newPositions[i]
			// })

			out := meshops.MortonShuffle(
				*in,
				modeling.PositionAttribute,
				ctx.Int(warpSizeFlag),
				ctx.Uint("resolution"),
			)

			return gltf.SaveBinary(ctx.String("out"), gltf.PolyformScene{
				Models: []gltf.PolyformModel{{Mesh: &out}},
			})
		},
	}

	if err := cli.Run(os.Args); err != nil {
		panic(err)
	}
}
