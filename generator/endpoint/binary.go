package endpoint

import (
	"io"
	"net/http"
)

// ============================================================================

type BinaryRequestReader struct{}

func (jrbi BinaryRequestReader) Interpret(r *http.Request) ([]byte, error) {
	return io.ReadAll(r.Body)
}

// ============================================================================

type BinaryResponseWriter struct{}

func (jrw BinaryResponseWriter) Serialize(w http.ResponseWriter, response []byte) (err error) {
	_, err = w.Write(response)
	return err
}

func (jrw BinaryResponseWriter) ContentType() ContentType {
	return BinaryContentType
}
