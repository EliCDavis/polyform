package generator

import (
	"fmt"
	"io"
	"net/http"
)

func (as *AppServer) GraphEndpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	as.producerLock.Lock()
	defer as.producerLock.Unlock()

	var err error

	switch r.Method {
	case "GET", "":
		err = as.graphEndpoint_Get(w)

	case "POST":
		err = as.graphEndpoint_Post(w, r)
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}

func (as *AppServer) graphEndpoint_Get(w http.ResponseWriter) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	_, err = w.Write(as.app.Graph())
	return err
}

func (as *AppServer) graphEndpoint_Post(w http.ResponseWriter, r *http.Request) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()
	data, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	err = as.app.ApplyGraph(data)
	if err != nil {
		return
	}
	_, err = w.Write([]byte("{}"))
	return
}
