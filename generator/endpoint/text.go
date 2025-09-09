package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
)

// ============================================================================

type TextRequestReader struct{}

func (jrbi TextRequestReader) Interpret(r *http.Request) (string, error) {
	body, err := io.ReadAll(r.Body)
	return string(body), err
}

// ============================================================================

type TextResponseWriter struct{}

func (jrw TextResponseWriter) Serialize(w http.ResponseWriter, response string) (err error) {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (jrw TextResponseWriter) ContentType(r *http.Request) ContentType {
	return PlainTextContentType
}
