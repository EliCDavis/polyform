package run

import "github.com/EliCDavis/polyform/generator/manifest"

type CreateManifestResponse struct {
	Manifest manifest.Manifest `json:"manifest"`
	Id       string            `json:"id"`
}

type AvailableManifest struct {
	Name string `json:"name"`
	Port string `json:"port"`
}
