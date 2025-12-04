package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"agent-memory/internal/domain/task"
)

func setupTestRepo(t *testing.T) (*Repository, string, func()) {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "agent-memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	repo, err := NewRepository(tmpDir)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create repository: %v", err)
	}

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tmpDir)
	}

	return repo, tmpDir, cleanup
}

func TestNewRepository(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "agent-memory-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	repoPath := filepath.Join(tmpDir, "tasks")

	repo, err := NewRepository(repoPath)
	if err != nil {
		t.Fatalf("NewRepository() error = %v", err)
	}
	defer repo.Close()

	// Verify directory was created
	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		t.Error("NewRepository() should create base directory")
	}
}

func TestRepository_CreateProject(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	project.Description = "A test project"

	err := repo.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Verify project was created
	got, err := repo.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if got.ID != project.ID {
		t.Errorf("GetProject().ID = %q, want %q", got.ID, project.ID)
	}
	if got.Name != project.Name {
		t.Errorf("GetProject().Name = %q, want %q", got.Name, project.Name)
	}
	if got.Description != project.Description {
		t.Errorf("GetProject().Description = %q, want %q", got.Description, project.Description)
	}
}

func TestRepository_CreateProject_AlreadyExists(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")

	err := repo.CreateProject(ctx, project)
	if err != nil {
		t.Fatalf("CreateProject() first call error = %v", err)
	}

	// Try to create again
	err = repo.CreateProject(ctx, project)
	if err != task.ErrProjectAlreadyExists {
		t.Errorf("CreateProject() error = %v, want ErrProjectAlreadyExists", err)
	}
}

func TestRepository_GetProject_NotFound(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	_, err := repo.GetProject(ctx, task.ProjectID("nonexistent"))
	if err != task.ErrProjectNotFound {
		t.Errorf("GetProject() error = %v, want ErrProjectNotFound", err)
	}
}

