<h1 align="center">cc9s</h1>

<p align="center">
  <strong>A k9s-inspired TUI for managing Claude Code sessions and skills</strong>
</p>

<p align="center">
  <a href="https://go.dev"><img src="https://img.shields.io/badge/Go-1.25%2B-00ADD8?logo=go" alt="Go version"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue.svg" alt="License"></a>
  <a href="https://github.com/kincoy/cc9s/releases"><img src="https://img.shields.io/badge/Release-v0.1.0-green.svg" alt="Release"></a>
</p>

<p align="center">
 English | <a href="README_zh.md">简体中文</a>
</p>

---

## Why cc9s?

Claude Code stores session data as JSONL files under `~/.claude/`. When you accumulate hundreds of sessions across dozens of projects, finding and managing them becomes painful.

cc9s solves this by providing a full-screen terminal UI — inspired by [k9s](https://github.com/derailed/k9s) — that lets you browse, search, inspect, and resume sessions, and inspect local Claude Code skills, without leaving your keyboard.

## Demo

Watch a terminal recording here: [asciinema demo](https://asciinema.org/a/vABD89zAYT8G7Y6C)

One common flow:

1. Press `:` and run `sessions`
2. Press `/` to start real-time search
3. Type `active` to find active sessions, or `stale` to inspect unreliable ones
4. Press `d` to open details for the selected session

## Features

- **Two-level navigation** — Projects → Sessions, drill down with `Enter`, go back with `Esc`
- **Session resume** — Jump directly into a Claude Code session from the TUI
- **Search & filter** — `/` to search, `:context <name>` to filter by project
- **Batch delete** — `Space` to select, `Ctrl+D` to delete multiple sessions
- **Session details** — View session stats, summary, and tool call logs
- **Skill resource browser** — View available Claude Code skills and commands from project, user, and plugin scopes
- **Tab completion** — Auto-complete commands and project names
- **Fully keyboard-driven** — No mouse required

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

On first launch, cc9s scans `~/.claude/projects/` for projects and sessions, then discovers available skill resources from project roots, user roots, and installed plugins, including both `skills` and `commands`. This may take a moment if you have many local resources.

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
| `s` | Toggle sort order |
| `d` | View session or skill details |
| `e` | Edit selected skill or command |
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
| `:sessions` | Show sessions across projects |
| `:projects` | Show projects |
| `:context all` | Show sessions from all projects |
| `:context <name>` | Filter sessions by project name |
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
├── plugins/                  # Installed plugin cache and metadata
└── sessions/                 # Global active session index
```

cc9s reads JSONL files from `~/.claude/projects/`, then discovers available skill resources from project `.claude/skills` and `.claude/commands`, user `~/.claude/skills` and `~/.claude/commands`, plus installed plugin resources. It does **not** modify Claude Code session data — deletion operations still require explicit confirmation.

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
