package modeling

import "fmt"

const (
	PositionAttribute  = "Position"
	NormalAttribute    = "Normal"
	ColorAttribute     = "Color"
	TexCoordAttribute  = "TexCoord"
	ClassAttribute     = "Class"
	IntensityAttribute = "Intensity"
	JointAttribute     = "Joint"
	WeightAttribute    = "Weight"
	ScaleAttribute     = "Scale"
	RotationAttribute  = "Rotation"
	OpacityAttribute   = "Opacity"
	FDCAttribute       = "FDC"
)

// Coefficient 0 is the first degree-1 term; the DC term lives in FDCAttribute.
func SHAttribute(coefficient int) string {
	return fmt.Sprintf("SH_%d", coefficient)
}

func SHCoefficientCount(m Mesh) int {
	count := 0
	for m.HasFloat3Attribute(SHAttribute(count)) {
		count++
	}
	return count
}
