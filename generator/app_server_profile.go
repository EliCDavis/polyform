package generator

import (
	"fmt"
	"io"
	"net/http"
	"path"
)

func (as *AppServer) ProfileEndpoint(w http.ResponseWriter, r *http.Request) {
	as.producerLock.Lock()
	defer as.producerLock.Unlock()

	profileID := path.Base(r.URL.Path)

	var err error
	switch r.Method {
	case "GET", "":
		err = as.profileEndpoint_get(w, profileID)
	case "POST":
		err = as.profileEndpoint_post(w, r, profileID)
	}

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		writeJSONError(w, err)
	}
}

func (as *AppServer) profileEndpoint_get(w http.ResponseWriter, profileID string) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	n, ok := as.app.Schema().Nodes[profileID]
	if !ok {
		return fmt.Errorf("no node registered with ID: '%s'", profileID)
	}

	w.Write(n.parameter.ToMessage())
	return nil
}

func (as *AppServer) profileEndpoint_post(w http.ResponseWriter, r *http.Request, profileID string) (err error) {
	defer func() {
		if recErr := recover(); recErr != nil {
			err = fmt.Errorf("panic recover: %v", recErr)
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	_, err = as.ApplyMessage(profileID, body)
	if err != nil {
		return err
	}
	as.incModelVersion()
	w.Write([]byte("{}"))
	return nil
}
