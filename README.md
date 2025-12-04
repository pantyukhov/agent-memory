# Agent Memory

An MCP (Model Context Protocol) server that provides persistent memory and context management for AI agents. It enables agents to maintain context, track progress, and resume work across multiple sessions.

## Features

- **Project Management** - Organize work into separate projects with workspace paths
- **Task Tracking** - Create and manage tasks (features, bugs, investigations) with status tracking
- **Artifact Storage** - Save timestamped work logs including notes, code snippets, decisions, and references
- **Workspace Operations** - Read files, list directories, and search content within project workspaces
- **Context Restoration** - Retrieve previous work and artifacts across sessions
- **Full-text Search** - Search across artifacts and workspace files

## Installation

### Download Pre-built Binary

Download the latest release for your platform from [GitHub Releases](https://github.com/pantyukhov/agent-memory/releases):

| Platform | Architecture | File |
|----------|--------------|------|
| Linux | x64 | `agent-memory-linux-amd64.tar.gz` |
| Linux | ARM64 | `agent-memory-linux-arm64.tar.gz` |
| macOS | Intel | `agent-memory-darwin-amd64.tar.gz` |
| macOS | Apple Silicon | `agent-memory-darwin-arm64.tar.gz` |
| Windows | x64 | `agent-memory-windows-amd64.zip` |

```bash
# Example for macOS Apple Silicon
curl -LO https://github.com/pantyukhov/agent-memory/releases/latest/download/agent-memory-darwin-arm64.tar.gz
tar -xzf agent-memory-darwin-arm64.tar.gz
chmod +x agent-memory

# Move to a directory in your PATH
sudo mv agent-memory /usr/local/bin/
```

### Build from Source

Prerequisites: Go 1.24.0 or later

```bash
git clone https://github.com/pantyukhov/agent-memory.git
cd agent-memory

make build
```

## Usage

### Running the Server

```bash
# Run with default storage (~/.agent-memory/tasks)
./build/agent-memory

# Run with custom storage path
./build/agent-memory -tasks-path /path/to/storage

# Set log level (debug, info, warn, error)
./build/agent-memory -log-level debug
```

### MCP Client Configuration

#### Claude Desktop

Add to `~/Library/Application Support/Claude/claude_desktop_config.json` (macOS) or `%APPDATA%\Claude\claude_desktop_config.json` (Windows):

```json
{
  "mcpServers": {
    "agent-memory": {
      "command": "/usr/local/bin/agent-memory",
      "args": ["-tasks-path", "~/.agent-memory/tasks"]
    }
  }
}
```

#### Claude Code

Add to your Claude Code MCP settings:

```json
{
  "mcpServers": {
    "agent-memory": {
      "command": "/usr/local/bin/agent-memory",
      "args": ["-tasks-path", "~/.agent-memory/tasks"]
    }
  }
}
```

#### Cursor

Add to Cursor's MCP configuration:

```json
{
  "mcpServers": {
    "agent-memory": {
      "command": "/usr/local/bin/agent-memory",
      "args": []
    }
  }
}
```

### Configure Your Project's CLAUDE.md

Add this to your project's `CLAUDE.md` to enable automatic context persistence:

```markdown
## Agent Memory

Save task context to agent-memory MCP for session persistence.

**Start:** `list_projects()` → `list_tasks(project_id)` → `list_artifacts(project_id, task_id)`

**Work:** `create_project(id, workspace_path)` → `create_task(project_id, id)` → `save_artifact(project_id, task_id, content, type)`

**Types:** note, code, decision, reference

Save progress frequently: findings, decisions, blockers, patterns.
```

## MCP Tools

### Project Management

| Tool | Description |
|------|-------------|
| `create_project` | Create a new project with workspace path |
| `get_project` | Retrieve project details |
| `list_projects` | List all projects |
| `update_project` | Modify project settings |
| `delete_project` | Delete project and all tasks |

### Task Management

| Tool | Description |
|------|-------------|
| `create_task` | Create a task within a project |
| `get_task` | Retrieve task details |
| `list_tasks` | List project tasks |
| `list_all_tasks` | List tasks from ALL projects |
| `update_task` | Modify task |
| `delete_task` | Delete task and all artifacts |

### Artifact Management

| Tool | Description |
|------|-------------|
| `save_artifact` | Save work artifacts |
| `get_artifact` | Retrieve a specific artifact |
| `list_artifacts` | List task artifacts |
| `search_artifacts` | Full-text search across artifacts |
| `delete_artifact` | Remove an artifact |

### Workspace Operations

| Tool | Description |
|------|-------------|
| `read_file` | Read files from workspace |
| `list_files` | List workspace directory structure |
| `search_files` | Search file contents |

## Data Model

- **Project** - Top-level organizational unit with workspace path
- **Task** - Unit of work with status (`open`, `in_progress`, `completed`, `archived`)
- **Artifact** - Timestamped record (types: `note`, `code`, `decision`, `discussion`, `reference`)

## Storage

Data is stored on the filesystem:

```text
~/.agent-memory/tasks/
  /<project-id>/
    project.json
    /<task-id>/
      task.json
      /artifacts/
        note.1234567890.md
```

## Development

```bash
make deps          # Download dependencies
make test          # Run tests
make test-coverage # Run tests with coverage
make fmt           # Format code
make lint          # Run linter
make build-all     # Build for all platforms
make clean         # Clean build artifacts
```

## Architecture

The project follows Clean Architecture:

```text
internal/
├── domain/task/           # Core entities and repository interfaces
├── application/service/   # Business logic services
├── infrastructure/storage # Filesystem repository implementation
└── transport/mcp/         # MCP protocol handlers
```

## License

MIT
