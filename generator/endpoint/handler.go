package endpoint

import (
	"fmt"
	"net/http"
)

type Handler struct {
	Methods map[string]Method
}

func (se Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method, ok := se.Methods[r.Method]

	if !ok {
		panic(fmt.Errorf("endpoint has not implemented HTTP method: '%s'", r.Method))
	}

	w.Header().Set("Content-Type", string(method.ContentType()))
	method.Handle(w, r)
}
