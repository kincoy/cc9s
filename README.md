<h1 align="center">cc9s</h1>

<p align="center">
  <strong>A k9s-inspired TUI for managing Claude Code sessions</strong>
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

cc9s solves this by providing a full-screen terminal UI — inspired by [k9s](https://github.com/derailed/k9s) — that lets you browse, search, inspect, and resume sessions without leaving your keyboard.

## Features

- **Two-level navigation** — Projects → Sessions, drill down with `Enter`, go back with `Esc`
- **Session resume** — Jump directly into a Claude Code session from the TUI
- **Search & filter** — `/` to search, `:context <name>` to filter by project
- **Batch delete** — `Space` to select, `Ctrl+D` to delete multiple sessions
- **Session details** — View session stats, summary, and tool call logs
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

On first launch, cc9s scans `~/.claude/projects/` for projects and sessions. This may take a moment if you have many sessions.

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
| `/` | Search sessions |
| `s` | Toggle sort order |
| `d` | View session details |
| `Space` | Toggle select session |
| `Ctrl+D` | Delete selected session(s) |
| `l` | View session log |
| `0` | Switch to "all projects" context |
| `?` | Help panel |

### Command Mode

Type `:` to enter command mode. Press `Tab` to autocomplete.

| Command | Description |
|---------|-------------|
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
└── sessions/                 # Global active session index
```

cc9s reads JSONL files from `~/.claude/projects/` and presents them in a structured TUI. It does **not** modify any Claude Code data — deletion operations require explicit confirmation.

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
