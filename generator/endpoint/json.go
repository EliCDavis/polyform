package endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime/debug"
)

func readJSON[T any](body io.Reader) (T, error) {
	var v T
	data, err := io.ReadAll(body)
	if err != nil {
		return v, err
	}
	return v, json.Unmarshal(data, &v)
}

func writeJSONError(out io.Writer, err error) error {
	var d struct {
		Error string `json:"error"`
	} = struct {
		Error string `json:"error"`
	}{
		Error: err.Error(),
	}
	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	_, err = out.Write(data)
	return err
}

const (
	JsonContentType = "application/json"
)

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

func (jrw JsonResponseWriter[Response]) ContentType() string {
	return JsonContentType
}

// ============================================================================

type JsonMethod[Body any, Response any] struct {
	Handler func(request Request[Body]) (Response, error)
}

func (jse JsonMethod[Body, Response]) ContentType() string {
	return JsonContentType
}

func (jse JsonMethod[Body, Response]) runHandler(request Request[Body]) (resp Response, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	resp, err = jse.Handler(request)
	return
}

func (jse JsonMethod[Body, Response]) Handle(w http.ResponseWriter, r *http.Request) {

	request, err := readJSON[Body](r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
		return
	}

	response, err := jse.runHandler(Request[Body]{
		Body: request,
		Url:  r.URL.Path,
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
		return
	}

	data, err := json.Marshal(response)
	if err != nil {
		panic(err)
	}

	_, err = w.Write(data)
	if err != nil {
		panic(err)
	}
}
