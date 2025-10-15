package texturing

import (
	"fmt"

	"github.com/EliCDavis/vector/vector2"
)

type InvalidDimension vector2.Vector[int]

func (id InvalidDimension) Error() string {
	v := vector2.Vector[int](id)
	return fmt.Sprintf("invalid texture dimensions %dx%d", v.X(), v.Y())
}
