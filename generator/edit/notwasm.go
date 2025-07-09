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
	addr := fmt.Sprintf("%s:%s", as.Host, as.Port)
	url := fmt.Sprintf("%s://%s", protocol, addr)
	fmt.Printf("Serving over: %s\n", url)
	if as.LaunchWebbrowser {
		openURL(url)
	}

	if as.Tls {
		return http.ListenAndServeTLS(addr, as.CertPath, as.KeyPath, mux)
	}

	return http.ListenAndServe(addr, mux)
}
