package endpoint

import (
	"net/http"
)

type Func func(r *http.Request) error

func (jse Func) ContentType() ContentType {
	return JsonContentType
}

func (jse Func) Handle(w http.ResponseWriter, r *http.Request) {
	err := safeRun(func() error {
		return jse(r)
	})
	if err != nil {
		writeJSONError(w, err)
		return
	}
}
