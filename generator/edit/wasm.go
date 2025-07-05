//go:build wasm

package edit

import (
	"log"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

func (as *EditServer) Serve() error {
	mux, err := as.Handler("/app.html")
	if err != nil {
		return err
	}

	log.Print("Starting wasm serve...")

	if _, err = wasmhttp.Serve(mux); err != nil {
		return err
	}
	// f()
	select {}

	// return err
}
