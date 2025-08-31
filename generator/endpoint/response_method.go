package endpoint

import (
	"net/http"
)

type ResponseMethod[Response any] struct {
	ResponseWriter ResponseWriter[Response]
	Handler        func(r *http.Request) (Response, error)
}

func (jse ResponseMethod[Response]) ContentType(r *http.Request) ContentType {
	return jse.ResponseWriter.ContentType(r)
}

func (jse ResponseMethod[Response]) Handle(w http.ResponseWriter, r *http.Request) {
	response, err := safeReturn(func() (Response, error) {
		return jse.Handler(r)
	})

	if err != nil {
		writeJSONError(w, err)
		return
	}

	err = jse.ResponseWriter.Serialize(w, response)
	if err != nil {
		panic(err)
	}
}

type ResponseMethodContentTypeOverride[Response any] struct {
	ResponseWriter ResponseWriter[Response]
	Content        func(r *http.Request) ContentType
	Handler        func(r *http.Request) (Response, error)
}

func (jse ResponseMethodContentTypeOverride[Response]) ContentType(r *http.Request) ContentType {
	return jse.Content(r)
}

func (jse ResponseMethodContentTypeOverride[Response]) Handle(w http.ResponseWriter, r *http.Request) {
	response, err := safeReturn(func() (Response, error) {
		return jse.Handler(r)
	})

	if err != nil {
		writeJSONError(w, err)
		return
	}

	err = jse.ResponseWriter.Serialize(w, response)
	if err != nil {
		panic(err)
	}
}
