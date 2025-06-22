package properties

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"sort"

	"github.com/EliCDavis/polyform/examples/ply-utils/flags"
	"github.com/EliCDavis/polyform/formats/ply"
	"github.com/urfave/cli/v2"
)

type Analyzer interface {
	Analyze(buf []byte, endian binary.ByteOrder)
	Print(out io.Writer, printHistogram bool)
}

type PropertyAnalyzer[T Number] struct {
	Name   string
	Offset int
	End    int
	Min    T
	Max    T
	Counts map[T]int
}

type Number interface {
	int64 | float64 | int8 | byte | int16 | uint16 | int32 | uint32 | float32
}

type histogramEntry[T Number] struct {
	key   T
	count int
}

type SortByHistogramKey[T Number] []histogramEntry[T]

func (a SortByHistogramKey[T]) Len() int           { return len(a) }
func (a SortByHistogramKey[T]) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a SortByHistogramKey[T]) Less(i, j int) bool { return a[i].key < a[j].key }

func (pa *PropertyAnalyzer[T]) Print(out io.Writer, printHistogram bool) {
	fmt.Printf("[%s] min: %v; max: %v\n", pa.Name, pa.Min, pa.Max)

	if !printHistogram {
		return
	}

	entries := make([]histogramEntry[T], 0, len(pa.Counts))
	for key, count := range pa.Counts {
		entries = append(entries, histogramEntry[T]{
			key:   key,
			count: count,
		})
	}

	sort.Sort(SortByHistogramKey[T](entries))

	for _, entry := range entries {
		fmt.Fprintf(out, "%v: %d\n", entry.key, entry.count)
	}
}

func (pa *PropertyAnalyzer[T]) Analyze(buf []byte, endian binary.ByteOrder) {
	switch cpa := any(pa).(type) {
	case *PropertyAnalyzer[int8]:
		v := int8(buf[pa.Offset])
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[byte]:
		v := buf[pa.Offset]
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[int16]:
		v := int16(endian.Uint16(buf[pa.Offset:pa.End]))
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[uint16]:
		v := endian.Uint16(buf[pa.Offset:pa.End])
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[int32]:
		v := int32(endian.Uint32(buf[pa.Offset:pa.End]))
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[uint32]:
		v := endian.Uint32(buf[pa.Offset:pa.End])
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[float32]:
		v := math.Float32frombits(endian.Uint32(buf[pa.Offset:pa.End]))
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	case *PropertyAnalyzer[float64]:
		v := math.Float64frombits(endian.Uint64(buf[pa.Offset:pa.End]))
		cpa.Min = min(cpa.Min, v)
		cpa.Max = max(cpa.Max, v)
		cpa.Counts[v] = cpa.Counts[v] + 1

	default:
		panic(fmt.Errorf("unsupported type: %+v!!!", pa))
	}
}

func buildAnalyzer(prop ply.ScalarProperty, offset int) Analyzer {
	switch prop.Type {
	case ply.Char:
		return &PropertyAnalyzer[int8]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 1,
			Min:    127,
			Max:    -128,
			Counts: make(map[int8]int),
		}

	case ply.UChar:
		return &PropertyAnalyzer[byte]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 1,
			Min:    255,
			Max:    0,
			Counts: make(map[byte]int),
		}

	case ply.Short:
		return &PropertyAnalyzer[int16]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 2,
			Min:    32767,
			Max:    -32768,
			Counts: make(map[int16]int),
		}

	case ply.UShort:
		return &PropertyAnalyzer[uint16]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 2,
			Min:    65535,
			Max:    0,
			Counts: make(map[uint16]int),
		}

	case ply.Int:
		return &PropertyAnalyzer[int32]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 4,
			Min:    2147483647,
			Max:    -2147483648,
			Counts: make(map[int32]int),
		}

	case ply.UInt:
		return &PropertyAnalyzer[uint32]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 4,
			Min:    4294967295,
			Max:    0,
			Counts: make(map[uint32]int),
		}

	case ply.Float:
		return &PropertyAnalyzer[float32]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 4,
			Min:    math.MaxFloat32,
			Max:    -math.MaxFloat32,
			Counts: make(map[float32]int),
		}

	case ply.Double:
		return &PropertyAnalyzer[float64]{
			Name:   prop.Name(),
			Offset: offset,
			End:    offset + 8,
			Min:    math.MaxFloat64,
			Max:    -math.MaxFloat64,
			Counts: make(map[float64]int),
		}

	}
	panic(fmt.Errorf("unsupported prop type %s (found on %s)", prop.Type, prop.PropertyName))
}

var analyzePropertiesCommand = &cli.Command{
	Name:      "analyze",
	Usage:     "Create a summary for properties within a PLY file",
	Args:      true,
	ArgsUsage: "[names of the properties to analyze]",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:    "out",
			Aliases: []string{"o"},
		},
		&cli.BoolFlag{
			Name: "histogram",
		},
	},
	Action: func(ctx *cli.Context) error {
		f, err := flags.OpenPlyFile()
		if err != nil {
			return err
		}
		defer f.Close()

		header, err := ply.ReadHeader(f)
		if err != nil {
			return err
		}

		if len(header.Elements) != 1 {
			return fmt.Errorf("unsupported number of elements %d. Feel free to open up a PR", len(header.Elements))
		}

		var endian binary.ByteOrder

		switch header.Format {
		case ply.BinaryBigEndian:
			endian = binary.BigEndian

		case ply.BinaryLittleEndian:
			endian = binary.LittleEndian

		default:
			return fmt.Errorf("%s currenlty unsupported, feel free to open up a pr", header.Format)

		}

		specifiedProperties := ctx.Args().Slice()

		analyzers := make([]Analyzer, 0)
		pointSize := 0
		for _, p := range header.Elements[0].Properties {
			scalar, ok := p.(ply.ScalarProperty)
			if !ok {
				return fmt.Errorf("analyze currently does not support list properties. Feel free to open up a PR")
			}

			if len(specifiedProperties) == 0 || InSlice(scalar.PropertyName, specifiedProperties) {
				analyzers = append(analyzers, buildAnalyzer(scalar, pointSize))
			}

			pointSize += scalar.Size()
		}

		pointBuf := make([]byte, pointSize)
		reader := bufio.NewReader(f)

		for i := int64(0); i < header.Elements[0].Count; i++ {
			io.ReadFull(reader, pointBuf)
			for _, a := range analyzers {
				a.Analyze(pointBuf, endian)
			}
		}

		for _, a := range analyzers {
			a.Print(ctx.App.Writer, ctx.Bool("histogram"))
		}

		return nil
	},
}
