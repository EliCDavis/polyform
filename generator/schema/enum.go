package schema

import "github.com/EliCDavis/polyform/generator/enum"

type Enum struct {
	Name        string        `json:"name"`
	Description string        `json:"description,omitempty"`
	Members     []enum.Member `json:"members"`
}
