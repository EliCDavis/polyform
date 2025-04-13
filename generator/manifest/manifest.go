package manifest

type Entry struct {
	Metdata  map[string]any
	Artifact Artifact
}

type Manifest interface {
	Main() string
	Artifacts() map[string]Entry
}
