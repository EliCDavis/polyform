package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCastsMatchThePolyformApp(t *testing.T) {
	app, err := os.ReadFile("../polyform/casts.go")
	require.NoError(t, err)
	own, err := os.ReadFile("casts.go")
	require.NoError(t, err)
	assert.Equal(t, string(app), string(own))
}
