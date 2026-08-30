package mcp

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"sort"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/generator/schema"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type DescribeGraphInput struct {
	Scope string `json:"scope,omitempty" jsonschema:"the root graph if omitted, or a subgraph id to inspect its interior"`
}

type PortReferenceSummary struct {
	NodeId string `json:"nodeId"`
	Port   string `json:"port"`
}

type NodeInstanceSummary struct {
	Id            string                          `json:"id"`
	Type          string                          `json:"type"`
	Name          string                          `json:"name,omitempty"`
	AssignedInput map[string]PortReferenceSummary `json:"assignedInput,omitempty" jsonschema:"input port name -> the output port feeding it"`
	Outputs       []string                        `json:"outputs,omitempty" jsonschema:"names of this node's output ports"`
	IsParameter   bool                            `json:"isParameter,omitempty"`
	SubGraphId    string                          `json:"subGraphId,omitempty" jsonschema:"set if this node is an instance of a subgraph"`
}

type ProducerSummary struct {
	Name   string `json:"name"`
	NodeId string `json:"nodeId"`
	Port   string `json:"port"`
}

type VariableSummary struct {
	Path        string `json:"path"`
	Description string `json:"description,omitempty"`
	Type        string `json:"type"`
	Value       any    `json:"value"`
}

// DescribeGraphOutput is a flattened, schema-safe view of a graph.Instance's
// state. It deliberately avoids returning generator/schema.Graph directly:
// that type's Variables field is a self-referential tree
// (schema.NestedGroup[Variable]) which the MCP SDK's reflection-based
// output-schema generator cannot represent.
type DescribeGraphOutput struct {
	Name        string                `json:"name,omitempty" jsonschema:"only meaningful for the root graph; set via set_graph_info"`
	Description string                `json:"description,omitempty"`
	Version     string                `json:"version,omitempty"`
	Nodes       []NodeInstanceSummary `json:"nodes"`
	Producers   []ProducerSummary     `json:"producers,omitempty"`
	Variables   []VariableSummary     `json:"variables,omitempty"`
	SubGraphs   []SubgraphSummary     `json:"subGraphs,omitempty"`
}

func summarizeGraph(g schema.Graph) DescribeGraphOutput {
	var out DescribeGraphOutput

	for id, n := range g.Nodes {
		summary := NodeInstanceSummary{
			Id:          id,
			Type:        n.Type,
			Name:        n.Name,
			IsParameter: n.Parameter != nil,
			SubGraphId:  n.SubGraphId,
		}
		for port, ref := range n.AssignedInput {
			if summary.AssignedInput == nil {
				summary.AssignedInput = map[string]PortReferenceSummary{}
			}
			summary.AssignedInput[port] = PortReferenceSummary{NodeId: ref.NodeId, Port: ref.PortName}
		}
		for port := range n.Output {
			summary.Outputs = append(summary.Outputs, port)
		}
		sort.Strings(summary.Outputs)
		out.Nodes = append(out.Nodes, summary)
	}
	sort.Slice(out.Nodes, func(i, j int) bool { return out.Nodes[i].Id < out.Nodes[j].Id })

	for name, p := range g.Producers {
		out.Producers = append(out.Producers, ProducerSummary{Name: name, NodeId: p.NodeID, Port: p.Port})
	}
	sort.Slice(out.Producers, func(i, j int) bool { return out.Producers[i].Name < out.Producers[j].Name })

	g.Variables.Traverse(func(path string, v schema.Variable) bool {
		out.Variables = append(out.Variables, VariableSummary{
			Path:        path,
			Description: v.Description,
			Type:        v.Type,
			Value:       v.Value,
		})
		return true
	})
	sort.Slice(out.Variables, func(i, j int) bool { return out.Variables[i].Path < out.Variables[j].Path })

	for id, sg := range g.SubGraphs {
		out.SubGraphs = append(out.SubGraphs, SubgraphSummary{Id: id, Name: sg.Name, Description: sg.Description})
	}
	sort.Slice(out.SubGraphs, func(i, j int) bool { return out.SubGraphs[i].Id < out.SubGraphs[j].Id })

	return out
}

func (s *Server) describeGraph(ctx context.Context, req *mcpsdk.CallToolRequest, in DescribeGraphInput) (*mcpsdk.CallToolResult, DescribeGraphOutput, error) {
	var out DescribeGraphOutput
	var err error
	s.atomic(&err, func() error {
		inst, e := s.resolveScope(in.Scope)
		if e != nil {
			return e
		}
		out = summarizeGraph(inst.Schema())
		out.Name = inst.GetName()
		out.Description = inst.GetDescription()
		out.Version = inst.GetVersion()
		return nil
	})
	return nil, out, err
}

type SetGraphInfoInput struct {
	Name        string `json:"name,omitempty" jsonschema:"if set, replaces the graph's display name"`
	Description string `json:"description,omitempty" jsonschema:"if set, replaces the graph's description"`
	Version     string `json:"version,omitempty" jsonschema:"if set, replaces the graph's version string"`
}

type SetGraphInfoOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

func (s *Server) setGraphInfo(ctx context.Context, req *mcpsdk.CallToolRequest, in SetGraphInfoInput) (*mcpsdk.CallToolResult, SetGraphInfoOutput, error) {
	var out SetGraphInfoOutput
	var err error
	s.atomic(&err, func() error {
		if in.Name != "" {
			s.graph.SetName(in.Name)
		}
		if in.Description != "" {
			s.graph.SetDescription(in.Description)
		}
		if in.Version != "" {
			s.graph.SetVersion(in.Version)
		}
		out.Name = s.graph.GetName()
		out.Description = s.graph.GetDescription()
		out.Version = s.graph.GetVersion()
		return nil
	})
	return nil, out, err
}

