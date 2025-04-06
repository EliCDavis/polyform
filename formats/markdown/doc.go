package markdown

import (
	"io"

	"github.com/EliCDavis/polyform/formats/txt"
)

var h1 = []byte("# ")
var h2 = []byte("## ")
var h3 = []byte("### ")
var bullet = []byte("* ")

func NewWriter(out io.Writer) *Writer {
	return &Writer{
		writer: txt.NewWriter(out),
	}
}

type Writer struct {
	writer *txt.Writer
}

func (w *Writer) header(header []byte, text string) (int, error) {
	w.writer.StartEntry()
	w.writer.Write(header)
	w.writer.String(text)
	w.writer.NewLine()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *Writer) Header1(text string) (int, error) {
	return w.header(h1, text)
}

func (w *Writer) Header2(text string) (int, error) {
	return w.header(h2, text)
}

func (w *Writer) Header3(text string) (int, error) {
	return w.header(h3, text)
}

func (w *Writer) Bullet(text string) (int, error) {
	w.writer.StartEntry()
	w.writer.Write(bullet)
	w.writer.String(text)
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *Writer) NewLine() (int, error) {
	w.writer.StartEntry()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *Writer) Paragraph(text string) (int, error) {
	w.writer.StartEntry()
	w.writer.String(text)
	w.writer.NewLine()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *Writer) Error() error {
	return w.writer.Error()
}
