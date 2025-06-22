package main

import (
	"bufio"
	"encoding/binary"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

func csvHeader(csvPath string) ([]string, error) {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return nil, fmt.Errorf("unable to open CSV file %q: %w", csvPath, err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)

	return reader.Read()
}

func scanCSV(csvPath string, f func(entries []string) error) error {
	csvFile, err := os.Open(csvPath)
	if err != nil {
		return fmt.Errorf("unable to open CSV file %q: %w", csvPath, err)
	}
	defer csvFile.Close()
	reader := csv.NewReader(csvFile)

	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("unable to read csv header: %w", err)
	}

	for {
		entries, err := reader.Read()
		if err == io.EOF {
			return nil
		}

		if err != nil {
			return fmt.Errorf("unable to read row: %w", err)
		}

		err = f(entries)
		if err != nil {
			return err
		}
	}

}

var FromCSVCommand = &cli.Command{
	Name:  "from-csv",
	Usage: "converts a CSV file to PLY",
	Args:  true,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Usage:   "Path to write the PLY to",
			Aliases: []string{"o"},
			Value:   "out.ply",
		},
	},
	Action: func(ctx *cli.Context) error {
		start := time.Now()
		if ctx.Args().Len() != 1 {
			return errors.New("expected single argument for CSV path")
		}

		csvPath := ctx.Args().First()

		csvHeader, err := csvHeader(csvPath)
		if err != nil {
			return fmt.Errorf("unable to read csv header from %q: %w", csvHeader, err)
		}

		// Get all columns that have numerical values
		validColumns := make([]bool, len(csvHeader))
		rows := 0
		err = scanCSV(csvPath, func(entries []string) error {
			rows++
			for i, entry := range entries {
				if validColumns[i] {
					continue
				}

				_, err = strconv.ParseFloat(entry, 64)
				if err == nil {
					validColumns[i] = true
				}
			}
			return nil
		})

		if err != nil {
			return err
		}

		// Get all rows where all columns have a numerical value
		validRows := 0
		err = scanCSV(csvPath, func(entries []string) error {
			valid := true
			for i, entry := range entries {
				if !validColumns[i] {
					continue
				}

				_, err = strconv.ParseFloat(entry, 64)
				if err != nil {
					valid = false
					break
				}
			}

			if valid {
				validRows++
			}
			return nil
		})

		properties := make([]ply.Property, len(csvHeader))
		for i, entry := range csvHeader {
			if !validColumns[i] {
				continue
			}
			properties[i] = ply.ScalarProperty{
				PropertyName: strings.Replace(entry, " ", "_", -1),
				Type:         ply.Double,
			}
		}

		plyHeader := ply.Header{
			Format: ply.BinaryLittleEndian,
			Elements: []ply.Element{{
				Name:       ply.VertexElementName,
				Properties: properties,
				Count:      int64(validRows),
			}},
		}

		plyPath := ctx.String("out")
		outFile, err := os.Create(plyPath)
		if err != nil {
			return fmt.Errorf("unable to create PLY file %q: %w", plyPath, err)
		}
		defer outFile.Close()

		writer := bufio.NewWriter(outFile)
		if err = plyHeader.Write(writer); err != nil {
			return fmt.Errorf("unable to write PLY header: %w", err)
		}

		buf := make([]byte, len(properties)*8)
		err = scanCSV(csvPath, func(entries []string) error {
			valid := true
			for i, entry := range entries {
				if !validColumns[i] {
					continue
				}

				f, err := strconv.ParseFloat(entry, 64)
				if err != nil {
					valid = false
					break
				}

				bits := math.Float64bits(f)
				binary.LittleEndian.PutUint64(buf[i*8:], bits)
			}

			if valid {
				_, err = writer.Write(buf)
				return err
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("unable to populate PLY file %w", err)
		}

		log.Printf("Wrote %d of %d rows to %q in %s\n", validRows, rows, plyPath, time.Since(start))

		return writer.Flush()
	},
}
