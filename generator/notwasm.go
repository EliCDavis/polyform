//go:build !wasm

package generator

func os_setup(a *App) {

}

func isWasm() bool {
	return false
}
