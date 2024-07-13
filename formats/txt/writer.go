package txt

import (
	"io"
	"math"
	"strconv"
)

var space = []byte{' '}
var tab = []byte{'\t'}
var newLine = []byte{'\n'}

type Writer struct {
	out io.Writer
	err error
}

func NewWriter(out io.Writer) *Writer {
	return &Writer{
		out: out,
	}
}

func (w *Writer) String(s string) {
	w.Write([]byte(s))
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.err != nil {
		return 0, w.err
	}
	var i int
	i, w.err = w.out.Write(p)
	return i, w.err
}

func (w *Writer) Int(i int) {
	w.Write([]byte(strconv.FormatInt(int64(i), 10)))
}

func (w *Writer) Float64(f float64) {
	w.Write([]byte(strconv.FormatFloat(f, 'f', -1, 64)))
}

func (w *Writer) Float64MaxFigs(f float64, figs int) {
	d := math.Pow10(figs)
	w.Write([]byte(strconv.FormatFloat(math.Round(f*d)/d, 'f', -1, 64)))
}

func (w *Writer) Space() {
	w.Write(space)
}

func (w *Writer) Tab() {
	w.Write(tab)
}

func (w *Writer) NewLine() {
	w.Write(newLine)
}

func (w Writer) Error() error {
	return w.err
}
