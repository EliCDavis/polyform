package generator

type cliCommand struct {
	Name        string
	Description string
	Aliases     []string
	Run         func(args []string) error
}
