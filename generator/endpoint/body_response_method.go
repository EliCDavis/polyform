package endpoint

import (
	"net/http"
)

type BodyResponseMethod[Body any, Response any] struct {
	Request        RequestReader[Body]
	ResponseWriter ResponseWriter[Response]
	Handler        func(request Request[Body]) (Response, error)
}

func (jse BodyResponseMethod[Body, Response]) ContentType(r *http.Request) ContentType {
	return jse.ResponseWriter.ContentType(r)
}

func (jse BodyResponseMethod[Body, Response]) Handle(w http.ResponseWriter, r *http.Request) {

	request, err := jse.Request.Interpret(r)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	response, err := safeReturn(func() (Response, error) {
		return jse.Handler(Request[Body]{
			Body: request,
			Url:  r.URL.Path,
		})
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
