package mcp

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	_ "modernc.org/sqlite"
)

const issueSchema = `
CREATE TABLE IF NOT EXISTS issues (
	id         INTEGER PRIMARY KEY,
	created_at TEXT NOT NULL,
	project    TEXT NOT NULL,
	kind       TEXT NOT NULL,
	severity   TEXT NOT NULL,
	area       TEXT NOT NULL,
	title      TEXT NOT NULL,
	detail     TEXT NOT NULL,
	workaround TEXT NOT NULL,
	same_as    INTEGER REFERENCES issues(id),
	status     TEXT NOT NULL DEFAULT 'open',
	observed_via  TEXT NOT NULL DEFAULT '',
	observed_call TEXT NOT NULL DEFAULT '',
	graph_snapshot TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS issues_same_as ON issues(same_as);
CREATE INDEX IF NOT EXISTS issues_area ON issues(area);
`

var issueLaterColumns = map[string]string{
	"observed_via":  "TEXT NOT NULL DEFAULT ''",
	"observed_call":  "TEXT NOT NULL DEFAULT ''",
	"graph_snapshot": "TEXT NOT NULL DEFAULT ''",
}

func addMissingIssueColumns(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(issues)`)
	if err != nil {
		return err
	}
	present := map[string]bool{}
	for rows.Next() {
		var (
			cid, notNull, primary int
			name, typ             string
			fallback              sql.NullString
		)
		if err := rows.Scan(&cid, &name, &typ, &notNull, &fallback, &primary); err != nil {
			rows.Close()
			return err
		}
		present[name] = true
	}
	rows.Close()
	for _, name := range slices.Sorted(maps.Keys(issueLaterColumns)) {
		if present[name] {
			continue
		}
		if _, err := db.Exec(fmt.Sprintf(`ALTER TABLE issues ADD COLUMN %s %s`, name, issueLaterColumns[name])); err != nil {
			return err
		}
	}
	return nil
}

var (
	issueKinds      = []string{"bug", "missing", "docs", "usability"}
	issueSeverities = []string{"blocker", "major", "minor"}
)

func issueDBPath() string {
	return filepath.Join(DefaultOutputRoot(), "issues.sqlite")
}

func openIssueDB() (*sql.DB, error) {
	path := issueDBPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", "file:"+filepath.ToSlash(path)+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(issueSchema); err != nil {
		db.Close()
		return nil, err
	}
	if err := addMissingIssueColumns(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

type ReportIssueInput struct {
	Kind        string `json:"kind" jsonschema:"bug: something behaved wrong. missing: a node, tool or option that does not exist. docs: instructions that were wrong, stale or missing. usability: works, but cost far more turns or hard-coded numbers than it should"`
	Severity    string `json:"severity" jsonschema:"blocker: could not do it at all. major: the workaround cost quality, parametricity or many turns. minor: a small detour"`
	Area        string `json:"area" jsonschema:"the tool name, node type key or topic it is about, e.g. render_preview, sdf.SmoothUnionNode, orchestrator docs"`
	Title       string `json:"title" jsonschema:"one line naming the problem, not the build: 'no atan2 node', not 'part B tilt angle'"`
	Detail      string `json:"detail" jsonschema:"what you tried and what happened, with the real values, type keys and search queries, so a maintainer can act without asking"`
	Workaround  string `json:"workaround" jsonschema:"exactly what you did instead. 'none' only when blocked"`
	SameAs      int64  `json:"sameAs,omitempty" jsonschema:"id this tool returned for an issue you reported earlier in this build, when you hit the same problem again"`
	ObservedVia string `json:"observedVia" jsonschema:"the tool call that showed you the problem: render_preview when you saw it in an image, describe_manifest, describe_mesh or evaluate_node when you read a value, connect_nodes or create_nodes when a call refused, search_node_types when nothing was found. The arguments of your latest call to that tool are stored with the issue"`
}

type ReportIssueOutput struct {
	ID int64 `json:"id"`
}

func oneOf(field, value string, allowed []string) error {
	for _, a := range allowed {
		if value == a {
			return nil
		}
	}
	return fmt.Errorf("%s must be one of %s", field, strings.Join(allowed, ", "))
}

func (s *Server) reportIssue(ctx context.Context, req *mcpsdk.CallToolRequest, in ReportIssueInput) (*mcpsdk.CallToolResult, ReportIssueOutput, error) {
	var out ReportIssueOutput

	in.Area = strings.TrimSpace(in.Area)
	in.Title = strings.TrimSpace(in.Title)
	in.Detail = strings.TrimSpace(in.Detail)
	in.Workaround = strings.TrimSpace(in.Workaround)
	in.ObservedVia = strings.TrimSpace(in.ObservedVia)
	if err := oneOf("kind", in.Kind, issueKinds); err != nil {
		return nil, out, err
	}
	if err := oneOf("severity", in.Severity, issueSeverities); err != nil {
		return nil, out, err
	}
	for field, value := range map[string]string{"area": in.Area, "title": in.Title, "detail": in.Detail, "workaround": in.Workaround, "observedVia": in.ObservedVia} {
		if value == "" {
			return nil, out, fmt.Errorf("%s is required", field)
		}
	}

	s.mu.Lock()
	project := s.projectDir
	observedCall := s.lastCalls[in.ObservedVia]
	snapshot, snapshotErr := s.graph.EncodeToAppSchema()
	s.mu.Unlock()

	db, err := openIssueDB()
	if err != nil {
		return nil, out, err
	}
	defer db.Close()

	var sameAs sql.NullInt64
	if in.SameAs != 0 {
		root, err := issueRoot(ctx, db, in.SameAs)
		if err != nil {
			return nil, out, err
		}
		sameAs = sql.NullInt64{Int64: root, Valid: true}
	}

	res, err := db.ExecContext(ctx,
		`INSERT INTO issues (created_at, project, kind, severity, area, title, detail, workaround, same_as, observed_via, observed_call) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		time.Now().UTC().Format(time.RFC3339), project, in.Kind, in.Severity, in.Area, in.Title, in.Detail, in.Workaround, sameAs, in.ObservedVia, observedCall)
	if err != nil {
		return nil, out, err
	}
	if out.ID, err = res.LastInsertId(); err != nil {
		return nil, out, err
	}
	if snapshotErr == nil {
		saveIssueSnapshot(ctx, db, out.ID, snapshot)
	}
	return nil, out, nil
}

