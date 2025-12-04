package service

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"agent-memory/internal/infrastructure/storage/filesystem"
)

func setupWorkspaceTestService(t *testing.T) (*WorkspaceService, *TaskService, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agent-memory-workspace-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repoDir := filepath.Join(tmpDir, "repo")
	repo, err := filesystem.NewRepository(repoDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create repository: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	taskSvc := NewTaskService(repo, logger)
	workspaceSvc := NewWorkspaceService(repo, logger)

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	}

	return workspaceSvc, taskSvc, tmpDir, cleanup
}

func TestWorkspaceService_ReadFile(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with a file
	workspaceDir := filepath.Join(tmpDir, "workspace")
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		t.Fatalf("failed to create workspace dir: %v", err)
	}

	testFile := filepath.Join(workspaceDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("Hello, World!"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Create project and task with workspace
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Read file
	result, err := workspaceSvc.ReadFile(ctx, ReadFileRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		FilePath:  "test.txt",
		LogRead:   false,
	})
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if result.Content != "Hello, World!" {
		t.Errorf("ReadFile().Content = %q, want %q", result.Content, "Hello, World!")
	}
	if result.Size != 13 {
		t.Errorf("ReadFile().Size = %d, want 13", result.Size)
	}
}

func TestWorkspaceService_ReadFile_WithLogging(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with a file
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	testFile := filepath.Join(workspaceDir, "test.txt")
	os.WriteFile(testFile, []byte("Test content"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Read file with logging
	_, err := workspaceSvc.ReadFile(ctx, ReadFileRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		FilePath:  "test.txt",
		LogRead:   true,
	})
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	// Check that artifact was created
	artifacts, err := taskSvc.ListArtifacts(ctx, ListArtifactsRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}

	if len(artifacts.Items) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(artifacts.Items))
	}
}

func TestWorkspaceService_ReadFile_OutsideWorkspace(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)

	// Create a file outside workspace
	outsideFile := filepath.Join(tmpDir, "outside.txt")
	os.WriteFile(outsideFile, []byte("Secret"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Try to read file outside workspace
	_, err := workspaceSvc.ReadFile(ctx, ReadFileRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		FilePath:  "../outside.txt",
		LogRead:   false,
	})

	if err == nil {
		t.Error("ReadFile() should fail for path outside workspace")
	}
}

func TestWorkspaceService_ListFiles(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file2.txt"), []byte("content"), 0644)
	os.MkdirAll(filepath.Join(workspaceDir, "subdir"), 0755)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// List files
	result, err := workspaceSvc.ListFiles(ctx, ListFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Recursive: false,
	})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if result.Total != 3 { // file1.txt, file2.txt, subdir
		t.Errorf("ListFiles().Total = %d, want 3", result.Total)
	}
}

func TestWorkspaceService_ListFiles_WithPattern(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file2.go"), []byte("content"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file3.txt"), []byte("content"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// List only .txt files
	result, err := workspaceSvc.ListFiles(ctx, ListFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Pattern:   "*.txt",
		Recursive: false,
	})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("ListFiles().Total = %d, want 2", result.Total)
	}
}

func TestWorkspaceService_ListFiles_Recursive(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with nested files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "root.txt"), []byte("content"), 0644)

	subdir := filepath.Join(workspaceDir, "subdir")
	os.MkdirAll(subdir, 0755)
	os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("content"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// List files recursively
	result, err := workspaceSvc.ListFiles(ctx, ListFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Recursive: true,
	})
	if err != nil {
		t.Fatalf("ListFiles() error = %v", err)
	}

	if result.Total != 3 { // root.txt, subdir, subdir/nested.txt
		t.Errorf("ListFiles().Total = %d, want 3", result.Total)
	}
}

func TestWorkspaceService_SearchFiles(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("Hello World"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file2.txt"), []byte("Goodbye World"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file3.txt"), []byte("No match here"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Search for "World"
	result, err := workspaceSvc.SearchFiles(ctx, SearchFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Query:     "World",
		LogSearch: false,
	})
	if err != nil {
		t.Fatalf("SearchFiles() error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("SearchFiles().Total = %d, want 2", result.Total)
	}
}

func TestWorkspaceService_SearchFiles_CaseInsensitive(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("HELLO World"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file2.txt"), []byte("hello world"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Search case-insensitive
	result, err := workspaceSvc.SearchFiles(ctx, SearchFilesRequest{
		ProjectID:  "test-project",
		TaskID:     "fix-bug",
		Query:      "hello",
		IgnoreCase: true,
		LogSearch:  false,
	})
	if err != nil {
		t.Fatalf("SearchFiles() error = %v", err)
	}

	if result.Total != 2 {
		t.Errorf("SearchFiles().Total = %d, want 2", result.Total)
	}
}

func TestWorkspaceService_SearchFiles_WithPattern(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("Match content"), 0644)
	os.WriteFile(filepath.Join(workspaceDir, "file2.go"), []byte("Match content"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Search only in .txt files
	result, err := workspaceSvc.SearchFiles(ctx, SearchFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Query:     "Match",
		Pattern:   "*.txt",
		LogSearch: false,
	})
	if err != nil {
		t.Fatalf("SearchFiles() error = %v", err)
	}

	if result.Total != 1 {
		t.Errorf("SearchFiles().Total = %d, want 1", result.Total)
	}
}

func TestWorkspaceService_SearchFiles_WithLogging(t *testing.T) {
	workspaceSvc, taskSvc, tmpDir, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create workspace directory with files
	workspaceDir := filepath.Join(tmpDir, "workspace")
	os.MkdirAll(workspaceDir, 0755)
	os.WriteFile(filepath.Join(workspaceDir, "file1.txt"), []byte("Search me"), 0644)

	// Create project and task
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: workspaceDir,
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Search with logging
	_, err := workspaceSvc.SearchFiles(ctx, SearchFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Query:     "Search",
		LogSearch: true,
	})
	if err != nil {
		t.Fatalf("SearchFiles() error = %v", err)
	}

	// Check that artifact was created
	artifacts, err := taskSvc.ListArtifacts(ctx, ListArtifactsRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}

	if len(artifacts.Items) != 1 {
		t.Errorf("Expected 1 artifact, got %d", len(artifacts.Items))
	}
}

func TestWorkspaceService_NoWorkspacePath(t *testing.T) {
	workspaceSvc, taskSvc, _, cleanup := setupWorkspaceTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task without workspace
	projectReq := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	taskSvc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	taskSvc.CreateTask(ctx, taskReq)

	// Try to list files
	_, err := workspaceSvc.ListFiles(ctx, ListFilesRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
	})

	if err == nil {
		t.Error("ListFiles() should fail when no workspace configured")
	}
}
