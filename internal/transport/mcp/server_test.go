package mcp

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"agent-memory/internal/application/service"
	"agent-memory/internal/domain/task"
	"agent-memory/internal/infrastructure/storage/filesystem"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agent-memory-mcp-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repo, err := filesystem.NewRepository(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create repository: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	taskSvc := service.NewTaskService(repo, logger)
	workspaceSvc := service.NewWorkspaceService(repo, logger)

	server := NewServer(taskSvc, workspaceSvc, logger)

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	}

	return server, cleanup
}

func createCallToolRequest(name string, args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}
}

func TestServer_CreateProject(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	req := createCallToolRequest("create_project", map[string]interface{}{
		"id":          "test-project",
		"name":        "Test Project",
		"description": "A test project",
	})

	result, err := server.handleCreateProject(ctx, req)
	if err != nil {
		t.Fatalf("handleCreateProject() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleCreateProject() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		if err := json.Unmarshal([]byte(text.Text), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
	}

	if response["id"] != "test-project" {
		t.Errorf("response id = %v, want test-project", response["id"])
	}
}

func TestServer_CreateProject_InvalidID(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	req := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "",
		"name": "Test Project",
	})

	result, err := server.handleCreateProject(ctx, req)
	if err != nil {
		t.Fatalf("handleCreateProject() error = %v", err)
	}

	if !result.IsError {
		t.Error("handleCreateProject() should return error for invalid ID")
	}
}

func TestServer_GetProject(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	createReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createReq)

	// Get project
	getReq := createCallToolRequest("get_project", map[string]interface{}{
		"id": "test-project",
	})

	result, err := server.handleGetProject(ctx, getReq)
	if err != nil {
		t.Fatalf("handleGetProject() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleGetProject() returned error: %v", result.Content)
	}
}

func TestServer_GetProject_NotFound(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	req := createCallToolRequest("get_project", map[string]interface{}{
		"id": "nonexistent",
	})

	result, err := server.handleGetProject(ctx, req)
	if err != nil {
		t.Fatalf("handleGetProject() error = %v", err)
	}

	if !result.IsError {
		t.Error("handleGetProject() should return error for nonexistent project")
	}
}

func TestServer_ListProjects(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create projects
	for i := 0; i < 3; i++ {
		createReq := createCallToolRequest("create_project", map[string]interface{}{
			"id":   string(rune('a' + i)),
			"name": "Project",
		})
		server.handleCreateProject(ctx, createReq)
	}

	// List projects
	listReq := createCallToolRequest("list_projects", map[string]interface{}{})

	result, err := server.handleListProjects(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListProjects() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleListProjects() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		if err := json.Unmarshal([]byte(text.Text), &response); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
	}

	if response["total"].(float64) != 3 {
		t.Errorf("response total = %v, want 3", response["total"])
	}
}

func TestServer_ListProjects_Pagination(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create projects
	for i := 0; i < 10; i++ {
		createReq := createCallToolRequest("create_project", map[string]interface{}{
			"id":   string(rune('a' + i)),
			"name": "Project",
		})
		server.handleCreateProject(ctx, createReq)
	}

	// List with pagination
	listReq := createCallToolRequest("list_projects", map[string]interface{}{
		"limit":  float64(3),
		"offset": float64(0),
	})

	result, err := server.handleListProjects(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListProjects() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 10 {
		t.Errorf("response total = %v, want 10", response["total"])
	}
	if response["has_more"] != true {
		t.Error("response has_more should be true")
	}

	projects := response["projects"].([]interface{})
	if len(projects) != 3 {
		t.Errorf("response projects length = %d, want 3", len(projects))
	}
}

func TestServer_CreateTask(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	// Create task
	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id":  "test-project",
		"id":          "fix-bug",
		"name":        "Fix Login Bug",
		"description": "Fix the login issue",
	})

	result, err := server.handleCreateTask(ctx, createTaskReq)
	if err != nil {
		t.Fatalf("handleCreateTask() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleCreateTask() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["id"] != "fix-bug" {
		t.Errorf("response id = %v, want fix-bug", response["id"])
	}
	if response["status"] != "open" {
		t.Errorf("response status = %v, want open", response["status"])
	}
}

