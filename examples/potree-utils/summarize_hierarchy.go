package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"

	"github.com/EliCDavis/polyform/formats/potree"
	"github.com/urfave/cli/v2"
)

type levelSummary struct {
	Nodes   int     `json:"nodes"`
	Average float64 `json:"average"`
	Min     int     `json:"min"`
	Max     int     `json:"max"`
	Total   int     `json:"total"`
	Volume  float64 `json:"volume"`
	Spacing float64 `json:"spacing"`
}

func writeSummaryAsCSV(out io.Writer, summaries []levelSummary) error {
	writer := csv.NewWriter(out)
	err := writer.Write([]string{
		"Level",
		"Nodes",
		"Average",
		"Min",
		"Max",
		"Total",
		"Volume",
		"Spacing",
	})
	if err != nil {
		return err
	}

	for i, summary := range summaries {
		err = writer.Write([]string{
			strconv.Itoa(i),
			strconv.Itoa(summary.Nodes),
			strconv.FormatFloat(summary.Average, 'f', -1, 64),
			strconv.Itoa(summary.Min),
			strconv.Itoa(summary.Max),
			strconv.Itoa(summary.Total),
			strconv.FormatFloat(summary.Volume, 'f', -1, 64),
			strconv.FormatFloat(summary.Spacing, 'f', -1, 64),
		})
		if err != nil {
			return err
		}
	}
	writer.Flush()
	return nil
}

func writeSummaryAsMarkdown(out io.Writer, summaries []levelSummary) error {
	_, err := fmt.Fprintf(
		out,
		"| %6s | %8s | %8s | %8s | %8s | %12s | %15s | %12s |\n",
		"Level",
		"Nodes",
		"Average",
		"Min",
		"Max",
		"Total",
		"Volume",
		"Spacing",
	)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(out, "|--------|----------|----------|----------|----------|--------------|-----------------|--------------|")
	if err != nil {
		return err
	}

	for i, summary := range summaries {
		_, err = fmt.Fprintf(
			out,
			"| %6d | %8d | %8d | %8d | %8d | %12d | %15.4f | %12.7f |\n",
			i,
			summary.Nodes,
			int(summary.Average),
			summary.Min,
			summary.Max,
			summary.Total,
			summary.Volume,
			summary.Spacing,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

var SummarizeHierarchyCommand = &cli.Command{
	Name:  "summarize",
	Usage: "Builds a summary of the hierarchy data",
	Flags: []cli.Flag{
		metadataFlag,
		hierarchyFlag,
		&cli.StringFlag{
			Name:  "format",
			Usage: "format to write summary data too (markdown, json, csv)",
			Value: "markdown",
		},
		&cli.StringFlag{
			Name:  "out",
			Usage: "path to file to write output too",
		},
		&cli.BoolFlag{
			Name:  "include-zero-point-nodes",
			Usage: "Whether or not to include nodes that have 0 points in our calculations",
			Value: false,
		},
	},
	Action: func(ctx *cli.Context) error {
		_, hierarchy, err := loadHierarchy(ctx)
		if err != nil {
			return err
		}

		pointCountsPerLevel := make(map[int][]int)
		totalPoints := make(map[int]int)
		volume := make(map[int]float64)
		spacing := make(map[int]float64)
		ignoreZero := !ctx.Bool("include-zero-point-nodes")
		hierarchy.Walk(func(o *potree.OctreeNode) bool {
			if ignoreZero && o.NumPoints == 0 {
				return true
			}

			if _, ok := pointCountsPerLevel[o.Level]; !ok {
				pointCountsPerLevel[o.Level] = make([]int, 0, 1)
				volume[o.Level] = o.BoundingBox.Volume()
				spacing[o.Level] = o.Spacing
			}
			pointCountsPerLevel[o.Level] = append(pointCountsPerLevel[o.Level], int(o.NumPoints))
			totalPoints[o.Level] += int(o.NumPoints)
			return true
		})

		averagePoints := make(map[int]float64)
		minPoints := make(map[int]int)
		maxPoints := make(map[int]int)

		for level, entries := range pointCountsPerLevel {
			curMin := math.MaxInt
			curMax := 0
			for _, e := range entries {
				curMin = min(curMin, e)
				curMax = max(curMax, e)
			}
			minPoints[level] = curMin
			maxPoints[level] = curMax
			averagePoints[level] = float64(totalPoints[level]) / float64(len(entries))
		}

		summaries := make([]levelSummary, len(pointCountsPerLevel))
		for i := 0; i < len(pointCountsPerLevel); i++ {
			summaries[i] = levelSummary{
				Nodes:   len(pointCountsPerLevel[i]),
				Average: averagePoints[i],
				Min:     minPoints[i],
				Max:     maxPoints[i],
				Total:   totalPoints[i],
				Volume:  volume[i],
				Spacing: spacing[i],
			}
		}

		var out io.Writer = ctx.App.Writer
		if ctx.IsSet("out") {
			outPath := ctx.String("out")
			f, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer f.Close()
			out = f
		}

		format := ctx.String("format")
		switch format {
		case "markdown":
			return writeSummaryAsMarkdown(out, summaries)

		case "csv":
			return writeSummaryAsCSV(out, summaries)

		case "json":
			data, err := json.MarshalIndent(summaries, "", "    ")
			if err != nil {
				return err
			}
			_, err = out.Write(data)
			return err

		default:
			return fmt.Errorf("unrecognized format %s", format)
		}
	},
}
