package ply

import (
	"encoding/binary"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/vector/vector3"
)

const shPropertyPrefix = "f_rest_"

func shPropertyName(i int) string {
	return fmt.Sprintf("%s%d", shPropertyPrefix, i)
}

func shPropertyIndex(name string) (int, bool) {
	rest, ok := strings.CutPrefix(name, shPropertyPrefix)
	if !ok {
		return 0, false
	}
	i, err := strconv.Atoi(rest)
	if err != nil || i < 0 {
		return 0, false
	}
	return i, true
}

// PLY stores f_rest channel-major: f_rest_[c*count + j] is channel c of coefficient j.
type shProperties struct {
	offsets []int
	types   []ScalarPropertyType
}

func (p shProperties) coefficients() int {
	return len(p.offsets) / 3
}

func (p shProperties) index(channel, coefficient int) int {
	return channel*p.coefficients() + coefficient
}

func (p shProperties) claims(prop Property) bool {
	i, ok := shPropertyIndex(prop.Name())
	return ok && i < len(p.offsets)
}

func (p shProperties) newArrays(count int64) [][]vector3.Float64 {
	arrays := make([][]vector3.Float64, p.coefficients())
	for j := range arrays {
		arrays[j] = make([]vector3.Float64, count)
	}
	return arrays
}

func findSHProperties(element Element, byteOffsets bool) (shProperties, bool) {
	offsets := map[int]int{}
	types := map[int]ScalarPropertyType{}
	position := 0
	for column, prop := range element.Properties {
		scalar := prop.(ScalarProperty)
		if i, ok := shPropertyIndex(scalar.PropertyName); ok {
			if byteOffsets {
				offsets[i] = position
			} else {
				offsets[i] = column
			}
			types[i] = scalar.Type
		}
		position += scalar.Size()
	}

	count := len(offsets)
	if count == 0 || count%3 != 0 {
		return shProperties{}, false
	}

	found := shProperties{
		offsets: make([]int, count),
		types:   make([]ScalarPropertyType, count),
	}
	for i := 0; i < count; i++ {
		offset, ok := offsets[i]
		if !ok {
			return shProperties{}, false
		}
		found.offsets[i] = offset
		found.types[i] = types[i]
	}
	return found, true
}

func readBinaryScalar(buf []byte, t ScalarPropertyType, endian binary.ByteOrder) float64 {
	switch t {
	case Char:
		return float64(int8(buf[0]))
	case UChar:
		return float64(buf[0])
	case Short:
		return float64(int16(endian.Uint16(buf)))
	case UShort:
		return float64(endian.Uint16(buf))
	case Int:
		return float64(int32(endian.Uint32(buf)))
	case UInt:
		return float64(endian.Uint32(buf))
	case Float:
		return float64(math.Float32frombits(endian.Uint32(buf)))
	case Double:
		return math.Float64frombits(endian.Uint64(buf))
	}
	panic(fmt.Errorf("unimplemented %s", t))
}

type SHPropertyReader struct{}

func (SHPropertyReader) buildBinary(element Element, endian binary.ByteOrder) binaryPropertyReader {
	props, ok := findSHProperties(element, true)
	if !ok {
		return nil
	}
	return &builtBinarySHPropertyReader{
		props:  props,
		endian: endian,
		arrays: props.newArrays(element.Count),
	}
}

func (SHPropertyReader) buildAscii(element Element) asciiPropertyReader {
	props, ok := findSHProperties(element, false)
	if !ok {
		return nil
	}
	return &builtAsciiSHPropertyReader{
		props:  props,
		arrays: props.newArrays(element.Count),
	}
}

type builtBinarySHPropertyReader struct {
	props  shProperties
	endian binary.ByteOrder
	arrays [][]vector3.Float64
}

func (r *builtBinarySHPropertyReader) value(buf []byte, channel, coefficient int) float64 {
	i := r.props.index(channel, coefficient)
	return readBinaryScalar(buf[r.props.offsets[i]:], r.props.types[i], r.endian)
}

func (r *builtBinarySHPropertyReader) Read(buf []byte, i int64) {
	for j := range r.arrays {
		r.arrays[j][i] = vector3.New(r.value(buf, 0, j), r.value(buf, 1, j), r.value(buf, 2, j))
	}
}

func (r *builtBinarySHPropertyReader) ClaimsProperty(prop Property) bool {
	return r.props.claims(prop)
}

func (r *builtBinarySHPropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	for j, data := range r.arrays {
		m = m.SetFloat3Attribute(modeling.SHAttribute(j), data)
	}
	return m
}

type builtAsciiSHPropertyReader struct {
	props  shProperties
	arrays [][]vector3.Float64
}

func (r *builtAsciiSHPropertyReader) value(buf []string, channel, coefficient int) (float64, error) {
	return strconv.ParseFloat(buf[r.props.offsets[r.props.index(channel, coefficient)]], 64)
}

func (r *builtAsciiSHPropertyReader) Read(buf []string, i int64) error {
	for j := range r.arrays {
		x, err := r.value(buf, 0, j)
		if err != nil {
			return err
		}
		y, err := r.value(buf, 1, j)
		if err != nil {
			return err
		}
		z, err := r.value(buf, 2, j)
		if err != nil {
			return err
		}
		r.arrays[j][i] = vector3.New(x, y, z)
	}
	return nil
}

func (r *builtAsciiSHPropertyReader) ClaimsProperty(prop Property) bool {
	return r.props.claims(prop)
}

func (r *builtAsciiSHPropertyReader) UpdateMesh(m modeling.Mesh) modeling.Mesh {
	for j, data := range r.arrays {
		m = m.SetFloat3Attribute(modeling.SHAttribute(j), data)
	}
	return m
}
