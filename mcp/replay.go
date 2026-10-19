package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// Read-only calls are not skipped: what they evaluated can be part of a bug's state.
var replaySkipped = map[string]bool{
	"start_project":     true,
	"save_graph":        true,
	"generate":          true,
	"report_issue":      true,
	"run_variant_sweep": true,
}

func redirectRenderOutput(arguments json.RawMessage, dir string, line int) json.RawMessage {
	var args map[string]any
	if json.Unmarshal(arguments, &args) != nil {
		return arguments
	}
	args["outputPath"] = filepath.Join(dir, fmt.Sprintf("%d.png", line))
	redirected, err := json.Marshal(args)
	if err != nil {
		return arguments
	}
	return redirected
}

type ReplayDivergence struct {
	Line        int    `json:"line"`
	Tool        string `json:"tool"`
	LoggedError string `json:"loggedError,omitempty"`
	ReplayError string `json:"replayError,omitempty"`
}

type ReplayTarget struct {
	Line      int             `json:"line"`
	Tool      string          `json:"tool"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
	Error     string          `json:"error,omitempty"`
	Result    string          `json:"result,omitempty"`
}

type ReplayReport struct {
	Replayed    int                `json:"replayed"`
	Skipped     int                `json:"skipped"`
	Divergences []ReplayDivergence `json:"divergences,omitempty"`
	Target      *ReplayTarget      `json:"target,omitempty"`
}

const replayResultLimit = 4000

// Replay feeds the calls logged at logPath through s in order. With until
// set it stops before that 1-based line, saves the graph to savePath when
// given, then runs that line as the target; with until 0 it replays every
// call and saves at the end.
func Replay(ctx context.Context, s *Server, logPath string, until int, savePath string) (ReplayReport, error) {
	var report ReplayReport
	f, err := os.Open(logPath)
	if err != nil {
		return report, err
	}
	defer f.Close()

	clientTransport, serverTransport := mcpsdk.NewInMemoryTransports()
	if _, err := s.Connect(ctx, serverTransport, nil); err != nil {
		return report, err
	}
	session, err := mcpsdk.NewClient(&mcpsdk.Implementation{Name: "replay", Version: "0"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		return report, err
	}
	defer session.Close()

	renders, err := os.MkdirTemp("", "polyform-replay-renders")
	if err != nil {
		return report, err
	}
	defer os.RemoveAll(renders)

	save := func() error {
		if savePath == "" {
			return nil
		}
		if _, e := callForReplay(ctx, session, "save_graph", json.RawMessage(fmt.Sprintf(`{"path":%q}`, savePath))); e != "" {
			return fmt.Errorf("saving the replayed graph: %s", e)
		}
		return nil
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1<<20), 1<<30)
	for line := 1; scanner.Scan(); line++ {
		var entry callLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			return report, fmt.Errorf("line %d: %w", line, err)
		}
		if entry.Tool == "render_preview" {
			entry.Arguments = redirectRenderOutput(entry.Arguments, renders, line)
		}
		if until > 0 && line == until {
			if err := save(); err != nil {
				return report, err
			}
			result, callErr := callForReplay(ctx, session, entry.Tool, entry.Arguments)
			if len(result) > replayResultLimit {
				result = result[:replayResultLimit] + "..."
			}
			report.Target = &ReplayTarget{Line: line, Tool: entry.Tool, Arguments: entry.Arguments, Error: callErr, Result: result}
			return report, nil
		}
		if replaySkipped[entry.Tool] {
			report.Skipped++
			continue
		}
		_, callErr := callForReplay(ctx, session, entry.Tool, entry.Arguments)
		report.Replayed++
		if (callErr == "") != (entry.Error == "") {
			report.Divergences = append(report.Divergences, ReplayDivergence{Line: line, Tool: entry.Tool, LoggedError: entry.Error, ReplayError: callErr})
		}
	}
	if err := scanner.Err(); err != nil {
		return report, err
	}
	if until > 0 {
		return report, fmt.Errorf("the log has no line %d", until)
	}
	return report, save()
}

// callForReplay returns the call's text result, or its error message.
func callForReplay(ctx context.Context, session *mcpsdk.ClientSession, tool string, arguments json.RawMessage) (string, string) {
	if len(arguments) == 0 {
		arguments = json.RawMessage(`{}`)
	}
	res, err := session.CallTool(ctx, &mcpsdk.CallToolParams{Name: tool, Arguments: arguments})
	if err != nil {
		return "", err.Error()
	}
	var text string
	for _, c := range res.Content {
		if tc, ok := c.(*mcpsdk.TextContent); ok {
			text += tc.Text
		}
	}
	if res.IsError {
		return "", text
	}
	return text, ""
}
