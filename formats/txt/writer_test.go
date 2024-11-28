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

	writer.StartEntry()
	writer.String("testing")
	writer.Space()
	writer.Int(1)
	writer.Space()
	writer.Int(2)
	writer.Space()
	writer.Int(3)
	writer.NewLine()
	writer.Tab()
	writer.Append([]byte("...?"))
	writer.Space()
	writer.Float64(3.14159)
	writer.Space()
	writer.Float64MaxFigs(3.14159, 2)
	writer.FinishEntry()

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

func BenchmarkFormat_Fprint(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	for n := 0; n < b.N; n++ {
		fmt.Fprint(buf, "testing ", 1, 2, 3, "\n\t...? ", 3.14)
	}
	r = buf.String()
	result = r
}

func BenchmarkFormat_TextWriter(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	writer := txt.NewWriter(buf)
	for n := 0; n < b.N; n++ {
		writer.StartEntry()
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
		writer.FinishEntry()
	}
	r = buf.String()
	result = r
}

func BenchmarkFormat_ObjVert_Fprint(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	for n := 0; n < b.N; n++ {
		fmt.Fprint(buf, "v ", 1.234567, 1.234567, 1.234567, "\n")
	}
	r = buf.String()
	result = r
}

func BenchmarkFormat_ObjVert_Fprintf(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	for n := 0; n < b.N; n++ {
		fmt.Fprintf(buf, "v %f %f %f\n", 1.234567, 1.234567, 1.234567)
	}
	r = buf.String()
	result = r
}

func BenchmarkFormat_ObjVert_TextWriter(b *testing.B) {
	var r string
	buf := &bytes.Buffer{}
	writer := txt.NewWriter(buf)
	for n := 0; n < b.N; n++ {
		writer.StartEntry()
		writer.String("v ")
		writer.Float64(1.234567)
		writer.Space()
		writer.Float64(1.234567)
		writer.Space()
		writer.Float64(1.234567)
		writer.NewLine()
		writer.FinishEntry()
	}
	r = buf.String()
	result = r
}
