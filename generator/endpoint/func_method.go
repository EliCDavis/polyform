package endpoint

import (
	"fmt"
	"net/http"
	"runtime/debug"
)

type Func func(r *http.Request) error

func (jse Func) ContentType() ContentType {
	return JsonContentType
}

func (jse Func) runHandler(r *http.Request) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			fmt.Printf("panic: %v\nstacktrace from panic:\n%s", recErr, string(debug.Stack()))
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	err = jse(r)
	return
}

func (jse Func) Handle(w http.ResponseWriter, r *http.Request) {
	err := jse.runHandler(r)
	if err != nil {
		writeJSONError(w, err)
		return
	}
}
