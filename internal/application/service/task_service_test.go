package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"agent-memory/internal/domain/task"
	"agent-memory/internal/infrastructure/storage/filesystem"
)

func setupTestService(t *testing.T) (*TaskService, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agent-memory-service-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repo, err := filesystem.NewRepository(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create repository: %v", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	svc := NewTaskService(repo, logger)

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	}

	return svc, cleanup
}

func TestTaskService_CreateProject(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateProjectRequest{
		ID:          "test-project",
		Name:        "Test Project",
		Description: "A test project",
	}

	project, err := svc.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.ID != task.ProjectID("test-project") {
		t.Errorf("CreateProject().ID = %q, want %q", project.ID, "test-project")
	}
	if project.Name != "Test Project" {
		t.Errorf("CreateProject().Name = %q, want %q", project.Name, "Test Project")
	}
}

func TestTaskService_CreateProject_InvalidID(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateProjectRequest{
		ID:   "", // Invalid - empty
		Name: "Test Project",
	}

	_, err := svc.CreateProject(ctx, req)
	if err != task.ErrInvalidProjectID {
		t.Errorf("CreateProject() error = %v, want ErrInvalidProjectID", err)
	}
}

func TestTaskService_CreateProject_DefaultName(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	req := CreateProjectRequest{
		ID:   "test-project",
		Name: "", // Should default to ID
	}

	project, err := svc.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.Name != "test-project" {
		t.Errorf("CreateProject().Name = %q, want %q", project.Name, "test-project")
	}
}

func TestTaskService_GetProject(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	req := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	_, err := svc.CreateProject(ctx, req)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Get project
	project, err := svc.GetProject(ctx, "test-project")
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if project.Name != "Test Project" {
		t.Errorf("GetProject().Name = %q, want %q", project.Name, "Test Project")
	}
}

func TestTaskService_ListProjects(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create projects
	for i := 0; i < 5; i++ {
		req := CreateProjectRequest{
			ID:   string(rune('a' + i)),
			Name: "Project",
		}
		_, err := svc.CreateProject(ctx, req)
		if err != nil {
			t.Fatalf("CreateProject() error = %v", err)
		}
	}

	// List projects
	result, err := svc.ListProjects(ctx, ListProjectsRequest{Limit: 10})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(result.Items) != 5 {
		t.Errorf("ListProjects() returned %d items, want 5", len(result.Items))
	}
}

func TestTaskService_ListProjects_Pagination(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create projects
	for i := 0; i < 10; i++ {
		req := CreateProjectRequest{
			ID:   string(rune('a' + i)),
			Name: "Project",
		}
		_, err := svc.CreateProject(ctx, req)
		if err != nil {
			t.Fatalf("CreateProject() error = %v", err)
		}
	}

	// Test pagination
	result, err := svc.ListProjects(ctx, ListProjectsRequest{Limit: 3, Offset: 0})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(result.Items) != 3 {
		t.Errorf("ListProjects() returned %d items, want 3", len(result.Items))
	}
	if result.Total != 10 {
		t.Errorf("ListProjects().Total = %d, want 10", result.Total)
	}
	if !result.HasMore {
		t.Error("ListProjects().HasMore should be true")
	}
}