func TestRepository_ListProjects(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create multiple projects
	for i := 0; i < 5; i++ {
		id := task.ProjectID(string(rune('a' + i)))
		project := task.NewProject(id, "Project "+string(rune('A'+i)))
		if err := repo.CreateProject(ctx, project); err != nil {
			t.Fatalf("CreateProject() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// List all projects
	result, err := repo.ListProjects(ctx, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}

	if len(result.Items) != 5 {
		t.Errorf("ListProjects() returned %d items, want 5", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("ListProjects().Total = %d, want 5", result.Total)
	}
	if result.HasMore {
		t.Error("ListProjects().HasMore should be false")
	}
}

func TestRepository_ListProjects_Pagination(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create 10 projects
	for i := 0; i < 10; i++ {
		id := task.ProjectID(string(rune('a' + i)))
		project := task.NewProject(id, "Project "+string(rune('A'+i)))
		if err := repo.CreateProject(ctx, project); err != nil {
			t.Fatalf("CreateProject() error = %v", err)
		}
	}

	// Test pagination
	result, err := repo.ListProjects(ctx, task.ListOptions{Limit: 3, Offset: 0})
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

	// Test offset
	result2, err := repo.ListProjects(ctx, task.ListOptions{Limit: 3, Offset: 3})
	if err != nil {
		t.Fatalf("ListProjects() with offset error = %v", err)
	}

	if len(result2.Items) != 3 {
		t.Errorf("ListProjects() with offset returned %d items, want 3", len(result2.Items))
	}
}

func TestRepository_UpdateProject(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")

	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Update project
	project.Name = "Updated Name"
	project.Description = "Updated description"

	if err := repo.UpdateProject(ctx, project); err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	// Verify update
	got, err := repo.GetProject(ctx, project.ID)
	if err != nil {
		t.Fatalf("GetProject() error = %v", err)
	}

	if got.Name != "Updated Name" {
		t.Errorf("GetProject().Name = %q, want %q", got.Name, "Updated Name")
	}
	if got.Description != "Updated description" {
		t.Errorf("GetProject().Description = %q, want %q", got.Description, "Updated description")
	}
}

func TestRepository_DeleteProject(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")

	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if err := repo.DeleteProject(ctx, project.ID); err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// Verify deletion
	_, err := repo.GetProject(ctx, project.ID)
	if err != task.ErrProjectNotFound {
		t.Errorf("GetProject() after delete error = %v, want ErrProjectNotFound", err)
	}
}

func TestRepository_CreateTask(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project first
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create task
	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Login Bug")
	taskObj.Description = "Fix the login bug"

	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Verify task was created
	got, err := repo.GetTask(ctx, project.ID, taskObj.ID)
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}

	if got.ID != taskObj.ID {
		t.Errorf("GetTask().ID = %q, want %q", got.ID, taskObj.ID)
	}
	if got.Name != taskObj.Name {
		t.Errorf("GetTask().Name = %q, want %q", got.Name, taskObj.Name)
	}
	if got.Status != task.TaskStatusOpen {
		t.Errorf("GetTask().Status = %q, want %q", got.Status, task.TaskStatusOpen)
	}
}

func TestRepository_CreateTask_NoProject(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	taskObj := task.NewTask(task.ProjectID("nonexistent"), task.TaskID("fix-bug"), "Fix Bug")

	err := repo.CreateTask(ctx, taskObj)
	if err != task.ErrProjectNotFound {
		t.Errorf("CreateTask() error = %v, want ErrProjectNotFound", err)
	}
}

func TestRepository_CreateTask_AlreadyExists(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create task
	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() first call error = %v", err)
	}

	// Try to create again
	err := repo.CreateTask(ctx, taskObj)
	if err != task.ErrTaskAlreadyExists {
		t.Errorf("CreateTask() error = %v, want ErrTaskAlreadyExists", err)
	}
}

func TestRepository_ListTasks(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create tasks
	for i := 0; i < 5; i++ {
		taskObj := task.NewTask(project.ID, task.TaskID(string(rune('a'+i))), "Task")
		if err := repo.CreateTask(ctx, taskObj); err != nil {
			t.Fatalf("CreateTask() error = %v", err)
		}
	}

	// List all tasks
	result, err := repo.ListTasks(ctx, project.ID, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Items) != 5 {
		t.Errorf("ListTasks() returned %d items, want 5", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("ListTasks().Total = %d, want 5", result.Total)
	}
}

func TestRepository_ListTasks_StatusFilter(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create tasks with different statuses
	taskOpen := task.NewTask(project.ID, task.TaskID("open-task"), "Open Task")
	if err := repo.CreateTask(ctx, taskOpen); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	taskCompleted := task.NewTask(project.ID, task.TaskID("completed-task"), "Completed Task")
	taskCompleted.Status = task.TaskStatusCompleted
	if err := repo.CreateTask(ctx, taskCompleted); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Filter by status
	result, err := repo.ListTasks(ctx, project.ID, task.ListOptions{
		Limit:  10,
		Status: task.TaskStatusCompleted,
	})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("ListTasks() with status filter returned %d items, want 1", len(result.Items))
	}
	if result.Items[0].ID != task.TaskID("completed-task") {
		t.Errorf("ListTasks() returned wrong task: %q", result.Items[0].ID)
	}
}

func TestRepository_UpdateTask_StatusChange(t *testing.T) {
	repo, tmpDir, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create task
	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Verify initial directory name
	openDir := filepath.Join(tmpDir, "test-project", "[open]-fix-bug")
	if _, err := os.Stat(openDir); os.IsNotExist(err) {
		t.Error("Task directory should exist with [open] prefix")
	}

	// Update status
	taskObj.Status = task.TaskStatusCompleted
	if err := repo.UpdateTask(ctx, taskObj); err != nil {
		t.Fatalf("UpdateTask() error = %v", err)
	}

	// Verify directory was renamed
	completedDir := filepath.Join(tmpDir, "test-project", "[completed]-fix-bug")
	if _, err := os.Stat(completedDir); os.IsNotExist(err) {
		t.Error("Task directory should be renamed to [completed] prefix")
	}

	// Verify old directory no longer exists
	if _, err := os.Stat(openDir); !os.IsNotExist(err) {
		t.Error("Old task directory should not exist")
	}

	// Verify task can still be retrieved
	got, err := repo.GetTask(ctx, project.ID, taskObj.ID)
	if err != nil {
		t.Fatalf("GetTask() after status change error = %v", err)
	}
	if got.Status != task.TaskStatusCompleted {
		t.Errorf("GetTask().Status = %q, want %q", got.Status, task.TaskStatusCompleted)
	}
}

func TestRepository_DeleteTask(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create task
	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Delete task
	if err := repo.DeleteTask(ctx, project.ID, taskObj.ID); err != nil {
		t.Fatalf("DeleteTask() error = %v", err)
	}

	// Verify deletion
	_, err := repo.GetTask(ctx, project.ID, taskObj.ID)
	if err != task.ErrTaskNotFound {
		t.Errorf("GetTask() after delete error = %v, want ErrTaskNotFound", err)
	}
}

func TestRepository_SaveArtifact(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save artifact
	artifact := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "This is a note")
	if err := repo.SaveArtifact(ctx, artifact); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	// Verify artifact can be listed
	result, err := repo.ListArtifacts(ctx, project.ID, taskObj.ID, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("ListArtifacts() returned %d items, want 1", len(result.Items))
	}
}

func TestRepository_GetArtifact(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save artifact
	artifact := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "This is a note")
	if err := repo.SaveArtifact(ctx, artifact); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	// Get artifact by ID
	got, err := repo.GetArtifact(ctx, project.ID, taskObj.ID, artifact.ID)
	if err != nil {
		t.Fatalf("GetArtifact() error = %v", err)
	}

	if got.Content != "This is a note" {
		t.Errorf("GetArtifact().Content = %q, want %q", got.Content, "This is a note")
	}
}

