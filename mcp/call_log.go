package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type callLogEntry struct {
	Time          string          `json:"time"`
	Tool          string          `json:"tool"`
	DurationMs    int64           `json:"durationMs"`
	Arguments     json.RawMessage `json:"arguments,omitempty"`
	IsError       bool            `json:"isError,omitempty"`
	Error         string          `json:"error,omitempty"`
	ResultBytes   int             `json:"resultBytes,omitempty"`
	ResultPreview string          `json:"resultPreview,omitempty"`
}

const callLogPreviewLimit = 400

// EnableCallLog appends one JSON line per tools/call request (tool name,
// exact raw arguments as sent by the client, duration, error status, and a
// truncated preview of the result) to the file at path, creating it and any
// parent directories as needed. Call this before Serve/Connect. It returns
// a close function the caller may use to release the file handle (a
// long-running server process can safely ignore it and let process exit
// reclaim it, but anything that needs the file handle released
// deterministically, e.g. a test, should call it).
//
// This exists purely for diagnosing agent behavior after the fact: an MCP
// client's own conversation transcript doesn't reliably survive past the
// session that produced it (background-agent output files in particular
// are ephemeral and get cleared), so there was previously no durable way to
// answer questions like "did this build actually use search_node_types'
// regex option" or "exactly how many create_node calls did it make, and
// for what types" without the raw arguments of every call. This log is
// that durable record, written server-side, independent of whatever the
// calling client retains.
func (s *Server) EnableCallLog(path string) (func() error, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}

	var mu sync.Mutex
	enc := json.NewEncoder(f)

	s.sdk.AddReceivingMiddleware(func(next mcpsdk.MethodHandler) mcpsdk.MethodHandler {
		return func(ctx context.Context, method string, req mcpsdk.Request) (mcpsdk.Result, error) {
			if method != "tools/call" {
				return next(ctx, method, req)
			}

			start := time.Now()
			result, err := next(ctx, method, req)

			entry := callLogEntry{
				Time:       start.UTC().Format(time.RFC3339Nano),
				DurationMs: time.Since(start).Milliseconds(),
			}

			if params, ok := req.GetParams().(*mcpsdk.CallToolParamsRaw); ok {
				entry.Tool = params.Name
				entry.Arguments = params.Arguments
			}

			if err != nil {
				entry.Error = err.Error()
			} else if callResult, ok := result.(*mcpsdk.CallToolResult); ok {
				entry.IsError = callResult.IsError
				if toolErr := callResult.GetError(); toolErr != nil {
					// A tool-level failure (the handler returned an error,
					// not a transport-level one) - the SDK clears
					// StructuredContent in this case and puts the message
					// in Content instead, but GetError() gives it back
					// directly without needing to unpack Content.
					entry.Error = toolErr.Error()
				}
				if b, e := json.Marshal(callResult.StructuredContent); e == nil {
					entry.ResultBytes = len(b)
					if len(b) > callLogPreviewLimit {
						entry.ResultPreview = string(b[:callLogPreviewLimit]) + "..."
					} else {
						entry.ResultPreview = string(b)
					}
				}
			}

			mu.Lock()
			_ = enc.Encode(entry)
			mu.Unlock()

			return result, err
		}
	})

	return f.Close, nil
}
