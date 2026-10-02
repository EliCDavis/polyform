package mcp

import (
	"context"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type UndoInput struct {
	Steps int `json:"steps,omitempty" jsonschema:"how many steps to walk back; defaults to 1"`
}

type HistoryOutput struct {
	Undone  []string `json:"undone,omitempty" jsonschema:"the steps walked back, most recent first"`
	Redone  []string `json:"redone,omitempty" jsonschema:"the steps re-applied, oldest first"`
	CanUndo []string `json:"canUndo" jsonschema:"steps still available to undo, most recent first"`
	CanRedo []string `json:"canRedo" jsonschema:"steps available to redo, most recent first"`
}

func (s *Server) historyOutput() HistoryOutput {
	state := s.graph.History()
	return HistoryOutput{CanUndo: state.Undo, CanRedo: state.Redo}
}

// Undo and redo walk the history themselves rather than going through
// atomic, which would record the walk as another step.
func (s *Server) undo(ctx context.Context, req *mcpsdk.CallToolRequest, in UndoInput) (*mcpsdk.CallToolResult, HistoryOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	steps := max(in.Steps, 1)
	var out HistoryOutput
	for range steps {
		if !s.graph.CanUndo() {
			break
		}
		label, err := s.graph.Undo()
		if err != nil {
			return nil, out, err
		}
		out.Undone = append(out.Undone, label)
	}

	state := s.graph.History()
	out.CanUndo, out.CanRedo = state.Undo, state.Redo
	s.autosaveLocked()
	return nil, out, nil
}

func (s *Server) redo(ctx context.Context, req *mcpsdk.CallToolRequest, in UndoInput) (*mcpsdk.CallToolResult, HistoryOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	steps := max(in.Steps, 1)
	var out HistoryOutput
	for range steps {
		if !s.graph.CanRedo() {
			break
		}
		label, err := s.graph.Redo()
		if err != nil {
			return nil, out, err
		}
		out.Redone = append(out.Redone, label)
	}

	state := s.graph.History()
	out.CanUndo, out.CanRedo = state.Undo, state.Redo
	s.autosaveLocked()
	return nil, out, nil
}

type ListHistoryInput struct{}

func (s *Server) listHistory(ctx context.Context, req *mcpsdk.CallToolRequest, in ListHistoryInput) (*mcpsdk.CallToolResult, HistoryOutput, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil, s.historyOutput(), nil
}

func (s *Server) registerHistoryTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "undo",
		Description: "Walk the graph back to before the last tool call that changed it, named by the tool that made the change. Use it to abandon an experiment - a placement that looked wrong in the render, a subgraph extraction that grouped the wrong nodes - instead of hand-reversing each edit, which is how a graph ends up with orphaned literals and half-removed wiring. A failed call is already rolled back on its own and is not a step, so undo always lands on the last thing that actually took effect.",
	}, s.undo)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "redo",
		Description: "Re-apply steps taken back by undo, most recent first. Changing the graph after an undo abandons whatever was left to redo.",
	}, s.redo)

	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "list_history",
		Description: "The steps available to undo and redo, most recent first, without changing anything.",
	}, s.listHistory)
}
