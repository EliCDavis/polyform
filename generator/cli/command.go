package cli

type Command struct {
	Name        string
	Description string
	Aliases     []string
	Run         func(args []string) error
}
