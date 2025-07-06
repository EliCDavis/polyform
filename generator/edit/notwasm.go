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

	protocol := "http"
	if as.Tls {
		protocol = "https"
	}
	url := fmt.Sprintf("%s://%s:%s", protocol, as.Host, as.Port)
	fmt.Printf("Serving over: %s\n", url)
	if as.LaunchWebbrowser {
		openURL(url)
	}

	if as.Tls {
		return http.ListenAndServeTLS(url, as.CertPath, as.KeyPath, mux)
	}

	return http.ListenAndServe(url, mux)
}