func TestRepository_GetArtifact_NotFound(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Try to get non-existent artifact
	_, err := repo.GetArtifact(ctx, project.ID, taskObj.ID, "nonexistent")
	if err != task.ErrArtifactNotFound {
		t.Errorf("GetArtifact() error = %v, want ErrArtifactNotFound", err)
	}
}

func TestRepository_ListArtifacts_Pagination(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save multiple artifacts
	for i := 0; i < 10; i++ {
		artifact := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "Note content")
		if err := repo.SaveArtifact(ctx, artifact); err != nil {
			t.Fatalf("SaveArtifact() error = %v", err)
		}
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Test pagination
	result, err := repo.ListArtifacts(ctx, project.ID, taskObj.ID, task.ListOptions{Limit: 3})
	if err != nil {
		t.Fatalf("ListArtifacts() error = %v", err)
	}

	if len(result.Items) != 3 {
		t.Errorf("ListArtifacts() returned %d items, want 3", len(result.Items))
	}
	if result.Total != 10 {
		t.Errorf("ListArtifacts().Total = %d, want 10", result.Total)
	}
	if !result.HasMore {
		t.Error("ListArtifacts().HasMore should be true")
	}
}

func TestRepository_SearchArtifacts(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save artifacts with different content
	artifact1 := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "Authentication error found")
	if err := repo.SaveArtifact(ctx, artifact1); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	artifact2 := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "Fixed the database issue")
	if err := repo.SaveArtifact(ctx, artifact2); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	// Search for authentication
	result, err := repo.SearchArtifacts(ctx, "authentication", nil, nil, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("SearchArtifacts() error = %v", err)
	}

	if len(result.Items) != 1 {
		t.Errorf("SearchArtifacts() returned %d items, want 1", len(result.Items))
	}

	// Search for database
	result2, err := repo.SearchArtifacts(ctx, "database", nil, nil, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("SearchArtifacts() error = %v", err)
	}

	if len(result2.Items) != 1 {
		t.Errorf("SearchArtifacts() returned %d items, want 1", len(result2.Items))
	}
}

func TestRepository_SearchArtifacts_WithProjectFilter(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create two projects
	project1 := task.NewProject(task.ProjectID("project1"), "Project 1")
	if err := repo.CreateProject(ctx, project1); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	project2 := task.NewProject(task.ProjectID("project2"), "Project 2")
	if err := repo.CreateProject(ctx, project2); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	// Create tasks in each project
	task1 := task.NewTask(project1.ID, task.TaskID("task1"), "Task 1")
	if err := repo.CreateTask(ctx, task1); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	task2 := task.NewTask(project2.ID, task.TaskID("task2"), "Task 2")
	if err := repo.CreateTask(ctx, task2); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save artifacts with same keyword in both projects
	artifact1 := task.NewArtifact(project1.ID, task1.ID, task.ArtifactTypeNote, "Important keyword here")
	if err := repo.SaveArtifact(ctx, artifact1); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	artifact2 := task.NewArtifact(project2.ID, task2.ID, task.ArtifactTypeNote, "Another important keyword")
	if err := repo.SaveArtifact(ctx, artifact2); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	// Search all projects
	result, err := repo.SearchArtifacts(ctx, "keyword", nil, nil, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("SearchArtifacts() error = %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("SearchArtifacts() all projects returned %d items, want 2", len(result.Items))
	}

	// Search only project1
	pid := project1.ID
	result2, err := repo.SearchArtifacts(ctx, "keyword", &pid, nil, task.ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("SearchArtifacts() with project filter error = %v", err)
	}
	if len(result2.Items) != 1 {
		t.Errorf("SearchArtifacts() with project filter returned %d items, want 1", len(result2.Items))
	}
}

func TestRepository_DeleteArtifact(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	ctx := context.Background()

	// Create project and task
	project := task.NewProject(task.ProjectID("test-project"), "Test Project")
	if err := repo.CreateProject(ctx, project); err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	taskObj := task.NewTask(project.ID, task.TaskID("fix-bug"), "Fix Bug")
	if err := repo.CreateTask(ctx, taskObj); err != nil {
		t.Fatalf("CreateTask() error = %v", err)
	}

	// Save artifact
	artifact := task.NewArtifact(project.ID, taskObj.ID, task.ArtifactTypeNote, "This is a note")
	if err := repo.SaveArtifact(ctx, artifact); err != nil {
		t.Fatalf("SaveArtifact() error = %v", err)
	}

	// Delete artifact
	if err := repo.DeleteArtifact(ctx, project.ID, taskObj.ID, artifact.ID); err != nil {
		t.Fatalf("DeleteArtifact() error = %v", err)
	}

	// Verify deletion
	_, err := repo.GetArtifact(ctx, project.ID, taskObj.ID, artifact.ID)
	if err != task.ErrArtifactNotFound {
		t.Errorf("GetArtifact() after delete error = %v, want ErrArtifactNotFound", err)
	}
}

func TestRepository_Close(t *testing.T) {
	repo, _, cleanup := setupTestRepo(t)
	defer cleanup()

	err := repo.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
