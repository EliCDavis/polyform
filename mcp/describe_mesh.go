package mcp

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/EliCDavis/polyform/modeling"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/vector/vector3"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type Extent struct {
	Min    Vector3Input `json:"min"`
	Max    Vector3Input `json:"max"`
	Size   Vector3Input `json:"size"`
	Center Vector3Input `json:"center"`
}

// MeshPiece is one connected island of the mesh.
type MeshPiece struct {
	Triangles int          `json:"triangles"`
	Size      Vector3Input `json:"size"`
	Center    Vector3Input `json:"center"`
}

type DescribeMeshInput struct {
	NodeId     string  `json:"nodeId" jsonschema:"id of a node whose output is a mesh"`
	Port       string  `json:"port,omitempty" jsonschema:"output port name; omit to use the node's single mesh output"`
	Scope      string  `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id. On its own this evaluates the subgraph DEFINITION, whose boundary inputs read as zero - pass instanceId to measure a real instance."`
	InstanceId string  `json:"instanceId,omitempty" jsonschema:"id of a subgraph instance node in the root graph; evaluates that instance's own copy with its wired inputs applied"`
	WeldRadius float64 `json:"weldRadius,omitempty" jsonschema:"how close two vertices must be to count as the same point when working out which pieces are connected. Defaults to a millionth of the bounding box, which merges the duplicate vertices marching leaves at seams without joining parts that merely sit near each other. Raise it to ask whether two pieces are within a given distance of touching."`
}

type DescribeMeshOutput struct {
	Port      string   `json:"port" jsonschema:"the output port actually read"`
	Triangles int      `json:"triangles"`
	Vertices  int      `json:"vertices"`
	Extent    Extent   `json:"extent" jsonschema:"the whole mesh's bounding box. Compare size against the proportions you intended - a part that came out 3:1 when you wanted 15:1 shows up here without needing a render."`
	Attribute []string `json:"attributes,omitempty" jsonschema:"vertex attributes present, e.g. Position, Normal, Color. A mesh with no Color renders as its flat material color."`

	Pieces       int         `json:"pieces" jsonschema:"how many separate connected islands the mesh is in. This is the cheapest way to answer 'did that part actually attach' - a tooth floating off a jaw, or a fin that failed to union, is its own piece. One shared SDF march that was supposed to be a single body should be 1."`
	LargestFirst []MeshPiece `json:"largestFirst,omitempty" jsonschema:"each piece, biggest first, with where it sits and how big it is. A tiny piece far from the main mass is a part that came adrift."`

	Degenerate string `json:"degenerate,omitempty" jsonschema:"set when the mesh contains non-finite vertices or zero-area triangles, which marching produces from a degenerate field"`
	Note       string `json:"note,omitempty" jsonschema:"set when the mesh is empty, with the usual causes"`

	EvaluationContext string `json:"evaluationContext,omitempty" jsonschema:"present when these numbers came from a subgraph definition rather than a live instance, which makes them unreliable"`
}

// resolveMeshOutput finds the mesh output to read, the same way
// resolveFieldOutput does for fields: the port name varies by node and is
// the detail callers get wrong.
func resolveMeshOutput(node nodes.Node, nodeID, requested string) (nodes.Output[modeling.Mesh], string, error) {
	outputs := node.Outputs()

	if requested != "" {
		name := resolveOutputPortName(node, requested)
		port, ok := outputs[name]
		if !ok {
			return nil, "", fmt.Errorf("node %q has no output port %q; it has %s",
				nodeID, requested, strings.Join(sortedOutputNames(outputs), ", "))
		}
		mesh, ok := port.(nodes.Output[modeling.Mesh])
		if !ok {
			typeName := "unknown"
			if typed, ok := port.(nodes.Typed); ok {
				typeName = typed.Type()
			}
			return nil, "", fmt.Errorf("node %q's %q output is not a mesh, it's %s", nodeID, name, typeName)
		}
		return mesh, name, nil
	}

	found := []string{}
	for name, port := range outputs {
		if _, ok := port.(nodes.Output[modeling.Mesh]); ok {
			found = append(found, name)
		}
	}
	sort.Strings(found)

	switch len(found) {
	case 0:
		return nil, "", fmt.Errorf("node %q has no mesh output; its outputs are %s",
			nodeID, strings.Join(sortedOutputNames(outputs), ", "))
	case 1:
		return outputs[found[0]].(nodes.Output[modeling.Mesh]), found[0], nil
	default:
		return nil, "", fmt.Errorf("node %q has several mesh outputs (%s); pass \"port\" to say which",
			nodeID, strings.Join(found, ", "))
	}
}