func TestTaskService_UpdateProject(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	createReq := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	_, err := svc.CreateProject(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Update project
	newName := "Updated Name"
	newDesc := "Updated description"
	updateReq := UpdateProjectRequest{
		ID:          "test-project",
		Name:        &newName,
		Description: &newDesc,
	}

	project, err := svc.UpdateProject(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	if project.Name != "Updated Name" {
		t.Errorf("UpdateProject().Name = %q, want %q", project.Name, "Updated Name")
	}
	if project.Description != "Updated description" {
		t.Errorf("UpdateProject().Description = %q, want %q", project.Description, "Updated description")
	}
}

func TestTaskService_DeleteProject(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	createReq := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	_, err := svc.CreateProject(ctx, createReq)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Delete project
	err = svc.DeleteProject(ctx, "test-project")
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify deletion
	_, err = svc.GetProject(ctx, "test-project")
	if err != task.ErrProjectNotFound {
		t.Errorf("GetProject() after delete error = %v, want ErrProjectNotFound", err)
	}
}

func TestTaskService_CreateTask(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	projectReq := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	_, err := svc.CreateProject(ctx, projectReq)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create task
	taskReq := CreateTaskRequest{
		ProjectID:   "test-project",
		ID:          "fix-bug",
		Name:        "Fix Login Bug",
		Description: "Fix the login issue",
	}

	taskObj, err := svc.CreateTask(ctx, taskReq)
	if err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	if taskObj.ID != task.TaskID("fix-bug") {
		t.Errorf("CreateTask().ID = %q, want %q", taskObj.ID, "fix-bug")
	}
	if taskObj.Name != "Fix Login Bug" {
		t.Errorf("CreateTask().Name = %q, want %q", taskObj.Name, "Fix Login Bug")
	}
	if taskObj.Status != task.TaskStatusOpen {
		t.Errorf("CreateTask().Status = %q, want %q", taskObj.Status, task.TaskStatusOpen)
	}
}

func TestTaskService_CreateTask_InvalidID(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	projectReq := CreateProjectRequest{
		ID:   "test-project",
		Name: "Test Project",
	}
	_, err := svc.CreateProject(ctx, projectReq)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Try to create task with invalid ID
	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "", // Invalid
		Name:      "Fix Bug",
	}

	_, err = svc.CreateTask(ctx, taskReq)
	if err != task.ErrInvalidTaskID {
		t.Errorf("CreateTask() error = %v, want ErrInvalidTaskID", err)
	}
}

func TestTaskService_GetTask(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Get task
	taskObj, err := svc.GetTask(ctx, "test-project", "fix-bug")
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if taskObj.Name != "Fix Bug" {
		t.Errorf("GetTask().Name = %q, want %q", taskObj.Name, "Fix Bug")
	}
}

func TestTaskService_ListTasks(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	// Create tasks
	for i := 0; i < 5; i++ {
		taskReq := CreateTaskRequest{
			ProjectID: "test-project",
			ID:        string(rune('a' + i)),
			Name:      "Task",
		}
		svc.CreateTask(ctx, taskReq)
	}

	// List tasks
	result, err := svc.ListTasks(ctx, ListTasksRequest{ProjectID: "test-project", Limit: 10})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Items) != 5 {
		t.Errorf("ListTasks() returned %d items, want 5", len(result.Items))
	}
}

func TestTaskService_ListTasks_StatusFilter(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	// Create open task
	taskReq1 := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "open-task",
		Name:      "Open Task",
	}
	svc.CreateTask(ctx, taskReq1)

	// Create and complete a task
	taskReq2 := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "done-task",
		Name:      "Done Task",
	}
	svc.CreateTask(ctx, taskReq2)

	completedStatus := task.TaskStatusCompleted
	svc.UpdateTask(ctx, UpdateTaskRequest{
		ProjectID: "test-project",
		ID:        "done-task",
		Status:    &completedStatus,
	})

	// Filter by completed status
	result, err := svc.ListTasks(ctx, ListTasksRequest{
		ProjectID: "test-project",
		Limit:     10,
		Status:    task.TaskStatusCompleted,
	})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("ListTasks() with filter returned %d items, want 1", len(result.Items))
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Update task
	newName := "Updated Name"
	newStatus := task.TaskStatusInProgress
	updateReq := UpdateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      &newName,
		Status:    &newStatus,
	}

	taskObj, err := svc.UpdateTask(ctx, updateReq)
	if err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	if taskObj.Name != "Updated Name" {
		t.Errorf("UpdateTask().Name = %q, want %q", taskObj.Name, "Updated Name")
	}
	if taskObj.Status != task.TaskStatusInProgress {
		t.Errorf("UpdateTask().Status = %q, want %q", taskObj.Status, task.TaskStatusInProgress)
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Delete task
	err := svc.DeleteTask(ctx, "test-project", "fix-bug")
	if err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify deletion
	_, err = svc.GetTask(ctx, "test-project", "fix-bug")
	if err != task.ErrTaskNotFound {
		t.Errorf("GetTask() after delete error = %v, want ErrTaskNotFound", err)
	}
}

func TestTaskService_SaveArtifact(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Save artifact
	artifactReq := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Type:      task.ArtifactTypeNote,
		Content:   "This is a note",
	}

	artifact, err := svc.SaveArtifact(ctx, artifactReq)
	if err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	if artifact.Content != "This is a note" {
		t.Errorf("SaveArtifact().Content = %q, want %q", artifact.Content, "This is a note")
	}
	if artifact.Type != task.ArtifactTypeNote {
		t.Errorf("SaveArtifact().Type = %q, want %q", artifact.Type, task.ArtifactTypeNote)
	}
}

func TestTaskService_SaveArtifact_DefaultType(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Save artifact without type
	artifactReq := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Content:   "Generic content",
	}

	artifact, err := svc.SaveArtifact(ctx, artifactReq)
	if err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	if artifact.Type != task.ArtifactTypeGeneric {
		t.Errorf("SaveArtifact().Type = %q, want %q", artifact.Type, task.ArtifactTypeGeneric)
	}
}

