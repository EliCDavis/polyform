package endpoint

import "net/http"

type RequestReaderFunc[T any] func(r *http.Request) (T, error)

func (f RequestReaderFunc[T]) Interpret(r *http.Request) (T, error) {
	return f(r)
}
