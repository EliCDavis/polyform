package mcp

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// autosaveFileName is the fixed name autosaveLocked writes to within a
// project's directory. Fixed (not per-graph/timestamped) so there's always
// exactly one obvious place to look for recoverable state, and so repeated
// autosaves overwrite in place rather than accumulating.
const autosaveFileName = "autosave.json"

// DefaultOutputRoot is where polyform-mcp writes its own working files by
// default (an auto-generated project directory, the call log) -
// deliberately outside any git repository (the user's home directory, not
// the current working directory, which for this process is usually
// whatever repo it was launched from) so a build's renders/autosave/logs
// never land in a git working tree, never need a .gitignore entry, and
// never show up as untracked clutter no matter what a build happens to
// name its own subfolders. Falls back to the OS temp directory if the
// home directory can't be resolved.
func DefaultOutputRoot() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, "polyform-output")
	}
	return filepath.Join(os.TempDir(), "polyform-output")
}

// newProjectID returns a v4 UUID string drawn from crypto/rand. Two
// concurrent sessions (e.g. two chats each running their own polyform-mcp
// process) that both omit Path land in sibling directories instead of
// racing on the same one - the actual fix for that class of conflict,
// since coordinating on a shared conventional name never fully rules out
// a collision. A nanosecond-clock-seeded PRNG was tried first and
// rejected: two calls close enough together can land in the same clock
// tick on this platform (observed directly - two start_project calls
// milliseconds apart produced an identical seed and thus an identical
// "unique" id), which is exactly the collision this exists to prevent.
// crypto/rand has no such resolution ceiling.
func newProjectID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is essentially unheard of on any real OS;
		// fall back to something still unique enough to avoid a same-
		// process collision rather than erroring the whole tool call.
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10xx
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

type StartProjectInput struct {
	Name        string `json:"name,omitempty" jsonschema:"for a new build: a short folder name such as 'my-model'. The project is created under the projects folder (~/polyform-output) as that name, with -2, -3... appended if it is taken. Omit for a generated unique name"`
	Path        string `json:"path,omitempty" jsonschema:"only to resume an existing project, or when the user named a location. Never pick a location yourself (a scratchpad, a temp folder, the repo, the current directory): new builds always go under the projects folder via name"`
	OutsideRoot bool   `json:"outsideRoot,omitempty" jsonschema:"set true only when the user asked for the project at a path outside the projects folder; a new path outside it is refused otherwise"`
}

type StartProjectOutput struct {
	Path                string `json:"path"`
	AutosavePath        string `json:"autosavePath" jsonschema:"where the graph will be continuously autosaved from now on - not meant to be loaded directly, it's overwritten after every call; see recoveredFrom for a one-time exception"`
	RecoveredFrom       string `json:"recoveredFrom,omitempty" jsonschema:"set whenever an autosave already existed at this path - it's been moved here, out of the way of future autosaves. This is normal on a deliberate resume of a cleanly saved project, not a sign of a crash: if you already have a saved graph.json, load that; load this only when the last save may have been lost"`
	RecoveredModifiedAt string `json:"recoveredModifiedAt,omitempty" jsonschema:"last-modified time (RFC3339) of the recovered autosave, so staleness can be judged before deciding to resume it"`
}

func (s *Server) startProject(ctx context.Context, req *mcpsdk.CallToolRequest, in StartProjectInput) (*mcpsdk.CallToolResult, StartProjectOutput, error) {
	var out StartProjectOutput
	var err error
	s.atomic(&err, func() error {
		path, e := projectPath(DefaultOutputRoot(), in)
		if e != nil {
			return e
		}
		if e := os.MkdirAll(path, 0o755); e != nil {
			return e
		}

		autosavePath := filepath.Join(path, autosaveFileName)
		if info, statErr := os.Stat(autosavePath); statErr == nil {
			recoveredPath := filepath.Join(path, fmt.Sprintf("autosave-recovered-%s.json", time.Now().UTC().Format("20060102-150405")))
			if e := os.Rename(autosavePath, recoveredPath); e != nil {
				return e
			}
			out.RecoveredFrom = recoveredPath
			out.RecoveredModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
		}

		s.projectDir = path
		out.Path = path
		out.AutosavePath = autosavePath
		return nil
	})
	return nil, out, err
}

func projectPath(root string, in StartProjectInput) (string, error) {
	if in.Path != "" && in.Name != "" {
		return "", fmt.Errorf("pass name or path, not both")
	}

	if in.Path == "" {
		if in.Name == "" {
			return filepath.Join(root, newProjectID()), nil
		}
		base := filepath.Base(filepath.Clean(in.Name))
		if base != in.Name || base == "." || base == ".." {
			return "", fmt.Errorf("name %q is not a single folder name", in.Name)
		}
		if e := os.MkdirAll(root, 0o755); e != nil {
			return "", e
		}
		path := filepath.Join(root, base)
		for i := 2; ; i++ {
			// Mkdir rather than a stat check, so two sessions can't claim one name.
			e := os.Mkdir(path, 0o755)
			if e == nil {
				return path, nil
			}
			if !os.IsExist(e) {
				return "", e
			}
			path = filepath.Join(root, fmt.Sprintf("%s-%d", base, i))
		}
	}

	path, e := filepath.Abs(in.Path)
	if e != nil {
		return "", e
	}
	if in.OutsideRoot || isProject(path) {
		return path, nil
	}
	if rel, e := filepath.Rel(root, path); e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return path, nil
	}
	return "", fmt.Errorf("path %q is outside the projects folder %q", path, root)
}

func isProject(dir string) bool {
	for _, file := range []string{"graph.json", autosaveFileName} {
		if _, err := os.Stat(filepath.Join(dir, file)); err == nil {
			return true
		}
	}
	return false
}

func (s *Server) registerProjectTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "start_project",
		Description: "Start (or resume) a project directory for this build. Every new build lives in the projects folder (~/polyform-output): pass name for a readable folder, and pass path only to resume an existing project or when the user named a location. While active, the graph is autosaved to a fixed file in this directory after every successful tool call - crash/token-exhaustion insurance beyond whatever was last saved with save_graph. If an autosave already exists here from an earlier session, it's moved aside and reported back (recoveredFrom) instead of being silently overwritten. That happens on every resume, clean or not; load_graph your saved graph.json if you have one, and fall back to the recovered autosave only if the last save was lost.",
	}, s.startProject)
}
