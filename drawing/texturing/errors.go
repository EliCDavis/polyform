package texturing

import (
	"errors"
	"fmt"

	"github.com/EliCDavis/vector/vector2"
)

var ErrMismatchDimensions = errors.New("mismatch texture resolutions")

type InvalidDimension vector2.Vector[int]

func (id InvalidDimension) Error() string {
	v := vector2.Vector[int](id)
	return fmt.Sprintf("invalid texture dimensions %dx%d", v.X(), v.Y())
}
