# Working with the `polyform` MCP tools

This repo's `.mcp.json` connects a `polyform` MCP server exposing
`mcp__polyform__*` tools (`create_node`, `connect_nodes`, `render_preview`,
`start_project`, ...) for building/editing polyform node graphs. That
connection is session-wide — it's available here regardless of whether
you're a plain chat or a spawned subagent.

**Before calling any `mcp__polyform__*` tool directly, `Read`
[`.claude/agents/polyform-orchestrator.md`](.claude/agents/polyform-orchestrator.md) first.**
Tool access alone gets you none of what's in that file: the world
coordinate convention, the YAML-outline planning gate, mechanical-vs-organic
decomposition judgment, the `start_project`/autosave crash-recovery
workflow, `render_preview`'s actual capabilities and limits, and several
other standing rules earned the hard way (real bugs hit and fixed during
earlier sessions). None of it loads automatically just because the tools
are connected — it's a subagent definition, only injected into context
when spawned via the `Agent` tool with `subagent_type: polyform-orchestrator`.
A plain chat reaching for these tools directly starts with zero of that
context unless it reads the file itself.

For building or editing an actual 3D model/scene, prefer spawning the
`polyform-orchestrator` (or `polyform-part-builder` for a single
substantial part) subagent over calling the tools directly — see each
agent's `description` for when each is appropriate.

Don't put a project location in the dispatch prompt unless the user gave
one. New builds belong in `~/polyform-output/<name>` (`start_project`
with `name`); a scratchpad or temp path hides the build from the user.

**Multi-round builds use one orchestrator, continued with `SendMessage`.**
When the work is iterative — "do N iterations", review/critique rounds,
"keep refining" — spawn the orchestrator once, then deliver each round's
feedback to that same agent ID with `SendMessage`. Never spawn a fresh
orchestrator per round. A fresh spawn re-reads the full agent definition,
reloads and re-learns the graph (~250k tokens and ~12 minutes per round
measured), and forgets everything the previous round discovered — tool
limitations, workarounds, what was tried — so it re-pays for those
mistakes. Spawn a new one only if the existing agent is unrecoverable or
the user asks for a clean start.

A spawned subagent can't reach `SendUserFile`, so the orchestrator lists
its render paths under a `Renders:` heading in its final report. When that
report comes back, `SendUserFile` those paths to the user yourself, with
their captions — the render is the deliverable, not the report.
