package manifest

type Entry struct {
	Metadata map[string]any `json:"metadata"`
	Artifact Artifact
}

type Manifest struct {
	Main    string           `json:"main"`
	Entries map[string]Entry `json:"entries"`
}

func SingleEntryManifest(name string, entry Entry) Manifest {
	return Manifest{
		Main: name,
		Entries: map[string]Entry{
			name: entry,
		},
	}
}
