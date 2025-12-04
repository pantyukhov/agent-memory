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
		mcp.WithDescription("Create a new project to organize tasks (like Jira project). Projects can have a workspace directory for file operations."),
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
		mcp.WithDescription("List all projects, sorted by most recently updated."),
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
		mcp.WithDescription("Create a new task within a project. Tasks can override the project workspace directory."),
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
		mcp.WithDescription("List all tasks in a project, sorted by most recently updated."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListTasks)
}

func (s *Server) registerUpdateTask() {
	tool := mcp.NewTool("update_task",
		mcp.WithDescription("Update task details like name, description, workspace, or status."),
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
		mcp.WithDescription("Save an artifact (note, code, decision, etc.) to a task. Artifacts are stored as markdown files with timestamps."),
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
		mcp.WithDescription("List all artifacts for a task, sorted by most recent first."),
		mcp.WithString("project_id",
			mcp.Required(),
			mcp.Description("The project identifier."),
		),
		mcp.WithString("task_id",
			mcp.Required(),
			mcp.Description("The task identifier."),
		),
	)

	s.mcpServer.AddTool(tool, s.handleListArtifacts)
}

func (s *Server) registerSearchArtifacts() {
	tool := mcp.NewTool("search_artifacts",
		mcp.WithDescription("Search artifact content across all projects/tasks or within a specific project/task."),
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
		mcp.WithDescription("Read a file from the task's workspace. The read operation can be logged as an artifact for history tracking."),
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
		mcp.WithDescription("Search for text in files within the task's workspace. The search can be logged as an artifact."),
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
