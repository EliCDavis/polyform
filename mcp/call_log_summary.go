package mcp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CallLogSummary condenses a calls-*.jsonl log into the numbers worth
// comparing between runs: how much work, how much of it failed, and
// where the failures clustered.
type CallLogSummary struct {
	File          string              `json:"file"`
	Calls         int                 `json:"calls"`
	WallClock     string              `json:"wallClock"`
	ToolTime      string              `json:"toolTime" jsonschema:"sum of every call's duration"`
	Errors        int                 `json:"errors" jsonschema:"calls that returned a tool error"`
	SoftErrors    int                 `json:"softErrors" jsonschema:"successful batch calls whose result carried per-entry errors"`
	ErrorRate     float64             `json:"errorRate"`
	ByTool        map[string]ToolStat `json:"byTool"`
	ErrorsByTool  map[string]int      `json:"errorsByTool,omitempty"`
	NodesCreated  int                 `json:"nodesCreated" jsonschema:"across create_node, create_nodes and create_subgraph"`
	Subgraphs     []string            `json:"subgraphs" jsonschema:"ids from create_subgraph and the helper subgraph tools, in order"`
	Converted     int                 `json:"converted" jsonschema:"convert_to_subgraph calls"`
	Instantiated  map[string]int      `json:"instantiated,omitempty" jsonschema:"subgraph id -> instantiate_subgraph calls"`
	Renders       int                 `json:"renders"`
	OutlineSet    int                 `json:"outlineSet" jsonschema:"set_outline calls; more than one means the plan was revised (or rejected)"`
	OutlineChecks int                 `json:"outlineChecks"`
	Projects      []string            `json:"projects,omitempty" jsonschema:"start_project paths, in order; more than one distinct path, or a repeat, is a resume or a second agent"`
	FirstErrors   []string            `json:"firstErrors,omitempty" jsonschema:"the first few error messages, verbatim, to see what kind of trouble it was"`
	Repeated      map[string]int      `json:"repeatedCalls,omitempty" jsonschema:"identical tool+arguments sent more than once - a retry loop or two agents replaying each other"`
	Timeline      []CallLogInterval   `json:"timeline,omitempty" jsonschema:"calls bucketed by 10-minute interval, to see where time went"`
	Slowest       []SlowCall          `json:"slowest,omitempty" jsonschema:"the individual calls that took longest, worst first"`
}

// ToolStat is how much of a run one tool accounted for. Counts alone hide
// the expensive tools: a handful of calls can outweigh hundreds of cheap
// ones, and the count makes them look negligible.
type ToolStat struct {
	Calls   int     `json:"calls"`
	Total   string  `json:"total"`
	Mean    string  `json:"mean"`
	Slowest string  `json:"slowest" jsonschema:"the longest single call to this tool"`
	Share   float64 `json:"share" jsonschema:"percent of the run's total tool time"`

	totalMs int64
	maxMs   int64
}

type SlowCall struct {
	Tool      string `json:"tool"`
	Took      string `json:"took"`
	Arguments string `json:"arguments,omitempty"`

	ms int64
}

const slowestCallsReported = 10

type CallLogInterval struct {
	Start  string `json:"start"`
	Calls  int    `json:"calls"`
	Errors int    `json:"errors"`
}

var softErrorPattern = regexp.MustCompile(`"errors":\[`)

func SummarizeCallLog(path string) (CallLogSummary, error) {
	f, err := os.Open(path)
	if err != nil {
		return CallLogSummary{}, err
	}
	defer f.Close()
	return summarizeCallLog(path, f)
}

