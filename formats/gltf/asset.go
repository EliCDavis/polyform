package gltf

// Metadata about the glTF asset.
// https://github.com/KhronosGroup/glTF/blob/main/specification/2.0/schema/asset.schema.json
type Asset struct {
	Property
	Version    string `json:"version"`              // The glTF version in the form of `<major>.<minor>` that this asset targets.
	Generator  string `json:"generator,omitempty"`  // Tool that generated this glTF model.  Useful for debugging.
	Copyright  string `json:"copyright,omitempty"`  // A copyright message suitable for display to credit the content creator.
	MinVersion string `json:"minVersion,omitempty"` // The minimum glTF version in the form of `<major>.<minor>` that this asset targets. This property **MUST NOT** be greater than the asset version.
}
