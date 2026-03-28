<h1 align="center">cc9s</h1>

<p align="center">
  <strong>A k9s-inspired TUI for managing Claude Code sessions, skills, and agents</strong>
</p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/kincoy/cc9s/releases"><img src="https://img.shields.io/badge/Release-v0.2.3-green.svg" alt="Release"></a>
</p>

<p align="center">
 English | <a href="README_zh.md">简体中文</a>
</p>

---

## Why cc9s?

Claude Code stores session data as JSONL files under `~/.claude/`. When you accumulate hundreds of sessions across dozens of projects, finding and managing them becomes painful.

cc9s solves this by providing a full-screen terminal UI — inspired by [k9s](https://github.com/derailed/k9s) — that lets you browse, search, inspect, and resume sessions, and inspect local Claude Code skills and agents, without leaving your keyboard.

## Demo

Watch a terminal recording here: [asciinema demo](https://asciinema.org/a/vABD89zAYT8G7Y6C)

One common flow:

1. Press `:` and run `sessions`
2. Press `/` to start real-time search
3. Type `active` to find active sessions, or `stale` to inspect unreliable ones
4. Press `d` to open details for the selected session

## Features

- **Two-level navigation** — Projects → Sessions, drill down with `Enter`, go back with `Esc`
- **Project overview** — See local project session, skill, command, and agent summaries, then inspect project paths and local Claude roots with `d`
- **Session resume** — Jump directly into a Claude Code session from the TUI
- **Search & filter** — `/` to search, `:context <name>` to filter by project
- **Batch delete** — `Space` to select, `Ctrl+D` to delete multiple sessions
- **Session details** — View session stats, summary, and tool call logs
- **Skill resource browser** — View available Claude Code skills and commands from project, user, and plugin scopes
- **Agent resource browser** — View file-backed Claude Code agents from project, user, and plugin scopes with Ready / Invalid states
- **Tab completion** — Auto-complete commands and project names
- **Fully keyboard-driven** — No mouse required
- **Built-in themes** — 4 color presets (`default`, `dark-solid`, `high-contrast`, `gruvbox`), switchable via `--theme` flag or `CC9S_THEME` env
- **CLI mode** — Read-only command suite for shell scripts and automation (`cc9s status`, `cc9s projects list`, `cc9s sessions list`, etc.)
- **JSON output** — Structured JSON output via `--json` flag for AI agents and tooling

## Screenshots

**Project list** — Browse all Claude Code projects

<p align="center">
  <img src="assets/projects.jpg" alt="Project list" width="720">
</p>

**Search** — Press `/` to search sessions in real-time

<p align="center">
  <img src="assets/search.jpg" alt="Search" width="720">
</p>

**All sessions** — View sessions across all projects

<p align="center">
  <img src="assets/session-all.jpg" alt="All sessions" width="720">
</p>

**Project sessions** — Filter sessions by project context

<p align="center">
  <img src="assets/session-specify.jpg" alt="Project sessions" width="720">
</p>

**Session details** — Press `d` to view stats, summary, and tool usage

<p align="center">
  <img src="assets/d-description.jpg" alt="Session details" width="720">
</p>

**Session log** — Press `l` to browse conversation turns

<p align="center">
  <img src="assets/l-logs.jpg" alt="Session log" width="720">
</p>

**Command mode** — Press `:` to enter commands with Tab completion

<p align="center">
  <img src="assets/command.jpg" alt="Command mode" width="720">
</p>

**Skills** — `:skills` to browse available skills and commands

<p align="center">
  <img src="assets/skills.jpg" alt="Skills" width="720">
</p>

**Agents** — `:agents` to browse available file-backed agents

<p align="center">
  <img src="assets/agents.jpg" alt="Agents" width="720">
</p>

## Quick Start

### Prerequisites

- Go 1.25+
- Claude Code installed (sessions are read from `~/.claude/`)
- macOS / Linux (terminal with true color support recommended)

### Install

**Homebrew (macOS / Linux):**

```bash
brew tap kincoy/tap
brew install cc9s
```

**Download latest release with `curl`:**

```bash
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "unsupported arch: $ARCH" >&2; exit 1 ;;
esac

curl -fsSL "https://github.com/kincoy/cc9s/releases/latest/download/cc9s-${OS}-${ARCH}" -o cc9s
chmod +x cc9s
sudo mv cc9s /usr/local/bin/cc9s
```

**Go install:**

```bash
go install github.com/kincoy/cc9s@latest
```

**Build from source:**

```bash
git clone https://github.com/kincoy/cc9s.git
cd cc9s
go build -o cc9s .
```

### Run

```bash
cc9s
```

On first launch, cc9s scans `~/.claude/projects/` for projects and sessions, then discovers available resource pages from project roots, user roots, and installed plugins. Today that includes `skills`, `commands`, and file-backed `agents`. This may take a moment if you have many local resources.

## CLI

cc9s also ships with a read-only CLI for shell workflows, automation, and AI agents. Running `cc9s` with no arguments still launches the TUI; adding arguments switches to CLI mode.

### Start Here

Want to know whether your Claude Code environment looks healthy in one command?

```text
cc9s status

Example output:

Claude Code Environment

  Projects:   12
  Sessions:   148
  Resources:  39
  Total Size: 82.4 MB

Lifecycle
  Active:    2
  Idle:      9
  Completed: 121
  Stale:     16

Issues
  ! stale sessions (16) [11%]
    Run: cc9s sessions cleanup --dry-run
  ! invalid skills (1)
    Run: cc9s skills list --json

Top Projects
  alpha-service   42 sessions (1 active)  18.6 MB
  docs-site       31 sessions (0 active)   7.2 MB
  infra-tooling   27 sessions (1 active)  23.5 MB
  api-gateway     24 sessions (0 active)  11.4 MB
  playground      12 sessions (0 active)   4.1 MB
```

Need the same snapshot for tooling or an AI agent?

```bash
cc9s status --json
```

### Smart Cleanup

`cc9s sessions cleanup --dry-run` now scores session value and groups cleanup suggestions into actionable recommendation tiers:

```text
cc9s sessions cleanup --dry-run

Session Cleanup Preview (dry-run — no data was modified)

  Filters:  state=stale

Summary
  Matched:  16 sessions across 5 projects (4.2 MB)

Recommendations
  Delete:   12 sessions (safe to remove)
  Review:   3 sessions (check before deleting)
  Keep:     1 sessions (valuable content)
```

Each session is assessed from conversation depth, tool usage, token investment, and content volume. Use `--json` for full recommendation, score, and reason details.

### Full Help

The CLI surface is best viewed directly from the binary:

```text
cc9s -h

cc9s — Claude Code session manager

Usage:
  cc9s                      Launch TUI (default, no arguments)
  cc9s status               Environment health overview
  cc9s projects list        List all projects
  cc9s projects inspect <name>  Project details (match by name or path)
  cc9s sessions list        List sessions across all projects
  cc9s sessions inspect <id>   Session details (exact ID from list output)
  cc9s sessions cleanup --dry-run  Preview smart cleanup recommendations (read-only)
  cc9s skills list          List skills and commands
  cc9s agents list          List agents
  cc9s agents inspect <name>   Agent details (match by name or path)
  cc9s version              Print version
  cc9s themes               List available built-in themes
  cc9s help                 Print this help

Short flags:
  -h, --help                Show help
  -v, --version             Print version
  --theme <name>            Apply theme at startup (default, dark-solid, high-contrast, gruvbox)
  CC9S_THEME env            Same as --theme but set via environment variable

Commands and flags:
  status                   (no extra flags)
  projects list            --limit <n>  --sort <field>  --json
  projects inspect <name>  --json
  sessions list            --project <name>  --state <state>  --limit <n>  --sort <field>  --json
  sessions inspect <id>    --json
  sessions cleanup         --dry-run  --project <name>  --state <state>  --older-than <dur>  --json
  skills list              --project <name>  --scope <scope>  --type <type>  --json
  agents list              --project <name>  --scope <scope>  --json
  agents inspect <name>    --json

  --json is supported on all commands. Default output is human-readable text.

Enumerations:
  --state <state>          Active, Idle, Completed, Stale (case-insensitive partial match)
  --scope <scope>          User, Project, Plugin (case-insensitive partial match)
  --type <type>            Skill, Command (case-insensitive partial match)
  --sort <field>           projects: name, sessions | sessions: updated, state, project
  --older-than <dur>       Duration, e.g. 72h, 7d, 168h, 30m

Resource aliases:
  projects | project | proj
  sessions | session | ss
  skills   | skill   | sk
  agents   | agent   | ag

Output:
  list commands           -> JSON array of objects
  status / inspect / cleanup -> JSON single object
  errors                  -> {"error":"<message>"}
  All timestamps are RFC 3339. Paths are absolute.

Common patterns:
  cc9s status                              Quick environment health check
  cc9s status --json                        Machine-readable overview
  cc9s sessions list --state active --json  Find active sessions, get full IDs
  cc9s sessions inspect <id> --json         Full session details (model, tokens, lifecycle)
  cc9s sessions cleanup --dry-run           Preview smart cleanup recommendations
  cc9s projects inspect cc9s               Inspect a specific project
  cc9s skills list --project cc9s --json    Skills for one project
```

## Key Bindings

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Jump to top |
| `G` | Jump to bottom |
| `Enter` | Select / Drill down |
| `Esc` | Go back / Cancel |
| `q` | Quit |

### Actions

| Key | Action |
|-----|--------|
| `/` | Search current resource |
| `s` | Cycle sort field |
| `S` | Reverse sort order |
| `d` | View project, session, skill, or agent details |
| `e` | Edit selected skill, command, or agent file |
| `Space` | Toggle select session |
| `Ctrl+D` | Delete selected session(s) |
| `l` | View session log |
| `0` | Switch to "all projects" context |
| `?` | Help panel |

### Command Mode

Type `:` to enter command mode. Press `Tab` to autocomplete.

| Command | Description |
|---------|-------------|
| `:skills` | Show available skills and commands |
| `:agents` | Show available file-backed agents |
| `:sessions` | Show sessions across projects |
| `:projects` | Show projects |
| `:context all` | Switch current resource to all-project context |
| `:context <name>` | Filter current resource by project context |
| `:cleanup` | Toggle the RECOMMEND cleanup column in sessions view |
| `:q` | Quit |

## How It Works

```
~/.claude/
├── projects/
│   ├── <encoded-project-path>/
│   │   ├── *.jsonl          # Session data (conversation history)
│   │   └── sessions/        # Active session markers
│   │       └── <pid>.json
│   └── ...
├── skills/                   # User-level local skills
├── commands/                 # User-level local commands
├── agents/                   # User-level local agents
├── plugins/                  # Installed plugin cache and metadata
└── sessions/                 # Global active session index
```

cc9s reads JSONL files from `~/.claude/projects/`, then discovers available resource pages from project `.claude/skills`, `.claude/commands`, and `.claude/agents`, user `~/.claude/skills`, `~/.claude/commands`, and `~/.claude/agents`, plus installed plugin resources. Agent availability is reconciled against `claude agents`, and built-in agents are intentionally excluded from the v1 agent resource. cc9s does **not** modify Claude Code session data — deletion operations still require explicit confirmation.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'feat: add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Acknowledgements

- [k9s](https://github.com/derailed/k9s) — The design inspiration for cc9s's keyboard-driven TUI experience
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) / [Lip Gloss](https://github.com/charmbracelet/lipgloss) — The excellent Go TUI framework

## License

[MIT](LICENSE)
