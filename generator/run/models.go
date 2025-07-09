package run

import "github.com/EliCDavis/polyform/generator/manifest"

type CreateManifestResponse struct {
	Manifest manifest.Manifest
	Id       string
}
