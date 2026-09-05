// Package mcp exposes polyform's existing node-graph construction API
// (github.com/EliCDavis/polyform/generator/graph) as a set of tools callable
// over the Model Context Protocol. It has no knowledge of language models,
// prompts, or agents: it only adapts existing, non-LLM graph.Instance
// methods (CreateNode, ConnectNodes, CreateSubGraph, ...) into MCP tool
// calls. Whatever connects to this server over stdio decides what to build.
package mcp

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/EliCDavis/polyform/generator/graph"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Server wraps a single polyform graph.Instance and exposes tools for
// building it over MCP.
type Server struct {
	graph *graph.Instance
	sdk   *mcpsdk.Server
	mu    sync.Mutex

	// projectDir, once set by start_project, is where autosaveLocked writes
	// after every successful tool call. Empty means no project is active
	// and autosaving is disabled.
	projectDir string

	// suppressNextAutosave skips exactly one post-call autosave - set right
	// after save_graph deliberately deletes the autosave file, so that
	// deletion isn't immediately undone by atomic()'s own post-call
	// autosave for that same call.
	suppressNextAutosave bool
}

// NewServer creates an MCP server backed by the given graph instance. The
// instance is mutated in place by the tools this server registers, so
// callers should not otherwise touch it concurrently.
func NewServer(g *graph.Instance) *Server {
	s := &Server{graph: g}
	s.sdk = mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "polyform",
		Version: "0.1.0",
	}, nil)
	s.registerNodeTools()
	s.registerSubgraphTools()
	s.registerIOTools()
	s.registerRenderTools()
	s.registerVariableTools()
	s.registerVariantSetTools()
	s.registerProjectTools()
	s.registerEquationTools()
	s.registerFieldTools()
	s.registerTaperedCurveTools()
	s.registerVertexColorGradientTools()
	s.registerFlushPositionTools()
	s.registerSphereSurfacePointTools()
	return s
}

// Serve runs the server over stdin/stdout until the client disconnects or
// ctx is cancelled.
func (s *Server) Serve(ctx context.Context) error {
	return s.sdk.Run(ctx, &mcpsdk.StdioTransport{})
}

// Connect binds the server to an arbitrary MCP transport (e.g. an
// in-memory transport for tests) rather than stdio. Most callers want
// Serve instead.
func (s *Server) Connect(ctx context.Context, t mcpsdk.Transport, opts *mcpsdk.ServerSessionOptions) (*mcpsdk.ServerSession, error) {
	return s.sdk.Connect(ctx, t, opts)
}

// atomic runs f while holding the server's lock, converting any panic
// raised by the underlying graph.Instance (many of its methods panic on
// invalid node/port names rather than returning an error) into a plain
// error result instead of crashing the whole server process. On success,
// it also autosaves to the active project (if any) - this is the single
// chokepoint every tool call already passes through, so hooking it here
// covers every mutation without touching each tool individually.
func (s *Server) atomic(err *error, f func() error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				*err = e
			} else {
				*err = fmt.Errorf("%v", r)
			}
		}
		if *err == nil {
			s.autosaveLocked()
		}
	}()
	*err = f()
}

// autosaveLocked writes the current graph state to the active project's
// autosave file. A no-op if no project is active. Best-effort: a failure
// here doesn't fail the tool call that triggered it - the safety net
// having a problem shouldn't turn an otherwise-successful graph mutation
// into a reported error. Caller must already hold s.mu.
func (s *Server) autosaveLocked() {
	if s.suppressNextAutosave {
		s.suppressNextAutosave = false
		return
	}
	if s.projectDir == "" {
		return
	}

	data, err := s.graph.EncodeToAppSchema()
	if err != nil {
		return
	}

	path := filepath.Join(s.projectDir, autosaveFileName)
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return
	}
	os.Rename(tmp, path)
}

// resolveScope returns the graph.Instance a tool call should operate on:
// the root graph when scope is empty, or the named subgraph's own instance
// otherwise.
func (s *Server) resolveScope(scope string) (*graph.Instance, error) {
	if scope == "" {
		return s.graph, nil
	}
	return s.graph.SubGraphInstance(scope)
}
