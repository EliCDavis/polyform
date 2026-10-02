package graph

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSafelyTurnsAPanicIntoAnError(t *testing.T) {
	tests := map[string]struct {
		run  func() error
		want string
	}{
		"a panic with an error": {run: func() error { panic(errors.New("boom")) }, want: "boom"},
		"a panic with a string": {run: func() error { panic("boom") }, want: "boom"},
		"a nil map write":       {run: func() error { var m map[string]int; m["x"] = 1; return nil }, want: "nil map"},
		"an out of range index": {run: func() error { s := []int{}; _ = s[3]; return nil }, want: "range"},
		"a nil pointer deref":   {run: func() error { var p *int; _ = *p; return nil }, want: "nil pointer"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			err := safely(tc.run)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.want)
		})
	}
}

func TestSafelyPassesThroughWhatDoesNotPanic(t *testing.T) {
	assert.NoError(t, safely(func() error { return nil }))

	sentinel := errors.New("ordinary failure")
	assert.ErrorIs(t, safely(func() error { return sentinel }), sentinel)
}
