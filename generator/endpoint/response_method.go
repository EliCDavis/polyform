package endpoint

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type ResponseMethod[Response any] struct {
	ResponseWriter ResponseWriter[Response]
	Handler        func(r *http.Request) (Response, error)
}

func (jse ResponseMethod[Response]) ContentType() ContentType {
	return jse.ResponseWriter.ContentType()
}

func (jse ResponseMethod[Response]) runHandler(r *http.Request) (resp Response, err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Println("stacktrace from panic: \n" + string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	resp, err = jse.Handler(r)
	return
}

func (jse ResponseMethod[Response]) Handle(w http.ResponseWriter, r *http.Request) {
	response, err := jse.runHandler(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
		return
	}

	err = jse.ResponseWriter.Serialize(w, response)
	if err != nil {
		panic(err)
	}
}
