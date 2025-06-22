package properties

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/EliCDavis/polyform/examples/ply-utils/flags"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

func addPropertyWorker(
	inPly, outPly string,
	outHeaderOffset int,
	inPointSize, outPointSize int,
	pointIndex, pointCount int64,
	wg *sync.WaitGroup,
) {
	oldPly, err := os.Open(inPly)
	check(err)
	defer oldPly.Close()
	_, err = ply.ReadHeader(oldPly)
	check(err)
	_, err = oldPly.Seek((int64(inPointSize) * pointIndex), io.SeekCurrent)
	check(err)

	newPly, err := os.OpenFile(outPly, os.O_WRONLY, 0)
	check(err)
	defer newPly.Close()
	_, err = newPly.Seek(int64(outHeaderOffset)+(int64(outPointSize)*pointIndex), io.SeekStart)
	check(err)

	oldPlyBuf := make([]byte, inPointSize)
	newPlyBuf := make([]byte, outPointSize)

	reader := bufio.NewReader(oldPly)
	writer := bufio.NewWriter(newPly)

	for i := int64(0); i < pointCount; i++ {
		_, err := io.ReadFull(reader, oldPlyBuf)
		check(err)

		copy(newPlyBuf, oldPlyBuf)

		_, err = writer.Write(newPlyBuf)
		check(err)
	}
	writer.Flush()
	wg.Done()
}

func PropertiesToAddFromArguments(args []string) []ply.Property {
	if r := len(args) % 2; r != 0 {
		panic("invalid num of args")
	}

	props := make([]ply.Property, 0)
	for i := 0; i < len(args); i += 2 {
		props = append(props, ply.ScalarProperty{
			PropertyName: args[i],
			Type:         ply.ParseScalarPropertyType(args[i+1]),
		})
	}

	return props
}

var addPropertiesCommand = &cli.Command{
	Name:      "add",
	Usage:     "Create a new ply file with new properties",
	Args:      true,
	ArgsUsage: "[{property name, property type}]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
		},
		&cli.BoolFlag{
			Name:  "force",
			Usage: "whether or not to overwrite a already existing output file",
		},
	},
	Action: func(ctx *cli.Context) error {
		outPath, err := getOutpath(ctx)
		if err != nil {
			return err
		}

		f, err := flags.OpenPlyFile()
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := ply.ReadHeader(f)
		if err != nil {
			return err
		}

		if len(header.Elements) > 1 {
			return fmt.Errorf("unsupposrted situation where ply has %d elements. Feel free to open up a PR", len(header.Elements))
		}

		if len(header.Elements) == 0 {
			return fmt.Errorf("empty ply file? whatcha doin bud")
		}

		if header.Format == ply.ASCII {
			return errors.New("asccii format unsupported (we like to go fast around here)")
		}

		propertiesToAdd := PropertiesToAddFromArguments(ctx.Args().Slice())

		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()

		originalPointSize := calculateTotalPropertySize(header.Elements[0].Properties)
		propertiesToAddSize := calculateTotalPropertySize(propertiesToAdd)
		newPointSize := originalPointSize + propertiesToAddSize

		header.Elements[0].Properties = append(header.Elements[0].Properties, propertiesToAdd...)
		headerBytes := header.Bytes()
		totalPointCount := header.Elements[0].Count
		err = out.Truncate(int64(len(headerBytes)) + (int64(newPointSize) * totalPointCount))
		if err != nil {
			return err
		}

		writer := bufio.NewWriter(out)
		_, err = writer.Write(headerBytes)
		if err != nil {
			return err
		}

		fmt.Fprintf(ctx.App.Writer, "Old Point Size: %d bytes\n", originalPointSize)
		fmt.Fprintf(ctx.App.Writer, "New Point Size: %d bytes\n", newPointSize)
		addition := (float64(newPointSize) / float64(originalPointSize)) - 1
		fmt.Fprintf(
			ctx.App.Writer,
			"%.2f%% addition across %d points creating %s of data\n",
			addition*100.,
			totalPointCount,
			dataSizeFormat(totalPointCount*int64(newPointSize-originalPointSize)),
		)

		start := time.Now()
		if totalPointCount < 10000 {
			oldPlyBuf := make([]byte, originalPointSize)
			newPlyBuf := make([]byte, newPointSize)

			reader := bufio.NewReader(f)
			for i := int64(0); i < totalPointCount; i++ {
				_, err = io.ReadFull(reader, oldPlyBuf)
				if err != nil {
					return err
				}

				copy(newPlyBuf, oldPlyBuf)

				_, err := writer.Write(newPlyBuf)
				if err != nil {
					return err
				}
			}

			err = writer.Flush()
			if err != nil {
				return err
			}
		} else {
			workers := runtime.NumCPU()
			pointsPerWorker := math.Floor(float64(totalPointCount) / float64(workers))

			wg := &sync.WaitGroup{}
			wg.Add(workers)
			for w := 0; w < workers; w++ {
				pc := int64(pointsPerWorker)
				startingPoint := int64(pointsPerWorker) * int64(w)
				if w == workers-1 {
					pc = totalPointCount - startingPoint
				}
				go addPropertyWorker(
					ctx.String("in"),
					ctx.String("out"),
					len(headerBytes),
					originalPointSize,
					newPointSize,
					startingPoint,
					pc,
					wg,
				)
			}

			wg.Wait()
		}

		duration := time.Since(start)
		_, err = fmt.Fprintf(ctx.App.Writer, "PLY written in %s (%.2f points a second)", duration, float64(totalPointCount)/duration.Seconds())
		return err
	},
}
