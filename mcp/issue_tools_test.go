package mcp_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	polyformmcp "github.com/EliCDavis/polyform/mcp"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func issueArgs(area, title, workaround string) map[string]any {
	return map[string]any{
		"kind":        "missing",
		"severity":    "major",
		"area":        area,
		"title":       title,
		"detail":      "searched 'atan2', no results",
		"workaround":  workaround,
		"observedVia": "search_node_types",
	}
}

func issueDB(t *testing.T, root string) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(root, "issues.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestReportIssue_StoresTheArgumentsOfTheCallItWasObservedThrough(t *testing.T) {
	root := tempOutputRoot(t)
	session := testSession(t)
	callTool(t, session, "search_node_types", map[string]any{"query": "atan2"}, nil)

	var reported polyformmcp.ReportIssueOutput
	callTool(t, session, "report_issue", issueArgs("math", "no atan2 node", "hard-coded angle"), &reported)

	var via, call string
	require.NoError(t, issueDB(t, root).QueryRow(`SELECT observed_via, observed_call FROM issues WHERE id = ?`, reported.ID).Scan(&via, &call))
	assert.Equal(t, "search_node_types", via)
	assert.JSONEq(t, `{"query": "atan2"}`, call)
}

func TestReportIssue_RequiresTheObservingToolCall(t *testing.T) {
	tempOutputRoot(t)
	args := issueArgs("math", "no atan2 node", "hard-coded angle")
	args["observedVia"] = " "

	assert.Contains(t, callToolExpectingError(t, testSession(t), "report_issue", args), "observedVia is required")
}

func TestReportIssue_AddsTheObservedColumnsToAnOlderDatabase(t *testing.T) {
	root := tempOutputRoot(t)
	require.NoError(t, os.MkdirAll(root, 0o755))
	_, err := issueDB(t, root).Exec(`CREATE TABLE issues (id INTEGER PRIMARY KEY, created_at TEXT NOT NULL, project TEXT NOT NULL, kind TEXT NOT NULL,
		severity TEXT NOT NULL, area TEXT NOT NULL, title TEXT NOT NULL, detail TEXT NOT NULL, workaround TEXT NOT NULL,
		same_as INTEGER REFERENCES issues(id), status TEXT NOT NULL DEFAULT 'open');
		INSERT INTO issues (created_at, project, kind, severity, area, title, detail, workaround)
		VALUES ('2026-10-01T00:00:00Z', '', 'bug', 'minor', 'render_preview', 'an old report', 'detail', 'none')`)
	require.NoError(t, err)

	callTool(t, testSession(t), "report_issue", issueArgs("math", "no atan2 node", "hard-coded angle"), &polyformmcp.ReportIssueOutput{})

	var vias []string
	rows, err := issueDB(t, root).Query(`SELECT observed_via FROM issues ORDER BY id`)
	require.NoError(t, err)
	defer rows.Close()
	for rows.Next() {
		var via string
		require.NoError(t, rows.Scan(&via))
		vias = append(vias, via)
	}
	assert.Equal(t, []string{"", "search_node_types"}, vias)
}

func TestReportIssue_SameAsARecurrencePointsAtTheFirstReport(t *testing.T) {
	root := tempOutputRoot(t)
	session := testSession(t)

	var first, second, third polyformmcp.ReportIssueOutput
	callTool(t, session, "report_issue", issueArgs("math", "no atan2 node", "a"), &first)
	args := issueArgs("math", "no atan2 node", "b")
	args["sameAs"] = first.ID
	callTool(t, session, "report_issue", args, &second)
	args["sameAs"] = second.ID
	callTool(t, session, "report_issue", args, &third)

	var sameAs int64
	require.NoError(t, issueDB(t, root).QueryRow(`SELECT same_as FROM issues WHERE id = ?`, third.ID).Scan(&sameAs))
	assert.Equal(t, first.ID, sameAs)
}

func TestReportIssue_RequiresWorkaroundAndKnownKind(t *testing.T) {
	tempOutputRoot(t)
	session := testSession(t)

	msg := callToolExpectingError(t, session, "report_issue", issueArgs("math", "no atan2 node", " "))
	require.Contains(t, msg, "workaround is required")

	args := issueArgs("math", "no atan2 node", "x")
	args["kind"] = "feature"
	msg = callToolExpectingError(t, session, "report_issue", args)
	require.Contains(t, msg, "kind must be one of")

	args = issueArgs("math", "no atan2 node", "x")
	args["sameAs"] = 99
	msg = callToolExpectingError(t, session, "report_issue", args)
	require.Contains(t, msg, "no issue with id 99")
}

func TestReportIssue_ShowsAgentsNothingFromTheDatabase(t *testing.T) {
	tempOutputRoot(t)
	session := testSession(t)
	callTool(t, session, "report_issue", issueArgs("math", "no atan2 node", "a"), nil)

	tools, err := session.ListTools(context.Background(), &mcpsdk.ListToolsParams{})
	require.NoError(t, err)
	for _, tool := range tools.Tools {
		assert.NotEqual(t, "list_issues", tool.Name)
	}

	res, err := session.CallTool(context.Background(), &mcpsdk.CallToolParams{Name: "report_issue", Arguments: issueArgs("math", "no atan2 node", "b")})
	require.NoError(t, err)
	raw, err := json.Marshal(res.StructuredContent)
	require.NoError(t, err)
	var fields map[string]any
	require.NoError(t, json.Unmarshal(raw, &fields))
	assert.Len(t, fields, 1)
	assert.Contains(t, fields, "id")
}

func TestReportIssue_SavesTheGraphAsItWasWhenReported(t *testing.T) {
	root := tempOutputRoot(t)
	session := testSession(t)
	var made polyformmcp.CreateNodesOutput
	callTool(t, session, "create_nodes", map[string]any{"nodes": []any{
		map[string]any{"alias": "cube", "type": cubeNodeType},
	}}, &made)
	callTool(t, session, "search_node_types", map[string]any{"query": "atan2"}, nil)

	var reported polyformmcp.ReportIssueOutput
	callTool(t, session, "report_issue", issueArgs("math", "no atan2 node", "hard-coded angle"), &reported)
	callTool(t, session, "delete_node", map[string]any{"nodeId": made.Nodes["cube"]}, nil)

	var snapshot string
	require.NoError(t, issueDB(t, root).QueryRow(`SELECT graph_snapshot FROM issues WHERE id = ?`, reported.ID).Scan(&snapshot))
	require.NotEmpty(t, snapshot)
	assert.Equal(t, filepath.Join(root, "issue-snapshots"), filepath.Dir(snapshot))

	reloaded := testSession(t)
	callTool(t, reloaded, "load_graph", map[string]any{"path": snapshot}, nil)
	var graph map[string]any
	callTool(t, reloaded, "describe_graph", map[string]any{}, &graph)
	data, _ := json.Marshal(graph)
	assert.Contains(t, string(data), made.Nodes["cube"])
}
