package markdown

import "github.com/EliCDavis/polyform/formats/txt"

var h1 = []byte("# ")
var h2 = []byte("## ")
var h3 = []byte("### ")
var bullet = []byte("* ")

type markdownWriter struct {
	writer *txt.Writer
}

func (w *markdownWriter) header(header []byte, text string) (int, error) {
	w.writer.StartEntry()
	w.writer.Append(header)
	w.writer.String(text)
	w.writer.NewLine()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) Header1(text string) (int, error) {
	return w.header(h1, text)
}

func (w *markdownWriter) Header2(text string) (int, error) {
	return w.header(h2, text)
}

func (w *markdownWriter) Header3(text string) (int, error) {
	return w.header(h3, text)
}

func (w *markdownWriter) StartBulletList() (int, error) {
	return 0, nil
}

func (w *markdownWriter) EndBulletList() (int, error) {
	w.writer.StartEntry()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) StartBullet() (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("* "))
	return w.writer.FinishEntry()
}

func (w *markdownWriter) EndBullet() (int, error) {
	w.writer.StartEntry()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) StartBold() (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("**"))
	return w.writer.FinishEntry()
}

func (w *markdownWriter) EndBold() (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("**"))
	return w.writer.FinishEntry()
}

func (w *markdownWriter) Text(text string) (int, error) {
	w.writer.StartEntry()
	w.writer.String(text)
	return w.writer.FinishEntry()
}

func (w *markdownWriter) NewLine() (int, error) {
	w.writer.StartEntry()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) Paragraph(text string) (int, error) {
	w.writer.StartEntry()
	w.writer.String(text)
	w.writer.NewLine()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) Error() error {
	return w.writer.Error()
}
