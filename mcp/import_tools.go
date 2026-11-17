package mcp

import (
	"context"
	"encoding/json"
	"os"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type ImportSubgraphsInput struct {
	Path  string `json:"path" jsonschema:"saved graph JSON file whose subgraph definitions to import"`
	Group string `json:"group,omitempty" jsonschema:"group the imported subgraphs are listed under; defaults to the saved graph's name"`
}

type ImportedPort struct {
	Type    string `json:"type"`
	Default any    `json:"default,omitempty"`
}

type ImportedSubgraph struct {
	Id         string                  `json:"id"`
	Name       string                  `json:"name"`
	OriginalId string                  `json:"originalId,omitempty"`
	Inputs     map[string]ImportedPort `json:"inputs,omitempty"`
	Outputs    map[string]string       `json:"outputs,omitempty"`
}

type ImportSubgraphsOutput struct {
	Group     string             `json:"group,omitempty"`
	Subgraphs []ImportedSubgraph `json:"subgraphs" jsonschema:"imported subgraphs no other imported subgraph places"`
	Nested    []string           `json:"nested,omitempty" jsonschema:"ids of imported subgraphs placed inside the ones above"`
	Enums     []string           `json:"enums,omitempty"`
}

func (s *Server) importSubgraphs(ctx context.Context, req *mcpsdk.CallToolRequest, in ImportSubgraphsInput) (*mcpsdk.CallToolResult, ImportSubgraphsOutput, error) {
	var out ImportSubgraphsOutput
	var err error
	s.atomic(&err, func() error {
		data, e := os.ReadFile(in.Path)
		if e != nil {
			return e
		}
		result, e := s.graph.ImportSubGraphDefinitions(data, in.Group)
		if e != nil {
			return e
		}
		out.Group = result.Group
		for _, enum := range result.Enums {
			out.Enums = append(out.Enums, enum.ID)
		}
		sch := s.graph.Schema()
		out.Subgraphs = make([]ImportedSubgraph, 0, len(result.Imported))
		for _, imported := range result.Imported {
			if len(subgraphInstances(sch, imported.ID)) > 0 {
				out.Nested = append(out.Nested, imported.ID)
				continue
			}
			entry := ImportedSubgraph{
				Id:         imported.ID,
				Name:       imported.Name,
				OriginalId: imported.OriginalID,
				Inputs:     make(map[string]ImportedPort, len(imported.NodeType.Inputs)),
				Outputs:    make(map[string]string, len(imported.NodeType.Outputs)),
			}
			for name, port := range imported.NodeType.Inputs {
				entry.Inputs[name] = ImportedPort{Type: port.Type, Default: decodeDefault(port.Default)}
			}
			for name, port := range imported.NodeType.Outputs {
				entry.Outputs[name] = port.Type
			}
			out.Subgraphs = append(out.Subgraphs, entry)
		}
		return nil
	})
	return nil, out, err
}

func decodeDefault(raw json.RawMessage) any {
	var v any
	if len(raw) == 0 || json.Unmarshal(raw, &v) != nil {
		return nil
	}
	return v
}
