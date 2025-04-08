package markdown

import "github.com/EliCDavis/polyform/formats/txt"

var h1 = []byte("# ")
var h2 = []byte("## ")
var h3 = []byte("### ")
var bullet = []byte("* ")

type markdownWriter struct {
	writer      *txt.Writer
	indentLevel int
}

func (w *markdownWriter) header(header []byte, text string) (int, error) {
	w.writer.StartEntry()
	w.writer.Append(header)
	w.writer.String(text)
	w.writer.NewLine()
	w.writer.NewLine()
	return w.writer.FinishEntry()
}

func (w *markdownWriter) headerWithId(header []byte, text, id string) (int, error) {
	w.writer.StartEntry()
	w.writer.Append(header)
	w.writer.Append([]byte("<a id=\""))
	w.writer.String(id)
	w.writer.Append([]byte("\">"))
	w.writer.String(text)
	w.writer.Append([]byte("</a>"))
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

func (w *markdownWriter) Header2WithId(text string, id string) (int, error) {
	return w.headerWithId(h2, text, id)
}

func (w *markdownWriter) Header3(text string) (int, error) {
	return w.header(h3, text)
}

func (w *markdownWriter) Header3WithId(text string, id string) (int, error) {
	return w.headerWithId(h3, text, id)
}

func (w *markdownWriter) Link(text string, link string) (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("["))
	w.writer.String(text)
	w.writer.Append([]byte("](#"))
	w.writer.String(link)
	w.writer.Append([]byte(")"))
	return w.writer.FinishEntry()
}

func (w *markdownWriter) StartBulletList() (int, error) {
	w.indentLevel++
	return 0, nil
}

func (w *markdownWriter) EndBulletList() (int, error) {
	w.indentLevel--
	w.writer.StartEntry()
	if w.indentLevel == 0 {
		w.writer.NewLine()
	}
	return w.writer.FinishEntry()
}

func (w *markdownWriter) StartBullet() (int, error) {
	w.writer.StartEntry()
	for range w.indentLevel - 1 {
		w.writer.Append([]byte("    "))
	}
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

func (w *markdownWriter) StartItalics() (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("*"))
	return w.writer.FinishEntry()
}

func (w *markdownWriter) EndItalics() (int, error) {
	w.writer.StartEntry()
	w.writer.Append([]byte("*"))
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