type RenderMermaidInput struct{}

type RenderMermaidOutput struct {
	Mermaid string `json:"mermaid" jsonschema:"a mermaid flowchart of the root graph's nodes and connections, for visual sanity-checking"`
}

func (s *Server) renderMermaid(ctx context.Context, req *mcpsdk.CallToolRequest, in RenderMermaidInput) (*mcpsdk.CallToolResult, RenderMermaidOutput, error) {
	var out RenderMermaidOutput
	var err error
	s.atomic(&err, func() error {
		var buf bytes.Buffer
		if e := graph.WriteMermaid(s.graph, &buf); e != nil {
			return e
		}
		out.Mermaid = buf.String()
		return nil
	})
	return nil, out, err
}

type SaveGraphInput struct {
	Path string `json:"path" jsonschema:"filesystem path to write the graph JSON to; the result is directly loadable in polyform edit/generate"`
}

type SaveGraphOutput struct {
	Path string `json:"path"`
}

func (s *Server) saveGraph(ctx context.Context, req *mcpsdk.CallToolRequest, in SaveGraphInput) (*mcpsdk.CallToolResult, SaveGraphOutput, error) {
	var out SaveGraphOutput
	var err error
	s.atomic(&err, func() error {
		data, e := s.graph.EncodeToAppSchema()
		if e != nil {
			return e
		}
		if e := os.WriteFile(in.Path, data, 0o644); e != nil {
			return e
		}
		out.Path = in.Path
		return nil
	})
	return nil, out, err
}

type LoadGraphInput struct {
	Path string `json:"path" jsonschema:"filesystem path to an existing graph JSON file to load; replaces the current in-memory graph"`
}

type LoadGraphOutput struct {
	Loaded bool `json:"loaded"`
}

func (s *Server) loadGraph(ctx context.Context, req *mcpsdk.CallToolRequest, in LoadGraphInput) (*mcpsdk.CallToolResult, LoadGraphOutput, error) {
	var out LoadGraphOutput
	var err error
	s.atomic(&err, func() error {
		data, e := os.ReadFile(in.Path)
		if e != nil {
			return e
		}
		if e := s.graph.ApplyAppSchema(data); e != nil {
			return e
		}
		out.Loaded = true
		return nil
	})
	return nil, out, err
}

type SetProducerInput struct {
	NodeId string `json:"nodeId" jsonschema:"id of the node whose output produces a manifest artifact, e.g. a gltf.ManifestNode"`
	Port   string `json:"port" jsonschema:"name of the manifest-producing output port"`
	Name   string `json:"name" jsonschema:"the artifact's output name, e.g. 'car.glb'"`
}

type SetProducerOutput struct {
	Name string `json:"name"`
}

func (s *Server) setProducer(ctx context.Context, req *mcpsdk.CallToolRequest, in SetProducerInput) (*mcpsdk.CallToolResult, SetProducerOutput, error) {
	var out SetProducerOutput
	var err error
	s.atomic(&err, func() error {
		s.graph.SetNodeAsProducer(in.NodeId, in.Port, in.Name)
		out.Name = in.Name
		return nil
	})
	return nil, out, err
}

type GenerateInput struct {
	OutputDir string `json:"outputDir" jsonschema:"folder to write every producer's output artifacts into, one subfolder per producer"`
}

type GenerateOutput struct {
	Files []string `json:"files" jsonschema:"paths of every file written"`
}

func (s *Server) generate(ctx context.Context, req *mcpsdk.CallToolRequest, in GenerateInput) (*mcpsdk.CallToolResult, GenerateOutput, error) {
	var out GenerateOutput
	var err error
	s.atomic(&err, func() error {
		if e := graph.WriteToFolder(s.graph, in.OutputDir); e != nil {
			return e
		}
		return filepath.Walk(in.OutputDir, func(p string, info os.FileInfo, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !info.IsDir() {
				out.Files = append(out.Files, p)
			}
			return nil
		})
	})
	return nil, out, err
}

func (s *Server) registerIOTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "describe_graph",
		Description: "Inspect the current state of the graph (or a subgraph's interior): its name/description/version, every node, its type, connections, and named producers.",
	}, s.describeGraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_graph_info",
		Description: "Set the root graph's display name, description, and/or version string (the top-level metadata saved alongside it, shown e.g. in polyform edit). Only fields you provide are changed; omit a field to leave it as-is.",
	}, s.setGraphInfo)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "render_mermaid",
		Description: "Render the root graph as a Mermaid flowchart, for a quick visual sanity check of node structure.",
	}, s.renderMermaid)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "save_graph",
		Description: "Persist the current graph to a JSON file on disk, loadable later with load_graph or in polyform edit/generate.",
	}, s.saveGraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "load_graph",
		Description: "Replace the current in-memory graph with one loaded from a JSON file on disk.",
	}, s.loadGraph)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "set_producer",
		Description: "Name a node's manifest-producing output as a top-level output artifact (e.g. 'car.glb'), so it's included by name when generate runs.",
	}, s.setProducer)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "generate",
		Description: "Execute every manifest-producing output in the graph and write the resulting artifacts to a folder on disk.",
	}, s.generate)
}
