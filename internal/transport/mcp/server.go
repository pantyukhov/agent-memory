package mcp

import (
	"log/slog"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"agent-memory/internal/application/service"
)

// Server wraps the MCP server with project/task/artifact/workspace tools.
type Server struct {
	mcpServer        *server.MCPServer
	taskService      *service.TaskService
	workspaceService *service.WorkspaceService
	logger           *slog.Logger
}

// NewServer creates a new MCP server with all tools.
func NewServer(taskService *service.TaskService, workspaceService *service.WorkspaceService, logger *slog.Logger) *Server {
	mcpServer := server.NewMCPServer(
		"agent-memory",
		"1.0.0",
		server.WithLogging(),
	)

	s := &Server{
		mcpServer:        mcpServer,
		taskService:      taskService,
		workspaceService: workspaceService,
		logger:           logger,
	}

	s.registerTools()

	return s
}

// ServeStdio starts the MCP server using stdio transport.
func (s *Server) ServeStdio() error {
	return server.ServeStdio(s.mcpServer)
}

// registerTools registers all tools with the MCP server.
func (s *Server) registerTools() {
	// Project management
	s.registerCreateProject()
	s.registerGetProject()
	s.registerListProjects()
	s.registerUpdateProject()
	s.registerDeleteProject()

	// Task management
	s.registerCreateTask()
	s.registerGetTask()
	s.registerListTasks()
	s.registerListAllTasks()
	s.registerUpdateTask()
	s.registerDeleteTask()

	// Artifact management
	s.registerSaveArtifact()
	s.registerGetArtifact()
	s.registerListArtifacts()
	s.registerSearchArtifacts()
	s.registerDeleteArtifact()

	// Workspace/File operations
	s.registerReadFile()
	s.registerListFiles()
	s.registerSearchFiles()
}

// Project tool registrations

