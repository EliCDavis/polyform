package generator

type command struct {
	Name        string
	Description string
	Aliases     []string
	Run         func() error
}
