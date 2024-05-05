package generator

import "github.com/EliCDavis/jbtf"

type CustomGraphSerialization interface {
	ToJSON(encoder *jbtf.Encoder) ([]byte, error)
	FromJSON(decoder jbtf.Decoder, body []byte) error
}