func TestServer_ListTasks(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	// Create tasks
	for i := 0; i < 3; i++ {
		createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
			"project_id": "test-project",
			"id":         string(rune('a' + i)),
			"name":       "Task",
		})
		server.handleCreateTask(ctx, createTaskReq)
	}

	// List tasks
	listReq := createCallToolRequest("list_tasks", map[string]interface{}{
		"project_id": "test-project",
	})

	result, err := server.handleListTasks(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListTasks() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 3 {
		t.Errorf("response total = %v, want 3", response["total"])
	}
}

func TestServer_ListTasks_StatusFilter(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	// Create open task
	createTaskReq1 := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "open-task",
		"name":       "Open Task",
	})
	server.handleCreateTask(ctx, createTaskReq1)

	// Create and complete a task
	createTaskReq2 := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "done-task",
		"name":       "Done Task",
	})
	server.handleCreateTask(ctx, createTaskReq2)

	updateReq := createCallToolRequest("update_task", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "done-task",
		"status":     "completed",
	})
	server.handleUpdateTask(ctx, updateReq)

	// Filter by completed
	listReq := createCallToolRequest("list_tasks", map[string]interface{}{
		"project_id": "test-project",
		"status":     "completed",
	})

	result, err := server.handleListTasks(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListTasks() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 1 {
		t.Errorf("response total = %v, want 1", response["total"])
	}
}

func TestServer_UpdateTask(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Update task
	updateReq := createCallToolRequest("update_task", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"name":       "Updated Name",
		"status":     "in_progress",
	})

	result, err := server.handleUpdateTask(ctx, updateReq)
	if err != nil {
		t.Fatalf("handleUpdateTask() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["name"] != "Updated Name" {
		t.Errorf("response name = %v, want Updated Name", response["name"])
	}
	if response["status"] != "in_progress" {
		t.Errorf("response status = %v, want in_progress", response["status"])
	}
}

func TestServer_SaveArtifact(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Save artifact
	saveReq := createCallToolRequest("save_artifact", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"content":    "This is a note about the bug fix",
		"type":       "note",
	})

	result, err := server.handleSaveArtifact(ctx, saveReq)
	if err != nil {
		t.Fatalf("handleSaveArtifact() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleSaveArtifact() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["type"] != "note" {
		t.Errorf("response type = %v, want note", response["type"])
	}
}

func TestServer_ListArtifacts(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Save artifacts
	for i := 0; i < 3; i++ {
		saveReq := createCallToolRequest("save_artifact", map[string]interface{}{
			"project_id": "test-project",
			"task_id":    "fix-bug",
			"content":    "Note content",
			"type":       "note",
		})
		server.handleSaveArtifact(ctx, saveReq)
	}

	// List artifacts
	listReq := createCallToolRequest("list_artifacts", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
	})

	result, err := server.handleListArtifacts(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListArtifacts() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 3 {
		t.Errorf("response total = %v, want 3", response["total"])
	}
}

func TestServer_SearchArtifacts(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Save artifacts with different content
	saveReq1 := createCallToolRequest("save_artifact", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"content":    "Authentication error found",
		"type":       "note",
	})
	server.handleSaveArtifact(ctx, saveReq1)

	saveReq2 := createCallToolRequest("save_artifact", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"content":    "Database connection fixed",
		"type":       "note",
	})
	server.handleSaveArtifact(ctx, saveReq2)

	// Search
	searchReq := createCallToolRequest("search_artifacts", map[string]interface{}{
		"query": "authentication",
	})

	result, err := server.handleSearchArtifacts(ctx, searchReq)
	if err != nil {
		t.Fatalf("handleSearchArtifacts() error = %v", err)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 1 {
		t.Errorf("response total = %v, want 1", response["total"])
	}
}

