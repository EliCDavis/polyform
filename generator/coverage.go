package generator

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/EliCDavis/polyform/generator/graph"
	"github.com/EliCDavis/polyform/nodes"
	"github.com/EliCDavis/polyform/refutil"
)

type CoverageReport struct {
	Graphs []string
	Usage  map[string][]string
	Paths  map[string][]string
	names  map[string]string
}

func NodeCoverage(types *refutil.TypeFactory, root string) (CoverageReport, error) {
	report := CoverageReport{
		Usage: make(map[string][]string),
		Paths: make(map[string][]string),
		names: make(map[string]string),
	}

	for _, key := range types.Types() {
		node, ok := types.New(key).(nodes.Node)
		if !ok {
			continue
		}
		typeSchema := graph.BuildNodeTypeSchema(key, node)
		report.Paths[typeSchema.Path] = append(report.Paths[typeSchema.Path], key)
		report.names[key] = typeSchema.DisplayName
		report.Usage[key] = nil
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".json" || d.Name() == "autosave.json" {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var decoded any
		if json.Unmarshal(data, &decoded) != nil {
			return nil
		}

		found := make(map[string]bool)
		collectTypes(decoded, func(t string) {
			if _, registered := report.Usage[t]; registered {
				found[t] = true
			}
		})
		if len(found) == 0 {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		report.Graphs = append(report.Graphs, rel)
		for t := range found {
			report.Usage[t] = append(report.Usage[t], rel)
		}
		return nil
	})
	return report, err
}

func collectTypes(v any, visit func(string)) {
	switch value := v.(type) {
	case map[string]any:
		for key, child := range value {
			if s, ok := child.(string); ok && key == "type" {
				visit(s)
				continue
			}
			collectTypes(child, visit)
		}
	case []any:
		for _, child := range value {
			collectTypes(child, visit)
		}
	}
}

type CoverageAddition struct {
	Graph string
	Adds  int
}

func (r CoverageReport) Used() map[string]bool {
	used := make(map[string]bool)
	for key, graphs := range r.Usage {
		if len(graphs) > 0 {
			used[key] = true
		}
	}
	return used
}

func (r CoverageReport) RankAdditions(covered map[string]bool, limit int) []CoverageAddition {
	perGraph := make(map[string][]string)
	for key, graphs := range r.Usage {
		for _, g := range graphs {
			perGraph[g] = append(perGraph[g], key)
		}
	}

	seen := make(map[string]bool, len(covered))
	for key := range covered {
		seen[key] = true
	}

	var picks []CoverageAddition
	for len(picks) < limit {
		best, bestAdds := "", 0
		for _, g := range r.Graphs {
			adds := 0
			for _, key := range perGraph[g] {
				if !seen[key] {
					adds++
				}
			}
			if adds > bestAdds {
				best, bestAdds = g, adds
			}
		}
		if bestAdds == 0 {
			break
		}
		picks = append(picks, CoverageAddition{Graph: best, Adds: bestAdds})
		for _, key := range perGraph[best] {
			seen[key] = true
		}
	}
	return picks
}

func WriteAdditionsMarkdown(out io.Writer, base int, total int, picks []CoverageAddition) error {
	var b strings.Builder
	b.WriteString("## Best additions\n\n| Graph | Adds | Covered |\n| --- | --- | --- |\n")
	covered := base
	for _, p := range picks {
		covered += p.Adds
		fmt.Fprintf(&b, "| %s | %d | %d of %d |\n", p.Graph, p.Adds, covered, total)
	}
	_, err := io.WriteString(out, b.String())
	return err
}

func (r CoverageReport) WriteMarkdown(out io.Writer) error {
	used := len(r.Used())

	var b strings.Builder
	fmt.Fprintf(&b, "# Node Coverage\n\n%d of %d node types used across %d graphs.\n\n", used, len(r.Usage), len(r.Graphs))
	b.WriteString("| Package | Used | Total |\n| --- | --- | --- |\n")

	paths := make([]string, 0, len(r.Paths))
	for path := range r.Paths {
		paths = append(paths, path)
	}
	slices.Sort(paths)

	for _, path := range paths {
		pathUsed := 0
		for _, key := range r.Paths[path] {
			if len(r.Usage[key]) > 0 {
				pathUsed++
			}
		}
		fmt.Fprintf(&b, "| %s | %d | %d |\n", path, pathUsed, len(r.Paths[path]))
	}

	b.WriteString("\n## Unused\n")
	for _, path := range paths {
		var unused []string
		for _, key := range r.Paths[path] {
			if len(r.Usage[key]) == 0 {
				unused = append(unused, r.names[key])
			}
		}
		if len(unused) == 0 {
			continue
		}
		slices.Sort(unused)
		fmt.Fprintf(&b, "\n### %s\n\n", path)
		for _, name := range unused {
			fmt.Fprintf(&b, "- %s\n", name)
		}
	}

	_, err := io.WriteString(out, b.String())
	return err
}
