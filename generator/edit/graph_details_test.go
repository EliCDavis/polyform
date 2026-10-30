package edit_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/persistence"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const springDetails = `{
	"name": "Spring",
	"version": "v1.2.0",
	"description": "A coil",
	"authors": [{"name": "Eli", "contactInfo": [{"medium": "github.com", "value": "EliCDavis"}]}]
}`

func TestGraphDetailsCanBeEdited(t *testing.T) {
	handler, instance := historyServer(t)

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/graph/details", body: springDetails})

	assert.Equal(t, graph.Details{
		Name:        "Spring",
		Version:     "v1.2.0",
		Description: "A coil",
		Authors: []persistence.Author{{
			Name:        "Eli",
			ContactInfo: []persistence.AuthorContact{{Medium: "github.com", Value: "EliCDavis"}},
		}},
	}, instance.Details())

	var served map[string]any
	require.NoError(t, json.Unmarshal(requireOK(t, handler, httpStep{method: http.MethodGet, url: "/graph/details"}), &served))
	var sent map[string]any
	require.NoError(t, json.Unmarshal([]byte(springDetails), &sent))
	assert.Equal(t, sent, served)
}

func TestEditingGraphDetailsCanBeUndone(t *testing.T) {
	handler, instance := historyServer(t)
	before := instance.Details()

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/graph/details", body: springDetails})
	call(t, handler, http.MethodPost, "/graph/history/undo", "")

	assert.Equal(t, before, instance.Details())
}

func TestGraphDetailsDropAuthorsWithNoName(t *testing.T) {
	handler, instance := historyServer(t)

	requireOK(t, handler, httpStep{method: http.MethodPost, url: "/graph/details", body: `{
		"name": " Spring ",
		"authors": [{"name": "  "}, {"name": " Eli ", "contactInfo": [{"medium": " ", "value": " "}]}]
	}`})

	details := instance.Details()
	assert.Equal(t, "Spring", details.Name)
	assert.Equal(t, []persistence.Author{{Name: "Eli"}}, details.Authors)
}
