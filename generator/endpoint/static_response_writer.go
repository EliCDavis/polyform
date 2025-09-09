package endpoint

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type StaticResponse struct {
	Response []byte
	Type     string
}

func (srw StaticResponse) Handle(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(srw.Response)
	if err != nil {
		panic(fmt.Errorf("unable to write response: %s", err.Error()))
	}
}

func (srw StaticResponse) ContentType(r *http.Request) ContentType {
	return ContentType(srw.Type)
}

func StaticJson(data any) (StaticResponse, error) {
	serialized, err := json.Marshal(data)
	if err != nil {
		return StaticResponse{}, err
	}
	return StaticResponse{
		Response: serialized,
		Type:     string(JsonContentType),
	}, nil
}
