//go:build wasm

package generator

import (
	"log"

	wasmhttp "github.com/nlepage/go-wasm-http-server/v2"
)

var globalApp *App

// func wasmZip(this js.Value, cb []js.Value) interface{} { //

// 	if globalApp == nil {
// 		panic("global app not configured. Run app.Run(os.Args)")
// 	}

// 	b := bytes.Buffer{}
// 	err := globalApp.WriteZip(&b)
// 	if err != nil {
// 		log.Printf("error zipping: %s", err.Error())
// 	}

// 	log.Printf("completed")

// 	data := b.Bytes()

// 	dst := js.Global().Get("Uint8Array").New(len(data))
// 	js.CopyBytesToJS(dst, data)
// 	cb[0].Invoke(dst)
// 	return dst
// }

func os_setup(a *App) {
	// js.Global().Set("zipGeometry", js.FuncOf(wasmZip))
	globalApp = a
}

func (as *AppServer) Serve() error {
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

func isWasm() bool {
	return true
}
