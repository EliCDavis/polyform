package endpoint

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type BodyMethod[Body any] struct {
	Request RequestReader[Body]
	Handler func(request Request[Body]) error
}

func (jse BodyMethod[Body]) ContentType() ContentType {
	return ""
}

func (jse BodyMethod[Body]) runHandler(request Request[Body]) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	err = jse.Handler(request)
	return
}

func (jse BodyMethod[Body]) Handle(w http.ResponseWriter, r *http.Request) {

	request, err := jse.Request.Interpret(r)
	if err != nil {
		writeJSONError(w, err)
		return
	}

	err = jse.runHandler(Request[Body]{
		Body: request,
		Url:  r.URL.Path,
	})
	if err != nil {
		writeJSONError(w, err)
		return
	}
}
