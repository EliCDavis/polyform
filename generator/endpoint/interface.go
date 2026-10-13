package endpoint

import (
	"net/http"
	"net/url"
)

type Request[Body any] struct {
	Body  Body
	Url   string
	Query url.Values
}

type Method interface {
	ContentType(r *http.Request) ContentType
	Handle(w http.ResponseWriter, r *http.Request)
}

type RequestReader[T any] interface {
	Interpret(r *http.Request) (T, error)
}

type ResponseWriter[Response any] interface {
	Serialize(w http.ResponseWriter, response Response) (err error)
	ContentType(r *http.Request) ContentType
}
