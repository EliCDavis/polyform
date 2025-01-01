package endpoint

import "net/http"

type Request[Body any] struct {
	Body Body
	Url  string
}

type Method interface {
	ContentType() ContentType
	Handle(w http.ResponseWriter, r *http.Request)
}

type RequestReader[T any] interface {
	Interpret(r *http.Request) (T, error)
}

type ResponseWriter[Response any] interface {
	Serialize(w http.ResponseWriter, response Response) (err error)
	ContentType() ContentType
}