func summarizeCallLog(name string, r io.Reader) (CallLogSummary, error) {
	s := CallLogSummary{
		File:         name,
		ByTool:       map[string]ToolStat{},
		ErrorsByTool: map[string]int{},
		Instantiated: map[string]int{},
		Repeated:     map[string]int{},
		Subgraphs:    []string{},
	}
	var first, last time.Time
	var toolTime time.Duration
	var slowest []SlowCall
	seen := map[string]int{}
	buckets := map[time.Time]*CallLogInterval{}

	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 1024*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var entry callLogEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			continue
		}
		s.Calls++
		toolTime += time.Duration(entry.DurationMs) * time.Millisecond

		stat := s.ByTool[entry.Tool]
		stat.Calls++
		stat.totalMs += entry.DurationMs
		if entry.DurationMs > stat.maxMs {
			stat.maxMs = entry.DurationMs
		}
		s.ByTool[entry.Tool] = stat

		slowest = append(slowest, SlowCall{
			Tool:      entry.Tool,
			Took:      (time.Duration(entry.DurationMs) * time.Millisecond).Round(time.Millisecond).String(),
			Arguments: truncate(string(entry.Arguments), 160),
			ms:        entry.DurationMs,
		})

		if t, err := time.Parse(time.RFC3339Nano, entry.Time); err == nil {
			if first.IsZero() || t.Before(first) {
				first = t
			}
			if t.After(last) {
				last = t
			}
			b := t.Truncate(10 * time.Minute)
			if buckets[b] == nil {
				buckets[b] = &CallLogInterval{Start: b.Format("15:04")}
			}
			buckets[b].Calls++
			if entry.IsError {
				buckets[b].Errors++
			}
		}

		key := entry.Tool + " " + string(entry.Arguments)
		seen[key]++

		var args map[string]any
		_ = json.Unmarshal(entry.Arguments, &args)
		str := func(k string) string { v, _ := args[k].(string); return v }

		if entry.IsError {
			s.Errors++
			s.ErrorsByTool[entry.Tool]++
			if len(s.FirstErrors) < 8 {
				s.FirstErrors = append(s.FirstErrors, fmt.Sprintf("%s: %s", entry.Tool, truncate(entry.Error, 160)))
			}
		} else if softErrorPattern.MatchString(entry.ResultPreview) {
			s.SoftErrors++
		}

		switch entry.Tool {
		case "create_node":
			s.NodesCreated++
		case "create_nodes":
			if nodes, ok := args["nodes"].([]any); ok {
				s.NodesCreated += len(nodes)
			}
		case "create_subgraph":
			if nodes, ok := args["nodes"].([]any); ok {
				s.NodesCreated += len(nodes)
			}
			s.Subgraphs = append(s.Subgraphs, str("id"))
		case "create_equation_subgraph", "create_tapered_curve_subgraph", "create_vertex_color_gradient_subgraph",
			"create_flush_position_subgraph", "create_sphere_surface_point_subgraph":
			s.Subgraphs = append(s.Subgraphs, str("id")+" ("+strings.TrimSuffix(strings.TrimPrefix(entry.Tool, "create_"), "_subgraph")+")")
		case "convert_to_subgraph":
			s.Converted++
		case "instantiate_subgraph":
			s.Instantiated[str("subgraphId")]++
		case "render_preview":
			s.Renders++
		case "set_outline":
			s.OutlineSet++
		case "check_outline":
			s.OutlineChecks++
		case "start_project":
			s.Projects = append(s.Projects, str("path"))
		}
	}
	if err := scanner.Err(); err != nil {
		return s, err
	}

	for key, n := range seen {
		if n > 1 {
			s.Repeated[truncate(key, 120)] = n
		}
	}
	if s.Calls > 0 {
		s.ErrorRate = float64(s.Errors+s.SoftErrors) / float64(s.Calls)
	}
	if !first.IsZero() {
		s.WallClock = last.Sub(first).Round(time.Second).String()
	}
	s.ToolTime = toolTime.Round(time.Second).String()

	for name, stat := range s.ByTool {
		total := time.Duration(stat.totalMs) * time.Millisecond
		stat.Total = total.Round(time.Millisecond).String()
		stat.Mean = (total / time.Duration(stat.Calls)).Round(time.Millisecond).String()
		stat.Slowest = (time.Duration(stat.maxMs) * time.Millisecond).Round(time.Millisecond).String()
		if toolTime > 0 {
			stat.Share = math.Round(1000*float64(stat.totalMs)/float64(toolTime.Milliseconds())) / 10
		}
		s.ByTool[name] = stat
	}

	sort.Slice(slowest, func(i, j int) bool { return slowest[i].ms > slowest[j].ms })
	if len(slowest) > slowestCallsReported {
		slowest = slowest[:slowestCallsReported]
	}
	s.Slowest = slowest

	starts := make([]time.Time, 0, len(buckets))
	for t := range buckets {
		starts = append(starts, t)
	}
	sort.Slice(starts, func(i, j int) bool { return starts[i].Before(starts[j]) })
	for _, t := range starts {
		s.Timeline = append(s.Timeline, *buckets[t])
	}
	return s, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