func TestServer_ReadFile(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace with file
	tmpDir, _ := os.MkdirTemp("", "workspace-*")
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(testFile, []byte("Hello, World!"), 0644)

	// Create project and task with workspace
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":             "test-project",
		"name":           "Test Project",
		"workspace_path": tmpDir,
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Read file
	readReq := createCallToolRequest("read_file", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"file_path":  "test.txt",
		"log_read":   false,
	})

	result, err := server.handleReadFile(ctx, readReq)
	if err != nil {
		t.Fatalf("handleReadFile() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleReadFile() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["content"] != "Hello, World!" {
		t.Errorf("response content = %v, want Hello, World!", response["content"])
	}
}

func TestServer_ListFiles(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace with files
	tmpDir, _ := os.MkdirTemp("", "workspace-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("content"), 0644)

	// Create project and task with workspace
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":             "test-project",
		"name":           "Test Project",
		"workspace_path": tmpDir,
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// List files
	listReq := createCallToolRequest("list_files", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
	})

	result, err := server.handleListFiles(ctx, listReq)
	if err != nil {
		t.Fatalf("handleListFiles() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleListFiles() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 2 {
		t.Errorf("response total = %v, want 2", response["total"])
	}
}

func TestServer_SearchFiles(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace with files
	tmpDir, _ := os.MkdirTemp("", "workspace-*")
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("Hello World"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "file2.txt"), []byte("Goodbye"), 0644)

	// Create project and task with workspace
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":             "test-project",
		"name":           "Test Project",
		"workspace_path": tmpDir,
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Search files
	searchReq := createCallToolRequest("search_files", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"query":      "Hello",
		"log_search": false,
	})

	result, err := server.handleSearchFiles(ctx, searchReq)
	if err != nil {
		t.Fatalf("handleSearchFiles() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleSearchFiles() returned error: %v", result.Content)
	}

	// Parse response
	var response map[string]interface{}
	if text, ok := result.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &response)
	}

	if response["total"].(float64) != 1 {
		t.Errorf("response total = %v, want 1", response["total"])
	}
}

func TestServer_DeleteProject(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	createReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createReq)

	// Delete project
	deleteReq := createCallToolRequest("delete_project", map[string]interface{}{
		"id": "test-project",
	})

	result, err := server.handleDeleteProject(ctx, deleteReq)
	if err != nil {
		t.Fatalf("handleDeleteProject() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleDeleteProject() returned error: %v", result.Content)
	}

	// Verify deletion
	getReq := createCallToolRequest("get_project", map[string]interface{}{
		"id": "test-project",
	})

	getResult, _ := server.handleGetProject(ctx, getReq)
	if !getResult.IsError {
		t.Error("handleGetProject() should return error for deleted project")
	}
}

func TestServer_DeleteTask(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	// Delete task
	deleteReq := createCallToolRequest("delete_task", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
	})

	result, err := server.handleDeleteTask(ctx, deleteReq)
	if err != nil {
		t.Fatalf("handleDeleteTask() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleDeleteTask() returned error: %v", result.Content)
	}

	// Verify deletion
	getReq := createCallToolRequest("get_task", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
	})

	getResult, _ := server.handleGetTask(ctx, getReq)
	if !getResult.IsError {
		t.Error("handleGetTask() should return error for deleted task")
	}
}

