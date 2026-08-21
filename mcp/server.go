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
// error result instead of crashing the whole server process.
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
	}()
	*err = f()
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
