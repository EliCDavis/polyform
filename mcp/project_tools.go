package mcp

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
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
	Path string `json:"path,omitempty" jsonschema:"directory for this project; created if it doesn't exist. Omit to auto-generate a fresh, guaranteed-unique directory under your home directory (well outside any git repository, including whatever repo this server happens to be running from) - the right choice for a genuinely new build, since it can never collide with another concurrently-running session and never leaves working files behind in a project's source tree. Pass an explicit path only to deliberately resume a specific earlier project you (or the user) already know the path to. While a project is active, the graph is autosaved here after every successful tool call, so a killed/crashed session doesn't lose work back to the last deliberate save_graph"`
}

type StartProjectOutput struct {
	Path                string `json:"path"`
	AutosavePath        string `json:"autosavePath" jsonschema:"where the graph will be continuously autosaved from now on - not meant to be loaded directly, it's overwritten after every call; see recoveredFrom for a one-time exception"`
	RecoveredFrom       string `json:"recoveredFrom,omitempty" jsonschema:"set only if an autosave already existed at this path from a previous session - it's been moved here (out of the way of future autosaves) so it's safe to load_graph and resume, instead of starting the build over"`
	RecoveredModifiedAt string `json:"recoveredModifiedAt,omitempty" jsonschema:"last-modified time (RFC3339) of the recovered autosave, so staleness can be judged before deciding to resume it"`
}

func (s *Server) startProject(ctx context.Context, req *mcpsdk.CallToolRequest, in StartProjectInput) (*mcpsdk.CallToolResult, StartProjectOutput, error) {
	var out StartProjectOutput
	var err error
	s.atomic(&err, func() error {
		path := in.Path
		if path == "" {
			path = filepath.Join(DefaultOutputRoot(), newProjectID())
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

func (s *Server) registerProjectTools() {
	mcpsdk.AddTool(s.sdk, &mcpsdk.Tool{
		Name:        "start_project",
		Description: "Start (or resume) a project directory for this build. While active, the graph is autosaved to a fixed file in this directory after every successful tool call - crash/token-exhaustion insurance beyond whatever was last saved with save_graph. If an autosave already exists here from an earlier session, it's reported back (recoveredFrom) instead of being silently overwritten - load_graph it to resume that work rather than rebuilding from scratch.",
	}, s.startProject)
}
