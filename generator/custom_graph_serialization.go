package generator

import "github.com/EliCDavis/polyform/formats/pgtf"

type CustomGraphSerialization interface {
	ToJSON(encoder *pgtf.Encoder) ([]byte, error)
	FromJSON(decoder pgtf.Decoder, body []byte) error
}
