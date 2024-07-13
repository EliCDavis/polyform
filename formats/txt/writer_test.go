package txt_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/formats/txt"
	"github.com/stretchr/testify/assert"
)

func TestWriter(t *testing.T) {
	b := &bytes.Buffer{}
	writer := txt.NewWriter(b)

	writer.String("testing")
	writer.Space()
	writer.Int(1)
	writer.Space()
	writer.Int(2)
	writer.Space()
	writer.Int(3)
	writer.NewLine()
	writer.Tab()
	writer.String("...?")
	writer.Space()
	writer.Float64(3.14159)
	writer.Space()
	writer.Float64MaxFigs(3.14159, 2)

	assert.Equal(t, `testing 1 2 3
	...? 3.14159 3.14`, b.String())

	assert.NoError(t, writer.Error())
}

var result string

func BenchmarkFormat_Fprintf(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	for n := 0; n < b.N; n++ {
		fmt.Fprintf(buf, "%s %d %d %d\n\t%s %f", "testing", 1, 2, 3, "...?", 3.14)
	}
	r = buf.String()
	result = r
}

func BenchmarkFormat_TextWriter(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	writer := txt.NewWriter(buf)
	for n := 0; n < b.N; n++ {
		writer.String("testing")
		writer.Space()
		writer.Int(1)
		writer.Space()
		writer.Int(2)
		writer.Space()
		writer.Int(3)
		writer.NewLine()
		writer.Tab()
		writer.String("...?")
		writer.Space()
		writer.Float64(3.14)
	}
	r = buf.String()
	result = r
}