// Best effort: a snapshot that can't be written must not fail the report.
func saveIssueSnapshot(ctx context.Context, db *sql.DB, id int64, snapshot []byte) {
	dir := filepath.Join(filepath.Dir(issueDBPath()), "issue-snapshots")
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	path := filepath.Join(dir, fmt.Sprintf("%d.json", id))
	if os.WriteFile(path, snapshot, 0o644) != nil {
		return
	}
	_, _ = db.ExecContext(ctx, `UPDATE issues SET graph_snapshot = ? WHERE id = ?`, path, id)
}

func issueRoot(ctx context.Context, db *sql.DB, id int64) (int64, error) {
	var sameAs sql.NullInt64
	err := db.QueryRowContext(ctx, `SELECT same_as FROM issues WHERE id = ?`, id).Scan(&sameAs)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("no issue with id %d", id)
	}
	if err != nil {
		return 0, err
	}
	if sameAs.Valid {
		return sameAs.Int64, nil
	}
	return id, nil
}

func (s *Server) registerIssueTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "report_issue",
		Description: "Log a bug, missing node or tool, wrong docs, or usability problem the moment you hit it, with the workaround you used. Name the tool call that showed you the problem in observedVia; if it was an image, say render_preview and describe in detail exactly what you saw, separately from what you think caused it. Your own mistakes are not issues.",
	}, s.reportIssue)
}
