package endpoint

import (
	"net/http"
)

type BodyMethod[Body any] struct {
	Request RequestReader[Body]
	Handler func(request Request[Body]) error
}

func (jse BodyMethod[Body]) ContentType(r *http.Request) ContentType {
	return ""
}

func (jse BodyMethod[Body]) Handle(w http.ResponseWriter, r *http.Request) {

	request, err := jse.Request.Interpret(r)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	err = safeRun(func() error {
		return jse.Handler(Request[Body]{
			Body: request,
			Url:  r.URL.Path,
		})
	})

	if err != nil {
		writeJSONError(w, err)
		return
	}
}
