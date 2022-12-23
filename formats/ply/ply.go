package ply

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
)

type Format int64

const (
	ASCII Format = iota
	Binary
)

func scanToNextNonEmptyLine(scanner *bufio.Scanner) string {
	for scanner.Scan() {
		text := scanner.Text()
		if strings.TrimSpace(text) != "" {
			return text
		}
	}
	panic("end of scanner")
}

func readPlyHeaderFormat(scanner *bufio.Scanner) (Format, error) {
	line := scanToNextNonEmptyLine(scanner)
	contents := strings.Fields(line)

	if len(contents) != 3 {
		return -1, fmt.Errorf("unrecognized format line")
	}

	if contents[0] != "format" {
		return -1, fmt.Errorf("expected format line, received %s", contents[0])
	}

	if contents[2] != "1.0" {
		return -1, fmt.Errorf("unrecognized version format: %s", contents[2])
	}

	switch contents[1] {
	case "ascii":
		return ASCII, nil

	default:
		return -1, fmt.Errorf("unrecognized format: %s", contents[1])
	}
}

func readPlyProperty(contents []string) (Property, error) {
	if contents[1] == "list" {
		if len(contents) != 5 {
			return nil, errors.New("ill-formatted list property")
		}
		return ListProperty{
			name:      contents[4],
			countType: ScalarPropertyType(contents[2]),
			listType:  ScalarPropertyType(contents[3]),
		}, nil
	}

	if len(contents) != 3 {
		return nil, errors.New("ill-formatted scalar property")
	}

	return ScalarProperty{
		name: contents[2],
		Type: ScalarPropertyType(contents[1]),
	}, nil
}

func readPlyHeader(in io.Reader) (reader, error) {
	scanner := bufio.NewScanner(in)
	scanner.Scan()
	magicNumber := scanner.Text()
	if magicNumber != "ply" {
		return nil, fmt.Errorf("unrecognized magic number: '%s' (expected 'ply')", magicNumber)
	}

	format, err := readPlyHeaderFormat(scanner)
	if err != nil {
		return nil, err
	}

	if format != ASCII {
		return nil, fmt.Errorf("unimplemented format type: %d", format)
	}

	elements := make([]Element, 0)

	for {
		scanner.Scan()
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		if line == "end_header" {
			break
		}

		contents := strings.Fields(line)
		if contents[0] == "comment" {
			continue
		}

		if contents[0] == "element" {
			if len(contents) != 3 {
				return nil, errors.New("illegal element line in ply header")
			}

			elementCount, err := strconv.Atoi(contents[2])
			if err != nil {
				return nil, fmt.Errorf("unable to parse element count: %w", err)
			}

			elements = append(elements, Element{
				name:       contents[1],
				count:      elementCount,
				properties: make([]Property, 0),
			})
		}

		if contents[0] == "property" {
			property, err := readPlyProperty(contents)
			if err != nil {
				return nil, fmt.Errorf("unable to parse property: %w", err)
			}
			elements[len(elements)-1].properties = append(elements[len(elements)-1].properties, property)
		}
	}

	return &AsciiReader{elements: elements, scanner: scanner}, nil
}

func ToMesh(in io.Reader) (*modeling.Mesh, error) {
	reader, err := readPlyHeader(in)
	if err != nil {
		return nil, err
	}

	return reader.ReadMesh()
}
