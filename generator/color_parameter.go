package generator

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"strconv"

	"github.com/EliCDavis/polyform/drawing/coloring"
)

type ColorCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *string
}

type ColorParameter struct {
	Name           string
	DefaultValue   coloring.WebColor
	appliedProfile *coloring.WebColor
	CLI            *ColorCliParameterConfig
}

func (fp *ColorParameter) Reset() {
	fp.appliedProfile = nil
}

func (fp *ColorParameter) ApplyJsonMessage(msg json.RawMessage) error {
	colorStr := ""
	err := json.Unmarshal(msg, &colorStr)
	if err != nil {
		return err
	}
	color := hexToRGBA(colorStr)
	fp.appliedProfile = &color
	return nil
}

func rgbToHex(v coloring.WebColor) string {
	return fmt.Sprintf(
		"#%02x%02x%02x",
		v.R,
		v.G,
		v.B,
	)
}

func hexToRGBA(hex string) coloring.WebColor {
	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)
	return coloring.WebColor{
		R: byte(r),
		G: byte(g),
		B: byte(b),
		A: 255,
	}
}

func (cp ColorParameter) Schema() ParameterSchema {
	return ColorParameterSchema{
		ParameterSchemaBase: ParameterSchemaBase{
			Name: cp.Name,
			Type: "Color",
		},
		DefaultValue: cp.DefaultValue,
		CurrentValue: cp.WebValue(),
	}
}

func (fp ColorParameter) DisplayName() string {
	return fp.Name
}

func (fp ColorParameter) WebValue() coloring.WebColor {
	if fp.appliedProfile != nil {
		return (*fp.appliedProfile)
	}

	if fp.CLI != nil && fp.CLI.value != nil {
		return hexToRGBA(*fp.CLI.value)
	}
	return fp.DefaultValue
}

func (fp ColorParameter) Value() color.RGBA {
	return fp.WebValue().RGBA()
}

func (fp ColorParameter) initializeForCLI(set *flag.FlagSet) {
	if fp.CLI == nil {
		return
	}

	fp.CLI.value = set.String(
		fp.CLI.FlagName,
		rgbToHex(fp.DefaultValue),
		fp.CLI.Usage,
	)
}