func (s *Server) registerCreateProject() {
	tool := mcp.NewTool("create_project",
		mcp.WithDescription(`Create a new project to organize tasks (like Jira project). Projects can have a workspace directory for file operations.

WORKFLOW GUIDANCE:
- Create a project for each repository/codebase you work with
- Set workspace_path to the repository root for file operations
- Use meaningful IDs like 'myapp-backend' or 'frontend-v2'
- Projects persist between sessions - reuse existing ones`),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Unique project identifier, like 'myapp' or 'backend-api'. Will be normalized to lowercase with dashes."),
		),
		mcp.WithString("name",
			mcp.Description("Human-readable name for the project. Defaults to the ID if not provided."),
		),
		mcp.WithString("description",
			mcp.Description("Detailed description of what this project is about."),
		),
		mcp.WithString("workspace_path",
			mcp.Description("Absolute path to the workspace/repository root directory for file operations."),
		),
		mcp.WithObject("metadata",
			mcp.Description("Additional key-value metadata for the project."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleCreateProject)
}

func (s *Server) registerGetProject() {
	tool := mcp.NewTool("get_project",
		mcp.WithDescription("Get details about a specific project including workspace path and task count."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleGetProject)
}

func (s *Server) registerListProjects() {
	tool := mcp.NewTool("list_projects",
		mcp.WithDescription(`List all projects, sorted by most recently updated.

IMPORTANT: Always call this first when starting a new session to see existing projects and continue previous work.

PAGINATION: Use limit/offset for large project lists. Response includes total count and has_more flag.`),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of projects to return (default: 50)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of projects to skip for pagination (default: 0)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListProjects)
}

func (s *Server) registerUpdateProject() {
	tool := mcp.NewTool("update_project",
		mcp.WithDescription("Update project details like name, description, or workspace."),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("name",
			mcp.Description("New name for the project."),
		),
		mcp.WithString("description",
			mcp.Description("New description for the project."),
		),
		mcp.WithString("workspace_path",
			mcp.Description("New workspace path for file operations."),
		),
		mcp.WithObject("metadata",
			mcp.Description("New metadata for the project."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleUpdateProject)
}

func (s *Server) registerDeleteProject() {
	tool := mcp.NewTool("delete_project",
		mcp.WithDescription("Delete a project and ALL its tasks and artifacts. This action is irreversible!"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleDeleteProject)
}

// Task tool registrations

func (s *Server) registerCreateTask() {
	tool := mcp.NewTool("create_task",
		mcp.WithDescription(`Create a new task within a project. Tasks represent a unit of work (feature, bug, investigation).

WORKFLOW GUIDANCE:
- Create one task per user request/feature/bug
- Use descriptive IDs: 'add-auth-middleware', 'fix-login-bug', 'investigate-perf-issue'
- Task persists all work artifacts - use it to restore context later
- When resuming work, list_artifacts to see what was done before`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier to create the task in."),
		),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Unique task identifier within the project, like 'feature-login' or 'bug-123'."),
		),
		mcp.WithString("name",
			mcp.Description("Human-readable name for the task. Defaults to the ID if not provided."),
		),
		mcp.WithString("description",
			mcp.Description("Detailed description of what this task is about."),
		),
		mcp.WithString("workspace_path",
			mcp.Description("Absolute path to workspace (overrides project workspace)."),
		),
		mcp.WithObject("metadata",
			mcp.Description("Additional key-value metadata for the task."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleCreateTask)
}

func (s *Server) registerGetTask() {
	tool := mcp.NewTool("get_task",
		mcp.WithDescription("Get details about a specific task including workspace path."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleGetTask)
}

func (s *Server) registerListTasks() {
	tool := mcp.NewTool("list_tasks",
		mcp.WithDescription(`List all tasks in a project, sorted by most recently updated.

Use this to find previous work and restore context. The most recently updated task is likely the one to continue.

FILTERING: Use status parameter to filter by task status (open, in_progress, completed, archived).
PAGINATION: Use limit/offset for large task lists. Response includes total count and has_more flag.

NOTE: Task directories use kanban-style naming like [completed]-fix-login-bug for visual organization.`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("status",
			mcp.Description("Filter by status: 'open', 'in_progress', 'completed', 'archived'. Empty = all."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of tasks to return (default: 50)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of tasks to skip for pagination (default: 0)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListTasks)
}

func (s *Server) registerListAllTasks() {
	tool := mcp.NewTool("list_all_tasks",
		mcp.WithDescription(`List all tasks from ALL projects, sorted by most recently updated.

Use this to get a global view of all work across all projects. Helpful for:
- Finding recent work across multiple projects
- Getting an overview of all ongoing tasks
- Discovering tasks you may have forgotten about

FILTERING: Use status parameter to filter by task status (open, in_progress, completed, archived).
PAGINATION: Use limit/offset for large task lists. Response includes total count and has_more flag.`),
		mcp.WithString("status",
			mcp.Description("Filter by status: 'open', 'in_progress', 'completed', 'archived'. Empty = all."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of tasks to return (default: 50)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of tasks to skip for pagination (default: 0)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListAllTasks)
}

func (s *Server) registerUpdateTask() {
	tool := mcp.NewTool("update_task",
		mcp.WithDescription(`Update task details like name, description, workspace, or status.

STATUS WORKFLOW:
- 'open': Task created, not started
- 'in_progress': Actively working on it (set when you start)
- 'completed': Work finished successfully (save final summary artifact first!)
- 'archived': Old task, kept for reference

IMPORTANT: Before marking 'completed', save a summary artifact with:
- What was accomplished
- Any follow-up items
- Key files changed`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
		mcp.WithString("name",
			mcp.Description("New name for the task."),
		),
		mcp.WithString("description",
			mcp.Description("New description for the task."),
		),
		mcp.WithString("workspace_path",
			mcp.Description("New workspace path for file operations."),
		),
		mcp.WithString("status",
			mcp.Description("New status: 'open', 'in_progress', 'completed', or 'archived'."),
		),
		mcp.WithObject("metadata",
			mcp.Description("New metadata for the task."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleUpdateTask)
}

func (s *Server) registerDeleteTask() {
	tool := mcp.NewTool("delete_task",
		mcp.WithDescription("Delete a task and ALL its artifacts. This action is irreversible!"),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleDeleteTask)
}

// Artifact tool registrations

func (s *Server) registerSaveArtifact() {
	tool := mcp.NewTool("save_artifact",
		mcp.WithDescription(`Save an artifact (note, code, decision, etc.) to a task. Artifacts are stored as markdown files with timestamps.

CRITICAL FOR CONTEXT RESTORATION:
Artifacts are your memory! Save meaningful artifacts that help restore context later.

ARTIFACT TYPES AND WHEN TO USE:
- 'note': General observations, summaries, progress updates
- 'decision': Architectural choices, why you chose approach X over Y
- 'code': Important code snippets, implementations worth remembering
- 'discussion': Key points from user conversation
- 'reference': Links, docs, external resources consulted

BEST PRACTICES FOR CONTENT:
1. START with a clear title/summary line
2. INCLUDE the "why" not just the "what"
3. REFERENCE related artifacts: "Continues from artifact <id>" or "Related to decision about X"
4. END with next steps or open questions

EXAMPLE STRUCTURE:
"## Summary: Implemented JWT authentication

### Context
User requested auth middleware for API endpoints.

### What was done
- Added jwt-go dependency
- Created middleware in internal/auth/middleware.go
- Integrated with routes in cmd/api/main.go

### Key decisions
- Chose RS256 over HS256 for better security (see decision artifact <id>)
- Token expiry set to 24h based on user preference

### Next steps
- Add refresh token support
- Write tests for edge cases"

CHAINING ARTIFACTS:
When continuing work, reference previous artifacts to create a traceable chain of context.`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier to save the artifact to."),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The artifact content (markdown supported)."),
		),
		mcp.WithString("type",
			mcp.Description("Artifact type: 'note', 'code', 'decision', 'discussion', 'reference', or 'artifact' (default)."),
		),
		mcp.WithObject("metadata",
			mcp.Description("Additional metadata for the artifact."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSaveArtifact)
}

func (s *Server) registerGetArtifact() {
	tool := mcp.NewTool("get_artifact",
		mcp.WithDescription("Get a specific artifact from a task."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
		mcp.WithString("artifact_id",
			mcp.Required(),
			mcp.Description("The artifact identifier (timestamp)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleGetArtifact)
}

func (s *Server) registerListArtifacts() {
	tool := mcp.NewTool("list_artifacts",
		mcp.WithDescription(`List all artifacts for a task, sorted by most recent first.

CONTEXT RESTORATION WORKFLOW:
1. Call list_artifacts to see all work history
2. Read the MOST RECENT artifact first - it has latest state and next steps
3. If needed, read earlier artifacts to understand the full journey
4. Look for 'decision' artifacts to understand key choices made

The artifact chain tells the story of the work - use it to resume exactly where you left off.

PAGINATION: Use limit/offset for tasks with many artifacts. Response includes total count and has_more flag.`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of artifacts to return (default: 50)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of artifacts to skip for pagination (default: 0)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListArtifacts)
}

func (s *Server) registerSearchArtifacts() {
	tool := mcp.NewTool("search_artifacts",
		mcp.WithDescription(`Search artifact content across all projects/tasks or within a specific project/task.

Use this to find relevant past work:
- Search for error messages you've seen before
- Find decisions about specific technologies
- Locate code patterns you've implemented
- Discover related work across different tasks

PAGINATION: Use limit/offset for large result sets. Response includes total count and has_more flag.`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Search query to find in artifact content."),
		),
		mcp.WithString("project_id",
			mcp.Description("Optional: limit search to a specific project."),
		),
		mcp.WithString("task_id",
			mcp.Description("Optional: limit search to a specific task (requires project_id)."),
		),
		mcp.WithNumber("limit",
			mcp.Description("Maximum number of results to return (default: 50)."),
		),
		mcp.WithNumber("offset",
			mcp.Description("Number of results to skip for pagination (default: 0)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSearchArtifacts)
}

func (s *Server) registerDeleteArtifact() {
	tool := mcp.NewTool("delete_artifact",
		mcp.WithDescription("Delete an artifact from a task."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
		mcp.WithString("artifact_id",
			mcp.Required(),
			mcp.Description("The artifact identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleDeleteArtifact)
}

// Workspace/File operation registrations

func (s *Server) registerReadFile() {
	tool := mcp.NewTool("read_file",
		mcp.WithDescription(`Read a file from the task's workspace. The read operation can be logged as an artifact for history tracking.

WHY LOG FILE READS:
When log_read=true, the file content is saved as an artifact. This helps:
- Remember which files you examined
- Restore context about codebase exploration
- Track the investigation journey

Use log_read=true for important files you're analyzing, false for quick lookups.`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier. Uses task workspace or falls back to project workspace."),
		),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("Path to the file. Relative to workspace root, or absolute path."),
		),
		mcp.WithBoolean("log_read",
			mcp.Description("Whether to log this file read as an artifact (default: true)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleReadFile)
}

func (s *Server) registerListFiles() {
	tool := mcp.NewTool("list_files",
		mcp.WithDescription("List files in the task's workspace directory. Useful for exploring project structure."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier. Uses task workspace or falls back to project workspace."),
		),
		mcp.WithString("path",
			mcp.Description("Subdirectory path relative to workspace root. Empty = workspace root."),
		),
		mcp.WithString("pattern",
			mcp.Description("File pattern to filter, e.g., '*.go', '*.ts'. Empty = all files."),
		),
		mcp.WithBoolean("recursive",
			mcp.Description("List files recursively (default: false)."),
		),
		mcp.WithNumber("max_depth",
			mcp.Description("Maximum depth for recursive listing (default: 10)."),
		),
		mcp.WithBoolean("log_list",
			mcp.Description("Whether to log this listing as an artifact (default: false)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListFiles)
}

func (s *Server) registerSearchFiles() {
	tool := mcp.NewTool("search_files",
		mcp.WithDescription(`Search for text in files within the task's workspace. The search can be logged as an artifact.

WHY LOG SEARCHES:
When log_search=true, the search query and results are saved. This helps:
- Remember what you were looking for
- Track investigation patterns
- Avoid repeating the same searches

Log important searches that reveal codebase understanding.`),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier. Uses task workspace or falls back to project workspace."),
		),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("Text to search for in file contents."),
		),
		mcp.WithString("pattern",
			mcp.Description("File pattern to search in, e.g., '*.go', '*.ts'. Empty = all text files."),
		),
		mcp.WithBoolean("ignore_case",
			mcp.Description("Case-insensitive search (default: false)."),
		),
		mcp.WithNumber("max_results",
			mcp.Description("Maximum number of results to return (default: 100)."),
		),
		mcp.WithBoolean("log_search",
			mcp.Description("Whether to log this search as an artifact (default: true)."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleSearchFiles)
}
