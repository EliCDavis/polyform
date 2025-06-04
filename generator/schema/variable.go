package schema

import "github.com/EliCDavis/polyform/generator/variable"

type VariableGroup struct {
	Variables map[string]variable.JsonContainer `json:"variables"`
	SubGroups map[string]VariableGroup          `json:"subgroups"`
}
