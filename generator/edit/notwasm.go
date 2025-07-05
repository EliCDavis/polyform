//go:build !wasm

package edit

import (
	"fmt"
	"net/http"
)

func (as *Server) Serve() error {
	mux, err := as.Handler("/")
	if err != nil {
		return err
	}

	connection := fmt.Sprintf("%s:%s", as.Host, as.Port)
	if as.Tls {
		url := fmt.Sprintf("https://%s", connection)
		fmt.Printf("Serving over: %s\n", url)
		if as.LaunchWebbrowser {
			openURL(url)
		}
		return http.ListenAndServeTLS(connection, as.CertPath, as.KeyPath, mux)

	} else {
		url := fmt.Sprintf("http://%s", connection)
		fmt.Printf("Serving over: %s\n", url)
		if as.LaunchWebbrowser {
			openURL(url)
		}
		return http.ListenAndServe(connection, mux)
	}
}
