// Command polyform-mcp is a standalone MCP server that exposes polyform's
// node-graph construction API as tools, so any MCP-compatible client (an
// AI agent, or otherwise) can build a polyform graph turn-by-turn. It is
// independent of the polyform CLI: it neither depends on nor is invoked
// through cmd/polyform.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/mcp"

	// Blank-imported so each package's init() registers its node types with
	// the shared generator type factory, retrieved below via
	// generator.Types(). This mirrors the import list in cmd/polyform/main.go.
	_ "github.com/EliCDavis/polyform/drawing/coloring"
	_ "github.com/EliCDavis/polyform/drawing/texturing"
	_ "github.com/EliCDavis/polyform/drawing/texturing/normals"
	_ "github.com/EliCDavis/polyform/drawing/texturing/pattern"

	_ "github.com/EliCDavis/polyform/formats/colmap"
	_ "github.com/EliCDavis/polyform/formats/gltf"
	_ "github.com/EliCDavis/polyform/formats/obj"
	_ "github.com/EliCDavis/polyform/formats/opensfm"
	_ "github.com/EliCDavis/polyform/formats/ply"
	_ "github.com/EliCDavis/polyform/formats/splat"
	_ "github.com/EliCDavis/polyform/formats/spz"
	_ "github.com/EliCDavis/polyform/formats/stl"

	_ "github.com/EliCDavis/polyform/generator/manifest/basics"
	_ "github.com/EliCDavis/polyform/generator/parameter"
	_ "github.com/EliCDavis/polyform/generator/subgraph/register"

	_ "github.com/EliCDavis/polyform/math"
	_ "github.com/EliCDavis/polyform/math/constant"
	_ "github.com/EliCDavis/polyform/math/geometry"
	_ "github.com/EliCDavis/polyform/math/noise"
	_ "github.com/EliCDavis/polyform/math/quaternion"
	_ "github.com/EliCDavis/polyform/math/sdf"
	_ "github.com/EliCDavis/polyform/math/sequence"
	_ "github.com/EliCDavis/polyform/math/trig"
	_ "github.com/EliCDavis/polyform/math/trs"
	_ "github.com/EliCDavis/polyform/math/unit"
	_ "github.com/EliCDavis/polyform/math/vector2"
	_ "github.com/EliCDavis/polyform/math/vector3"
	_ "github.com/EliCDavis/polyform/math/vector4"

	_ "github.com/EliCDavis/polyform/modeling"
	_ "github.com/EliCDavis/polyform/modeling/animation"
	_ "github.com/EliCDavis/polyform/modeling/extrude"
	_ "github.com/EliCDavis/polyform/modeling/marching"
	_ "github.com/EliCDavis/polyform/modeling/meshops"
	_ "github.com/EliCDavis/polyform/modeling/meshops/gausops"
	_ "github.com/EliCDavis/polyform/modeling/primitives"
	_ "github.com/EliCDavis/polyform/modeling/repeat"
	_ "github.com/EliCDavis/polyform/modeling/triangulation"
	_ "github.com/EliCDavis/polyform/modeling/voxelize"

	_ "github.com/EliCDavis/polyform/nodes/experimental"
	_ "github.com/EliCDavis/polyform/nodes/opearations"
)

func main() {
	graphPath := flag.String("graph", "", "optional path to an existing graph JSON file to preload")
	logPath := flag.String("log", filepath.Join("tmp", "mcp-logs", "calls-"+time.Now().Format("20060102-150405")+".jsonl"),
		"path to append a JSON-lines log of every tool call to (name, arguments, duration, error status, result preview) - for diagnosing agent behavior after a run. Defaults under tmp/, which is gitignored. Pass an empty string to disable.")
	flag.Parse()

	inst := graph.New(graph.Config{
		TypeFactory:     generator.Types(),
		VariableFactory: mcp.NewTypedVariable,
	})

	if *graphPath != "" {
		data, err := os.ReadFile(*graphPath)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		if err := inst.ApplyAppSchema(data); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	server := mcp.NewServer(inst)

	if *logPath != "" {
		if _, err := server.EnableCallLog(*logPath); err != nil {
			fmt.Fprintln(os.Stderr, "warning: could not enable call log:", err.Error())
		}
	}

	if err := server.Serve(context.Background()); err != nil {
		log.Fatal(err)
	}
}
