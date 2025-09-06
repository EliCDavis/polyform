package endpoint

import (
	"encoding/json"
	"io"
	"net/http"
)

type errorResponse struct {
	Error string `json:"error"`
}

func readJSON[T any](body io.Reader) (T, error) {
	var v T
	data, err := io.ReadAll(body)
	if err != nil {
		return v, err
	}
	return v, json.Unmarshal(data, &v)
}

func writeJSONError(w http.ResponseWriter, err error) error {
	w.Header().Set("Content-Type", string(JsonContentType))
	w.WriteHeader(http.StatusInternalServerError)

	data, err := json.Marshal(errorResponse{
		Error: err.Error(),
	})
	if err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// ============================================================================

type JsonRequestReader[T any] struct{}

func (jrbi JsonRequestReader[T]) Interpret(r *http.Request) (T, error) {
	return readJSON[T](r.Body)
}

// ============================================================================

type JsonResponseWriter[Response any] struct{}

func (jrw JsonResponseWriter[Response]) Serialize(w http.ResponseWriter, response Response) (err error) {
	data, err := json.Marshal(response)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

func (jrw JsonResponseWriter[Response]) ContentType(r *http.Request) ContentType {
	return JsonContentType
}

// ============================================================================

func JsonMethod[Body any, Response any](handler func(request Request[Body]) (Response, error)) BodyResponseMethod[Body, Response] {
	return BodyResponseMethod[Body, Response]{
		Request:        JsonRequestReader[Body]{},
		ResponseWriter: JsonResponseWriter[Response]{},
		Handler:        handler,
	}
}

func JsonBodyMethod[Body any](handler func(request Request[Body]) error) BodyMethod[Body] {
	return BodyMethod[Body]{
		Request: JsonRequestReader[Body]{},
		Handler: handler,
	}
}

func JsonResponseMethod[Response any](handler func(r *http.Request) (Response, error)) ResponseMethod[Response] {
	return ResponseMethod[Response]{
		Handler:        handler,
		ResponseWriter: JsonResponseWriter[Response]{},
	}
}