type unionFind struct{ parent []int }

func newUnionFind(n int) *unionFind {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &unionFind{parent: p}
}

func (u *unionFind) find(a int) int {
	for u.parent[a] != a {
		u.parent[a] = u.parent[u.parent[a]]
		a = u.parent[a]
	}
	return a
}

func (u *unionFind) union(a, b int) {
	ra, rb := u.find(a), u.find(b)
	if ra != rb {
		u.parent[ra] = rb
	}
}

// meshPieces groups triangles into connected islands.
//
// Vertices are matched by position rather than by index on purpose:
// marching cubes emits a separate vertex per triangle corner, so grouping
// by index would report one piece per triangle and say nothing. Quantizing
// to a grid merges the duplicates at a seam without joining two parts that
// merely sit close together.
func meshPieces(mesh modeling.Mesh, weld float64) []MeshPiece {
	count := mesh.PrimitiveCount()
	if count == 0 {
		return nil
	}

	positions := mesh.Float3Attribute(modeling.PositionAttribute)
	key := func(i int) [3]int64 {
		p := positions.At(i)
		return [3]int64{
			int64(math.Round(p.X() / weld)),
			int64(math.Round(p.Y() / weld)),
			int64(math.Round(p.Z() / weld)),
		}
	}

	// One representative triangle per welded vertex position; joining a
	// triangle to that representative connects everything sharing it.
	seen := make(map[[3]int64]int, count*3)
	uf := newUnionFind(count)
	for t := 0; t < count; t++ {
		tri := mesh.Tri(t)
		for _, v := range []int{tri.P1(), tri.P2(), tri.P3()} {
			k := key(v)
			if prev, ok := seen[k]; ok {
				uf.union(t, prev)
			} else {
				seen[k] = t
			}
		}
	}

	type acc struct {
		tris     int
		min, max vector3.Float64
	}
	groups := map[int]*acc{}
	for t := 0; t < count; t++ {
		root := uf.find(t)
		tri := mesh.Tri(t)
		g, ok := groups[root]
		if !ok {
			g = &acc{
				min: vector3.New(math.Inf(1), math.Inf(1), math.Inf(1)),
				max: vector3.New(math.Inf(-1), math.Inf(-1), math.Inf(-1)),
			}
			groups[root] = g
		}
		g.tris++
		for _, v := range []int{tri.P1(), tri.P2(), tri.P3()} {
			p := positions.At(v)
			g.min = vector3.New(math.Min(g.min.X(), p.X()), math.Min(g.min.Y(), p.Y()), math.Min(g.min.Z(), p.Z()))
			g.max = vector3.New(math.Max(g.max.X(), p.X()), math.Max(g.max.Y(), p.Y()), math.Max(g.max.Z(), p.Z()))
		}
	}

	out := make([]MeshPiece, 0, len(groups))
	for _, g := range groups {
		size := g.max.Sub(g.min)
		center := g.min.Add(size.Scale(0.5))
		out = append(out, MeshPiece{
			Triangles: g.tris,
			Size:      Vector3Input{X: size.X(), Y: size.Y(), Z: size.Z()},
			Center:    Vector3Input{X: center.X(), Y: center.Y(), Z: center.Z()},
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Triangles > out[j].Triangles })
	return out
}

// describeMesh reports what a mesh actually is, in numbers.
//
// A render answers "does this look right" and is the only way to judge
// form, but it is bad at the question that keeps costing builds real time:
// is the part there at all, and is it attached. An absent part among
// fifteen present ones looks like nothing, and a tooth floating a
// millimetre off a jaw looks fine from every angle that isn't edge-on.
// Both are unmistakable here - a missing part doesn't change the triangle
// count, and a detached one is its own piece.
//
// It is also cheaper than the render it replaces: the marching cost is the
// same, but there is no rasterizing, no PNG, and no image to read back.
func (s *Server) describeMesh(ctx context.Context, req *mcpsdk.CallToolRequest, in DescribeMeshInput) (*mcpsdk.CallToolResult, DescribeMeshOutput, error) {
	var out DescribeMeshOutput
	var err error
	s.atomic(&err, func() error {
		inst, contextWarning, e := s.resolveFieldGraph(in.Scope, in.InstanceId)
		if e != nil {
			return e
		}
		node := inst.Node(in.NodeId)
		if node == nil {
			return fmt.Errorf("no node exists with id %q", in.NodeId)
		}

		meshOut, portName, e := resolveMeshOutput(node, in.NodeId, in.Port)
		if e != nil {
			return e
		}
		out.Port = portName
		out.EvaluationContext = contextWarning

		mesh := meshOut.Value()
		out.Triangles = mesh.PrimitiveCount()
		out.Attribute = mesh.Float3Attributes()
		sort.Strings(out.Attribute)

		if !mesh.HasFloat3Attribute(modeling.PositionAttribute) || out.Triangles == 0 {
			out.Note = "this mesh is empty. From a March node that usually means the Resolution is too coarse for the Domain, or the field never crosses the surface inside it; from a boolean it usually means the two shapes don't overlap the way you expected."
			return nil
		}

		box := mesh.BoundingBox(modeling.PositionAttribute)
		size := box.Size()
		out.Extent = Extent{
			Min:    Vector3Input{X: box.Min().X(), Y: box.Min().Y(), Z: box.Min().Z()},
			Max:    Vector3Input{X: box.Max().X(), Y: box.Max().Y(), Z: box.Max().Z()},
			Size:   Vector3Input{X: size.X(), Y: size.Y(), Z: size.Z()},
			Center: Vector3Input{X: box.Center().X(), Y: box.Center().Y(), Z: box.Center().Z()},
		}

		positions := mesh.Float3Attribute(modeling.PositionAttribute)
		out.Vertices = positions.Len()

		nonFinite := 0
		for i := 0; i < positions.Len(); i++ {
			p := positions.At(i)
			if math.IsNaN(p.X()) || math.IsNaN(p.Y()) || math.IsNaN(p.Z()) ||
				math.IsInf(p.X(), 0) || math.IsInf(p.Y(), 0) || math.IsInf(p.Z(), 0) {
				nonFinite++
			}
		}
		if nonFinite > 0 {
			out.Degenerate = fmt.Sprintf("%d of %d vertices are not finite, so the field feeding this mesh is degenerate somewhere. sample_field the inputs of the node that produced it to find which one.", nonFinite, positions.Len())
			return nil
		}

		weld := in.WeldRadius
		if weld <= 0 {
			diagonal := size.Length()
			if diagonal == 0 {
				diagonal = 1
			}
			weld = diagonal * 1e-6
		}
		pieces := meshPieces(mesh, weld)
		out.Pieces = len(pieces)

		// Enough to spot the stray piece without pasting a whole part list.
		const maxListed = 24
		if len(pieces) > maxListed {
			out.LargestFirst = pieces[:maxListed]
		} else {
			out.LargestFirst = pieces
		}
		return nil
	})
	return nil, out, err
}