func TestTaskService_GetArtifact(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project, task, and artifact
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	artifactReq := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Type:      task.ArtifactTypeNote,
		Content:   "Test content",
	}
	artifact, _ := svc.SaveArtifact(ctx, artifactReq)

	// Get artifact
	got, err := svc.GetArtifact(ctx, "test-project", "fix-bug", artifact.ID)
	if err != nil {
		t.Fatalf("GetArtifact() error = %v", err)
	}

	if got.Content != "Test content" {
		t.Errorf("GetArtifact().Content = %q, want %q", got.Content, "Test content")
	}
}

func TestTaskService_ListArtifacts(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Create artifacts with delays to ensure different timestamps
	for i := 0; i < 5; i++ {
		artifactReq := SaveArtifactRequest{
			ProjectID: "test-project",
			TaskID:    "fix-bug",
			Type:      task.ArtifactTypeNote,
			Content:   "Note content",
		}
		svc.SaveArtifact(ctx, artifactReq)
		time.Sleep(10 * time.Millisecond)
	}

	// List artifacts
	result, err := svc.ListArtifacts(ctx, ListArtifactsRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Limit:     10,
	})
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}

	if len(result.Items) != 5 {
		t.Errorf("ListArtifacts() returned %d items, want 5", len(result.Items))
	}
}

func TestTaskService_SearchArtifacts(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	// Create artifacts with different content (with delay for unique timestamps)
	artifactReq1 := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Type:      task.ArtifactTypeNote,
		Content:   "Authentication error found",
	}
	svc.SaveArtifact(ctx, artifactReq1)

	time.Sleep(10 * time.Millisecond)

	artifactReq2 := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Type:      task.ArtifactTypeNote,
		Content:   "Database connection fixed",
	}
	svc.SaveArtifact(ctx, artifactReq2)

	// Search
	result, err := svc.SearchArtifacts(ctx, SearchArtifactsRequest{
		Query: "authentication",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("SearchArtifacts() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("SearchArtifacts() returned %d items, want 1", len(result.Items))
	}
}

func TestTaskService_DeleteArtifact(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project, task, and artifact
	projectReq := CreateProjectRequest{ID: "test-project", Name: "Test Project"}
	svc.CreateProject(ctx, projectReq)

	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "fix-bug",
		Name:      "Fix Bug",
	}
	svc.CreateTask(ctx, taskReq)

	artifactReq := SaveArtifactRequest{
		ProjectID: "test-project",
		TaskID:    "fix-bug",
		Type:      task.ArtifactTypeNote,
		Content:   "Test content",
	}
	artifact, _ := svc.SaveArtifact(ctx, artifactReq)

	// Delete artifact
	err := svc.DeleteArtifact(ctx, "test-project", "fix-bug", artifact.ID)
	if err != nil {
		t.Fatalf("DeleteArtifact() error = %v", err)
	}

	// Verify deletion
	_, err = svc.GetArtifact(ctx, "test-project", "fix-bug", artifact.ID)
	if err != task.ErrArtifactNotFound {
		t.Errorf("GetArtifact() after delete error = %v, want ErrArtifactNotFound", err)
	}
}

func TestTaskService_GetEffectiveWorkspacePath(t *testing.T) {
	svc, cleanup := setupTestService(t)
	defer cleanup()

	ctx := context.Background()

	// Create project with workspace
	projectReq := CreateProjectRequest{
		ID:            "test-project",
		Name:          "Test Project",
		WorkspacePath: "/project/workspace",
	}
	svc.CreateProject(ctx, projectReq)

	// Create task without workspace
	taskReq := CreateTaskRequest{
		ProjectID: "test-project",
		ID:        "task1",
		Name:      "Task 1",
	}
	svc.CreateTask(ctx, taskReq)

	// Should get project workspace
	path, err := svc.GetEffectiveWorkspacePath(ctx, "test-project", "task1")
	if err != nil {
		t.Fatalf("GetEffectiveWorkspacePath() error = %v", err)
	}
	if path != "/project/workspace" {
		t.Errorf("GetEffectiveWorkspacePath() = %q, want %q", path, "/project/workspace")
	}

	// Create task with workspace override
	taskReq2 := CreateTaskRequest{
		ProjectID:     "test-project",
		ID:            "task2",
		Name:          "Task 2",
		WorkspacePath: "/task/workspace",
	}
	svc.CreateTask(ctx, taskReq2)

	// Should get task workspace
	path2, err := svc.GetEffectiveWorkspacePath(ctx, "test-project", "task2")
	if err != nil {
		t.Fatalf("GetEffectiveWorkspacePath() error = %v", err)
	}
	if path2 != "/task/workspace" {
		t.Errorf("GetEffectiveWorkspacePath() = %q, want %q", path2, "/task/workspace")
	}
}
