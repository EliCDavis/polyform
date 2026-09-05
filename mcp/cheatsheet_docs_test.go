package mcp_test

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/EliCDavis/polyform/generator"
	"github.com/EliCDavis/polyform/generator/graph"
	polyformmcp "github.com/EliCDavis/polyform/mcp"
	"github.com/stretchr/testify/require"
)

// The agent docs carry cheat-sheet tables of exact node type keys and port
// names so a build doesn't spend a search_node_types round trip per node
// type it uses. That only saves anything while the tables are right, and a
// doc that drifts is worse than no doc: one build followed a documented
// set_parameter argument that had never been implemented and burned eight
// calls plus a describe_graph finding out, and a later one followed a
// wrapper sentence that was missing the module prefix. These tests hold
// both the tables and the sentence that explains them to the registry.
const docsDir = "../.claude/agents"

var (
	// | `PATH` | inputs | outputs |
	rowRE = regexp.MustCompile("^\\|\\s*`([^`]+)`\\s*\\|([^|]*)\\|([^|]*)\\|")

	// The backticked sentence telling the reader how to expand a PATH,
	// e.g. `github.com/.../nodes.Struct[github.com/.../PATH]`.
	wrapperRE = regexp.MustCompile("`([^`]*\\bPATH\\b[^`]*)`")
)

// docSheet is one markdown file's cheat sheet: the expansion templates it
// states, and the table rows that rely on them.
type docSheet struct {
	file      string
	templates []string
	rows      map[string][2]string
}

// readDocSheets keeps each file's rows with that same file's templates. A
// pooled set would let one correct doc mask another's wrong instruction,
// which is exactly the failure being guarded against.
func readDocSheets(t *testing.T) []docSheet {
	t.Helper()

	sheets := []docSheet{}
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return err
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(body)

		rows := map[string][2]string{}
		// Only read tables introduced by the cheat-sheet header. Other
		// tables in these docs list Go type names or JSON encodings,
		// which are not node type keys.
		inCheatsheet := false
		for _, line := range strings.Split(text, "\n") {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "|") && strings.Contains(line, "PATH") {
				inCheatsheet = true
				continue
			}
			if line == "" {
				inCheatsheet = false
				continue
			}
			if !inCheatsheet {
				continue
			}

			m := rowRE.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			key := strings.TrimSpace(m[1])
			// The generic parameter.Value[T] placeholder isn't concrete.
			if !strings.Contains(key, ".") || strings.Contains(key, "[T]") {
				continue
			}
			rows[key] = [2]string{m[2], m[3]}
		}
		if len(rows) == 0 {
			return nil
		}

		templates := []string{}
		seen := map[string]bool{}
		for _, m := range wrapperRE.FindAllStringSubmatch(text, -1) {
			if !seen[m[1]] {
				seen[m[1]] = true
				templates = append(templates, m[1])
			}
		}
		require.NotEmpty(t, templates,
			"%s has cheat-sheet rows but never says how to expand a PATH into a type key", path)

		sheets = append(sheets, docSheet{file: filepath.Base(path), templates: templates, rows: rows})
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, sheets, "found no cheat-sheet tables to check")
	return sheets
}

// expand turns a table's PATH into candidate registry keys using only the
// templates stated in that same document.
func (d docSheet) expand(path string) []string {
	// Only two shapes are accepted: a key already written out in full, or
	// a PATH expanded by a wrapper the document itself states. Special
	// casing any other prefix here would let the test supply a correct
	// answer the document never gave - which is exactly how a wrapper
	// sentence missing its module prefix passed while costing a build a
	// round trip.
	const pkg = "github.com/EliCDavis/polyform/"
	if strings.HasPrefix(path, pkg) {
		return []string{path}
	}
	out := make([]string, 0, len(d.templates))
	for _, tpl := range d.templates {
		out = append(out, strings.ReplaceAll(tpl, "PATH", path))
	}
	return out
}

func registeredTypes(t *testing.T) (*graph.Instance, map[string]bool) {
	t.Helper()

	inst := graph.New(graph.Config{
		TypeFactory:     generator.Types(),
		VariableFactory: polyformmcp.NewTypedVariable,
	})
	registered := map[string]bool{}
	for _, nt := range inst.BuildSchemaForAllNodeTypes() {
		registered[nt.Type] = true
	}
	return inst, registered
}

func firstRegistered(candidates []string, registered map[string]bool) string {
	for _, c := range candidates {
		if registered[c] {
			return c
		}
	}
	return ""
}

func TestDocumentedNodeTypesExist(t *testing.T) {
	_, registered := registeredTypes(t)

	for _, sheet := range readDocSheets(t) {
		for path := range sheet.rows {
			t.Run(sheet.file+"/"+path, func(t *testing.T) {
				candidates := sheet.expand(path)
				require.NotEmpty(t, firstRegistered(candidates, registered),
					"expanding %q the way %s says to gives %v, none of which is a registered type",
					path, sheet.file, candidates)
			})
		}
	}
}

func TestDocumentedPortsExist(t *testing.T) {
	inst, registered := registeredTypes(t)

	for _, sheet := range readDocSheets(t) {
		for path, ports := range sheet.rows {
			t.Run(sheet.file+"/"+path, func(t *testing.T) {
				key := firstRegistered(sheet.expand(path), registered)
				require.NotEmpty(t, key, "no registered type for %q", path)

				node, _, err := inst.CreateNode(key)
				require.NoError(t, err)

				for i, side := range []string{"input", "output"} {
					real := map[string]bool{}
					if i == 0 {
						for name := range node.Inputs() {
							real[name] = true
						}
					} else {
						for name := range node.Outputs() {
							real[name] = true
						}
					}

					for _, name := range documentedPorts(ports[i]) {
						require.True(t, real[name],
							"%s lists %s port %q on %s, which has %v", sheet.file, side, name, path, keysOf(real))
					}
				}
			})
		}
	}
}

// documentedPorts pulls port names out of one table cell, skipping the
// trailing "..." a long list is abbreviated with and the "(none)" marker.
func documentedPorts(cell string) []string {
	cell = strings.ReplaceAll(cell, "**", "")
	out := []string{}
	for _, raw := range strings.Split(cell, ",") {
		name := strings.TrimSpace(raw)
		name = strings.TrimSuffix(name, "[]")
		name = strings.Trim(name, "`* ")
		if name == "" || name == "..." || name == "(none)" || strings.HasPrefix(name, "*(") {
			continue
		}
		out = append(out, name)
	}
	return out
}

func keysOf(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
