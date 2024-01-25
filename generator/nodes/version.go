package nodes

type Versioned interface {
	Version() int
}

type VersionData struct {
	version int
}

func (v VersionData) Version() int {
	return v.version
}

func (v *VersionData) Increment() int {
	v.version++
	return v.version
}
