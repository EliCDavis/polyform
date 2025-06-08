package variable

type info struct {
	name        string
	description string
}

func (i info) Name() string {
	return i.name
}

func (i info) Description() string {
	return i.description
}

func (i *info) SetDescription(description string) {
	i.description = description
}

type Info interface {
	Name() string
	Description() string
	SetDescription(description string)
}
