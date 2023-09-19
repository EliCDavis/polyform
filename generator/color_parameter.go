package generator

import (
	"encoding/json"
	"flag"
	"fmt"
	"image/color"
	"strconv"
)

type ColorCliParameterConfig struct {
	FlagName string
	Usage    string
	value    *string
}

type ColorParameter struct {
	Name           string
	DefaultValue   color.RGBA
	appliedProfile *color.RGBA
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

func rgbToHex(v color.RGBA) string {
	return fmt.Sprintf(
		"#%02x%02x%02x",
		v.R,
		v.G,
		v.B,
	)
}

func hexToRGBA(hex string) color.RGBA {
	r, _ := strconv.ParseInt(hex[1:3], 16, 64)
	g, _ := strconv.ParseInt(hex[3:5], 16, 64)
	b, _ := strconv.ParseInt(hex[5:7], 16, 64)
	return color.RGBA{
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
		DefaultValue: rgbToHex(cp.DefaultValue),
		CurrentValue: rgbToHex(cp.Value()),
	}
}

func (fp ColorParameter) DisplayName() string {
	return fp.Name
}

func (fp ColorParameter) Value() color.RGBA {
	if fp.appliedProfile != nil {
		return *fp.appliedProfile
	}

	if fp.CLI != nil && fp.CLI.value != nil {
		return hexToRGBA(*fp.CLI.value)
	}
	return fp.DefaultValue
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
