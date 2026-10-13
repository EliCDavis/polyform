package graph_test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Leaf names repeat across groups so lifted inputs collide and get suffixed.
var fuzzVariablePaths = []string{"Size", "Wood/Size", "Metal/Size", "Metal/Depth"}

var fuzzSeedPlans = [][]byte{
	{},
	{0},
	{3, 4, 2, 1, 2, 0, 1, 1, 1, 1, 1, 1, 1, 1, 1},
	{1, 2, 1, 1, 0, 1, 2, 1, 1, 1, 0, 1},
	bytes.Repeat([]byte{255}, 64),
}

type planReader struct {
	plan []byte
	at   int
}

func (r *planReader) next(n int) int {
	if r.at >= len(r.plan) {
		return 0
	}
	r.at++
	return int(r.plan[r.at-1]) % n
}

// buildProject turns plan into a project whose sub-graphs each sum their
// variable reads and the Out of the sub-graphs they place. A sub-graph only
// places earlier ones, so placements never loop.
func buildProject(t testing.TB, name string, plan []byte, valueOffset float64) (*graph.Instance, []string) {
	t.Helper()
	r := &planReader{plan: plan}
	inst := importTestInstance(t, name)

	variables := fuzzVariablePaths[:1+r.next(len(fuzzVariablePaths))]
	for i, path := range variables {
		floatVariable(t, inst, path, float64(i+1)+valueOffset)
	}

	var defs []string
	for i := range 1 + r.next(5) {
		var reads []string
		for _, path := range variables {
			for range r.next(3) {
				reads = append(reads, path)
			}
		}
		id := fmt.Sprintf("d%d", i)
		def, sumID := readsVariableSummed(t, inst, id, reads...)
		for _, placed := range defs {
			for range r.next(2) {
				_, placementID := place(t, def, placed)
				require.NoError(t, def.ConnectNodes(placementID, "Out", sumID, "Values"))
			}
		}
		defs = append(defs, id)
	}
	return inst, defs
}

func valueOf(t testing.TB, inst *graph.Instance, id string) float64 {
	t.Helper()
	placement, _ := place(t, inst, id)
	return out(t, placement)
}

func FuzzImportedSubGraphsComputeWhatTheyDidAtTheirSource(f *testing.F) {
	for _, plan := range fuzzSeedPlans {
		f.Add(plan)
	}

	f.Fuzz(func(t *testing.T, plan []byte) {
		src, defs := buildProject(t, "Source", plan, 0)
		payload := encode(t, src)

		importInto := func() (*graph.Instance, graph.ImportSubGraphsResult) {
			dest, _ := buildProject(t, "Dest", plan, 100)
			result, err := dest.ImportSubGraphDefinitions(payload, "")
			require.NoError(t, err)
			return dest, result
		}
		dest, result := importInto()
		require.Len(t, result.Imported, len(defs))

		twin, _ := importInto()
		require.Equal(t, string(encode(t, dest)), string(encode(t, twin)), "importing is not deterministic")

		untouched, _ := buildProject(t, "Dest", plan, 100)
		for _, id := range defs {
			assert.Equal(t, valueOf(t, untouched, id), valueOf(t, dest, id), "the destination's own %s changed", id)
		}

		for _, imported := range result.Imported {
			original := imported.OriginalID
			require.NotEmpty(t, original, "every id was taken, so every import is renamed")
			assert.Equal(t, valueOf(t, src, original), valueOf(t, dest, imported.ID), "%s → %s", original, imported.ID)

			for _, input := range imported.Inputs {
				require.True(t, input.HasDefault)
				changed, _ := buildProject(t, "Source", plan, 0)
				_, err := changed.UpdateVariable(input.Variable, []byte("1000"))
				require.NoError(t, err)

				placement, placementID := place(t, dest, imported.ID)
				feed(t, dest, placementID, input.Port, 1000)
				assert.Equal(t, valueOf(t, changed, original), out(t, placement),
					"%s: feeding %s should act like setting %s", imported.ID, input.Port, input.Variable)
			}
		}
	})
}

func addPayloadSeeds(f *testing.F) {
	for _, plan := range fuzzSeedPlans {
		src, _ := buildProject(f, "Source", plan, 0)
		f.Add(encode(f, src))
	}
	f.Add(rawPayload(`{"a": {"name": "A", "nodes": ` + placing("a") + `}}`))
	f.Add(rawPayload(`{"a": {"name": "A", "nodes": {"v": {"type": "Ghost", "variable": "Ghost"}}}}`))
	f.Add(rawPayload(`{"a": {"name": "A", "nodes": ` + placing("missing") + `}}`))
	f.Add([]byte(`{"buffers": [], "data": {"subGraphs": {"": {"nodes": {"x": {"type": "nope"}}}}}}`))
}

func FuzzImportFailsCleanly(f *testing.F) {
	addPayloadSeeds(f)

	f.Fuzz(func(t *testing.T, payload []byte) {
		dest, _ := buildProject(t, "Dest", []byte{2, 2, 1, 1, 1}, 100)
		before := string(encode(t, dest))

		if _, err := dest.ImportSubGraphDefinitions(payload, ""); err != nil {
			require.Equal(t, before, string(encode(t, dest)), "a failed import changed the graph")
			return
		}

		saved := encode(t, dest)
		reloaded := importTestInstance(t, "Reloaded")
		require.NoError(t, reloaded.ApplyAppSchema(saved), "a successful import saved something that does not load")
		require.Equal(t, string(saved), string(encode(t, reloaded)))
	})
}

func FuzzLoadFailsCleanlyAndRoundTrips(f *testing.F) {
	addPayloadSeeds(f)

	f.Fuzz(func(t *testing.T, payload []byte) {
		inst, _ := buildProject(t, "Open", []byte{2, 2, 1, 1, 1}, 100)
		before := string(encode(t, inst))

		if err := inst.ApplyAppSchema(payload); err != nil {
			require.Equal(t, before, string(encode(t, inst)), "a failed load changed the graph")
			return
		}

		saved := encode(t, inst)
		reloaded := importTestInstance(t, "Reloaded")
		require.NoError(t, reloaded.ApplyAppSchema(saved), "a loaded graph saves something that does not load")
		require.Equal(t, string(saved), string(encode(t, reloaded)))
	})
}
