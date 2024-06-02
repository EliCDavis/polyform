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

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

type Blocks struct {
	start, offset int
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func removePropertyWorker(
	inPly, outPly string,
	outHeaderOffset int,
	inPointSize, outPointSize int,
	pointIndex, pointCount int64,
	blocks []Blocks,
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

		start := 0
		for _, b := range blocks {
			size := b.offset - b.start
			copy(newPlyBuf[start:start+size], oldPlyBuf[b.start:b.offset])
			start += size
		}

		_, err = writer.Write(newPlyBuf)
		check(err)
	}
	writer.Flush()
	wg.Done()
}

var removePropertiesCommand = &cli.Command{
	Name:      "remove",
	Usage:     "Create a new ply file with specific poroperties removed",
	Args:      true,
	ArgsUsage: "[names of the properties to remove]",
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

		f, err := openPlyFile(ctx)
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

		columnsToRmove := ctx.Args().Slice()

		shifts := make([]Blocks, 0)
		currentStart := 0
		currentOffset := 0
		newPointSize := 0
		remainingProps := make([]ply.Property, 0)
		for _, p := range header.Elements[0].Properties {
			scalar, ok := p.(ply.ScalarProperty)
			if !ok {
				return errors.New("unsupposrted situation where element has array property. Feel free to open up a PR")
			}

			removed := false

			for _, col := range columnsToRmove {
				if scalar.PropertyName == col {
					removed = true

					if currentStart != currentOffset {
						shifts = append(shifts, Blocks{
							start:  currentStart,
							offset: currentOffset,
						})
					}

					currentStart = currentOffset + scalar.Size()
				}
			}

			if !removed {
				newPointSize += scalar.Size()
				remainingProps = append(remainingProps, p)
			}

			currentOffset += scalar.Size()
		}

		if currentStart != currentOffset {
			shifts = append(shifts, Blocks{
				start:  currentStart,
				offset: currentOffset,
			})
		}

		out, err := os.Create(outPath)
		if err != nil {
			return err
		}
		defer out.Close()

		header.Elements[0].Properties = remainingProps
		headerBytes := header.Bytes()
		totalPointCount := header.Elements[0].Count
		out.Truncate(int64(len(headerBytes)) + (int64(newPointSize) * totalPointCount))
		writer := bufio.NewWriter(out)
		_, err = writer.Write(headerBytes)
		if err != nil {
			return err
		}

		fmt.Fprintf(ctx.App.Writer, "Old Point Size: %d bytes\n", currentOffset)
		fmt.Fprintf(ctx.App.Writer, "New Point Size: %d bytes\n", newPointSize)
		reduction := 1 - (float64(newPointSize) / float64(currentOffset))
		fmt.Fprintf(
			ctx.App.Writer,
			"%.2f%% reduction across %d points removing %s of data\n",
			reduction*100.,
			totalPointCount,
			dataSizeFormat(totalPointCount*int64(currentOffset-newPointSize)),
		)

		start := time.Now()
		if totalPointCount < 10000 {
			oldPlyBuf := make([]byte, currentOffset)
			newPlyBuf := make([]byte, newPointSize)

			reader := bufio.NewReader(f)
			for i := int64(0); i < totalPointCount; i++ {
				_, err = io.ReadFull(reader, oldPlyBuf)
				if err != nil {
					return err
				}

				start := 0
				for _, b := range shifts {
					size := b.offset - b.start
					copy(newPlyBuf[start:start+size], oldPlyBuf[b.start:b.offset])
					start += size
				}

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
				go removePropertyWorker(
					ctx.String("in"),
					ctx.String("out"),
					len(headerBytes),
					currentOffset,
					newPointSize,
					startingPoint,
					pc,
					shifts,
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
