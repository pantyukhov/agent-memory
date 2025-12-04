# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Agent Memory MCP Server

This is an MCP server that provides persistent memory for AI agents. **Always use agent-memory to save task context** so you can return to work later.

## How to Use Agent-Memory

### Starting a New Task

1. **Create or select a project:**
   ```
   mcp__agent-memory__create_project(id="project-name", workspace_path="/path/to/repo")
   ```

2. **Create a task for your work:**
   ```
   mcp__agent-memory__create_task(project_id="project-name", id="task-name", description="What you're working on")
   ```

3. **Save artifacts as you work** - notes, decisions, code snippets:
   ```
   mcp__agent-memory__save_artifact(project_id="...", task_id="...", content="...", type="note|code|decision")
   ```

### Resuming Work

1. **List projects to find previous work:**
   ```
   mcp__agent-memory__list_projects()
   ```

2. **List tasks in a project:**
   ```
   mcp__agent-memory__list_tasks(project_id="...")
   ```

3. **Get task details and artifacts:**
   ```
   mcp__agent-memory__get_task(project_id="...", task_id="...")
   mcp__agent-memory__list_artifacts(project_id="...", task_id="...")
   ```

### What to Save

- **Notes:** Progress updates, findings, context
- **Decisions:** Architectural choices and reasoning
- **Code:** Important snippets, examples, patterns found
- **References:** Links to relevant files, documentation

### Best Practices

- Save context frequently - don't wait until the end
- Use descriptive task IDs (e.g., `fix-auth-bug`, `add-user-settings`)
- Update task status: `open` → `in_progress` → `completed`
- Save key decisions with reasoning for future reference

## Build Commands

```bash
make build         # Build the server
make test          # Run tests
make lint          # Run linter
make fmt           # Format code
```

## Project Structure

- `cmd/mcp/` - Main entry point
- `internal/domain/task/` - Core entities
- `internal/application/service/` - Business logic
- `internal/infrastructure/storage/` - Filesystem storage
- `internal/transport/mcp/` - MCP protocol handlers
