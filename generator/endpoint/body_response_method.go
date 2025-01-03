package endpoint

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type BodyResponseMethod[Body any, Response any] struct {
	Request        RequestReader[Body]
	ResponseWriter ResponseWriter[Response]
	Handler        func(request Request[Body]) (Response, error)
}

func (jse BodyResponseMethod[Body, Response]) ContentType() ContentType {
	return jse.ResponseWriter.ContentType()
}

func (jse BodyResponseMethod[Body, Response]) runHandler(request Request[Body]) (resp Response, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	resp, err = jse.Handler(request)
	return
}

func (jse BodyResponseMethod[Body, Response]) Handle(w http.ResponseWriter, r *http.Request) {

	request, err := jse.Request.Interpret(r)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	response, err := jse.runHandler(Request[Body]{
		Body: request,
		Url:  r.URL.Path,
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
