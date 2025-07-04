//go:build !wasm

package generator

import (
	"fmt"
	"net/http"
)

func os_setup(a *App) {

}

func isWasm() bool {
	return false
}

func (as *EditServer) Serve() error {
	mux, err := as.Handler("/")
	if err != nil {
		return err
	}

	connection := fmt.Sprintf("%s:%s", as.host, as.port)
	if as.tls {
		url := fmt.Sprintf("https://%s", connection)
		fmt.Printf("Serving over: %s\n", url)
		if as.launchWebbrowser {
			openURL(url)
		}
		return http.ListenAndServeTLS(connection, as.certPath, as.keyPath, mux)

	} else {
		url := fmt.Sprintf("http://%s", connection)
		fmt.Printf("Serving over: %s\n", url)
		if as.launchWebbrowser {
			openURL(url)
		}
		return http.ListenAndServe(connection, mux)
	}
}