func TestServer_DeleteArtifact(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	ctx := context.Background()

	// Create project, task, and artifact
	createProjectReq := createCallToolRequest("create_project", map[string]interface{}{
		"id":   "test-project",
		"name": "Test Project",
	})
	server.handleCreateProject(ctx, createProjectReq)

	createTaskReq := createCallToolRequest("create_task", map[string]interface{}{
		"project_id": "test-project",
		"id":         "fix-bug",
		"name":       "Fix Bug",
	})
	server.handleCreateTask(ctx, createTaskReq)

	saveReq := createCallToolRequest("save_artifact", map[string]interface{}{
		"project_id": "test-project",
		"task_id":    "fix-bug",
		"content":    "Test content",
		"type":       "note",
	})
	saveResult, _ := server.handleSaveArtifact(ctx, saveReq)

	// Get artifact ID from response
	var saveResponse map[string]interface{}
	if text, ok := saveResult.Content[0].(mcp.TextContent); ok {
		json.Unmarshal([]byte(text.Text), &saveResponse)
	}
	artifactID := saveResponse["id"].(string)

	// Delete artifact
	deleteReq := createCallToolRequest("delete_artifact", map[string]interface{}{
		"project_id":  "test-project",
		"task_id":     "fix-bug",
		"artifact_id": artifactID,
	})

	result, err := server.handleDeleteArtifact(ctx, deleteReq)
	if err != nil {
		t.Fatalf("handleDeleteArtifact() error = %v", err)
	}

	if result.IsError {
		t.Errorf("handleDeleteArtifact() returned error: %v", result.Content)
	}

	// Verify deletion
	getReq := createCallToolRequest("get_artifact", map[string]interface{}{
		"project_id":  "test-project",
		"task_id":     "fix-bug",
		"artifact_id": artifactID,
	})

	getResult, _ := server.handleGetArtifact(ctx, getReq)
	if !getResult.IsError {
		t.Error("handleGetArtifact() should return error for deleted artifact")
	}
}

func TestErrorResult(t *testing.T) {
	result := errorResult("Test error message")

	if !result.IsError {
		t.Error("errorResult().IsError should be true")
	}

	if len(result.Content) != 1 {
		t.Errorf("errorResult().Content length = %d, want 1", len(result.Content))
	}

	if text, ok := result.Content[0].(mcp.TextContent); ok {
		if text.Text != "Test error message" {
			t.Errorf("errorResult().Content text = %q, want %q", text.Text, "Test error message")
		}
	} else {
		t.Error("errorResult().Content[0] should be TextContent")
	}
}

func TestProjectToMap(t *testing.T) {
	p := task.NewProject(task.ProjectID("test"), "Test Project")
	p.Description = "A description"
	p.WorkspacePath = "/some/path"

	m := projectToMap(p)

	if m["id"] != task.ProjectID("test") {
		t.Errorf("projectToMap id = %v, want test", m["id"])
	}
	if m["name"] != "Test Project" {
		t.Errorf("projectToMap name = %v, want Test Project", m["name"])
	}
	if m["description"] != "A description" {
		t.Errorf("projectToMap description = %v, want A description", m["description"])
	}
	if m["workspace_path"] != "/some/path" {
		t.Errorf("projectToMap workspace_path = %v, want /some/path", m["workspace_path"])
	}
}

func TestTaskToMap(t *testing.T) {
	projectID := task.ProjectID("test-project")
	taskObj := task.NewTask(projectID, task.TaskID("fix-bug"), "Fix Bug")
	taskObj.Description = "Bug description"
	taskObj.Status = task.TaskStatusInProgress

	m := taskToMap(taskObj)

	if m["id"] != task.TaskID("fix-bug") {
		t.Errorf("taskToMap id = %v, want fix-bug", m["id"])
	}
	if m["project_id"] != projectID {
		t.Errorf("taskToMap project_id = %v, want %v", m["project_id"], projectID)
	}
	if m["name"] != "Fix Bug" {
		t.Errorf("taskToMap name = %v, want Fix Bug", m["name"])
	}
	if m["status"] != task.TaskStatusInProgress {
		t.Errorf("taskToMap status = %v, want in_progress", m["status"])
	}
}

func TestArtifactToMap(t *testing.T) {
	projectID := task.ProjectID("test-project")
	taskID := task.TaskID("fix-bug")
	a := task.NewArtifact(projectID, taskID, task.ArtifactTypeNote, "Content here")

	m := artifactToMap(a)

	if m["project_id"] != projectID {
		t.Errorf("artifactToMap project_id = %v, want %v", m["project_id"], projectID)
	}
	if m["task_id"] != taskID {
		t.Errorf("artifactToMap task_id = %v, want %v", m["task_id"], taskID)
	}
	if m["type"] != task.ArtifactTypeNote {
		t.Errorf("artifactToMap type = %v, want note", m["type"])
	}
	if m["content"] != "Content here" {
		t.Errorf("artifactToMap content = %v, want Content here", m["content"])
	}
}
