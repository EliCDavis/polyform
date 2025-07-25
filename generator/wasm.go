//go:build wasm

package generator

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

func isWasm() bool {
	return true
}
