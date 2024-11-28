package txt

import (
	"io"
	"math"
	"strconv"
)

type Writer struct {
	out io.Writer
	buf []byte
	err error
}

func NewWriter(out io.Writer) *Writer {
	return &Writer{
		out: out,
		buf: make([]byte, 0),
	}
}

func (w *Writer) StartEntry() {
	w.buf = w.buf[:0]
}

func (w *Writer) FinishEntry() (int, error) {
	return w.Write(w.buf)
}

func (w Writer) Error() error {
	return w.err
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
	w.buf = strconv.AppendInt(w.buf, int64(i), 10)
}

func (w *Writer) Float64MaxFigs(f float64, figs int) {
	d := math.Pow10(figs)
	w.buf = strconv.AppendFloat(w.buf, math.Round(f*d)/d, 'f', -1, 64)
}

func (w *Writer) Append(p []byte) {
	w.buf = append(w.buf, p...)
}

const space = ' '

func (w *Writer) Space() {
	w.buf = append(w.buf, space)
}

const tab = '\t'

func (w *Writer) Tab() {
	w.buf = append(w.buf, tab)
}

const newLine = '\n'

func (w *Writer) NewLine() {
	w.buf = append(w.buf, newLine)
}

func (w *Writer) Float64(f float64) {
	w.buf = strconv.AppendFloat(w.buf, f, 'f', -1, 64)
}

func (w *Writer) String(s string) {
	w.buf = append(w.buf, s...)
}
