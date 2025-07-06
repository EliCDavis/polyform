package graph_test

import (
	"bytes"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/refutil"
	"github.com/stretchr/testify/assert"
)

func TestMermaid(t *testing.T) {
	// ARRANGE ================================================================
	tf := refutil.TypeFactory{}

	instance := graph.New(graph.Config{
		Name:        "Mermaid Test",
		Version:     "1.2.3",
		Description: "Yee haw",
		TypeFactory: &tf,
	})

	out := &bytes.Buffer{}

	// ACT ====================================================================
	err := graph.WriteMermaid(instance, out)

	// ASSERT =================================================================
	assert.NoError(t, err)
	assert.Equal(t, `---
title: Mermaid Test
---

flowchart LR
`, out.String())
}
