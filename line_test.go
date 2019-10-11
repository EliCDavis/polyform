package mesh

import (
	"testing"

	"github.com/EliCDavis/vector"
)

func TestIntersect(t *testing.T) {
	l1 := NewLine(vector.NewVector2(0, 0), vector.NewVector2(1, 1))
	l2 := NewLine(vector.NewVector2(0, 1), vector.NewVector2(1, 0))
	if doIntersect(l1, l2) == false {
		t.Error("Lines should have interesected")
	}
}
