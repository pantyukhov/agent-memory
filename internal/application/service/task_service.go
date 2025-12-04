package service

import (
	"context"
	"fmt"
	"log/slog"

	"agent-memory/internal/domain/task"
)

// TaskService provides project, task, and artifact management operations.
type TaskService struct {
	repo   task.Repository
	logger *slog.Logger
}

// NewTaskService creates a new task service.
func NewTaskService(repo task.Repository, logger *slog.Logger) *TaskService {
	return &TaskService{
		repo:   repo,
		logger: logger,
	}
}

// Project operations

// CreateProjectRequest contains parameters for creating a project.
type CreateProjectRequest struct {
	ID            string
	Name          string
	Description   string
	WorkspacePath string
	Metadata      map[string]string
}

// CreateProject creates a new project.
func (s *TaskService) CreateProject(ctx context.Context, req CreateProjectRequest) (*task.Project, error) {
	projectID := task.NewProjectID(req.ID)
	if !projectID.IsValid() {
		return nil, task.ErrInvalidProjectID
	}

	name := req.Name
	if name == "" {
		name = req.ID
	}

	p := task.NewProject(projectID, name)
	p.Description = req.Description
	p.WorkspacePath = req.WorkspacePath
	if req.Metadata != nil {
		p.Metadata = req.Metadata
	}

	if err := s.repo.CreateProject(ctx, p); err != nil {
		s.logger.Error("failed to create project", "id", projectID, "error", err)
		return nil, fmt.Errorf("creating project: %w", err)
	}

	s.logger.Info("project created", "id", projectID, "name", name)
	return p, nil
}

// GetProject retrieves a project by ID.
func (s *TaskService) GetProject(ctx context.Context, id string) (*task.Project, error) {
	projectID := task.NewProjectID(id)
	return s.repo.GetProject(ctx, projectID)
}

// ListProjects returns all projects.
func (s *TaskService) ListProjects(ctx context.Context) ([]*task.Project, error) {
	return s.repo.ListProjects(ctx)
}

// UpdateProjectRequest contains parameters for updating a project.
type UpdateProjectRequest struct {
	ID            string
	Name          *string
	Description   *string
	WorkspacePath *string
	Metadata      map[string]string
}

// UpdateProject updates a project.
func (s *TaskService) UpdateProject(ctx context.Context, req UpdateProjectRequest) (*task.Project, error) {
	projectID := task.NewProjectID(req.ID)

	p, err := s.repo.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		p.Name = *req.Name
	}
	if req.Description != nil {
		p.Description = *req.Description
	}
	if req.WorkspacePath != nil {
		p.WorkspacePath = *req.WorkspacePath
	}
	if req.Metadata != nil {
		p.Metadata = req.Metadata
	}

	if err := s.repo.UpdateProject(ctx, p); err != nil {
		s.logger.Error("failed to update project", "id", projectID, "error", err)
		return nil, fmt.Errorf("updating project: %w", err)
	}

	s.logger.Info("project updated", "id", projectID)
	return p, nil
}

// DeleteProject deletes a project and all its tasks.
func (s *TaskService) DeleteProject(ctx context.Context, id string) error {
	projectID := task.NewProjectID(id)

	if err := s.repo.DeleteProject(ctx, projectID); err != nil {
		s.logger.Error("failed to delete project", "id", projectID, "error", err)
		return err
	}

	s.logger.Info("project deleted", "id", projectID)
	return nil
}

// Task operations

// CreateTaskRequest contains parameters for creating a task.
type CreateTaskRequest struct {
	ProjectID     string
	ID            string
	Name          string
	Description   string
	WorkspacePath string // Overrides project workspace if set
	Metadata      map[string]string
}

// CreateTask creates a new task within a project.
func (s *TaskService) CreateTask(ctx context.Context, req CreateTaskRequest) (*task.Task, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.ID)

	if !taskID.IsValid() {
		return nil, task.ErrInvalidTaskID
	}

	name := req.Name
	if name == "" {
		name = req.ID
	}

	t := task.NewTask(projectID, taskID, name)
	t.Description = req.Description
	t.WorkspacePath = req.WorkspacePath
	if req.Metadata != nil {
		t.Metadata = req.Metadata
	}

	if err := s.repo.CreateTask(ctx, t); err != nil {
		s.logger.Error("failed to create task", "project_id", projectID, "task_id", taskID, "error", err)
		return nil, fmt.Errorf("creating task: %w", err)
	}

	s.logger.Info("task created", "project_id", projectID, "task_id", taskID, "name", name)
	return t, nil
}

// GetTask retrieves a task by project ID and task ID.
func (s *TaskService) GetTask(ctx context.Context, projectID, taskID string) (*task.Task, error) {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)
	return s.repo.GetTask(ctx, pid, tid)
}

// ListTasks returns all tasks for a project.
func (s *TaskService) ListTasks(ctx context.Context, projectID string) ([]*task.Task, error) {
	pid := task.NewProjectID(projectID)
	return s.repo.ListTasks(ctx, pid)
}

// UpdateTaskRequest contains parameters for updating a task.
type UpdateTaskRequest struct {
	ProjectID     string
	ID            string
	Name          *string
	Description   *string
	WorkspacePath *string
	Status        *task.TaskStatus
	Metadata      map[string]string
}

// UpdateTask updates a task.
func (s *TaskService) UpdateTask(ctx context.Context, req UpdateTaskRequest) (*task.Task, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.ID)

	t, err := s.repo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		t.Name = *req.Name
	}
	if req.Description != nil {
		t.Description = *req.Description
	}
	if req.WorkspacePath != nil {
		t.WorkspacePath = *req.WorkspacePath
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	if req.Metadata != nil {
		t.Metadata = req.Metadata
	}

	if err := s.repo.UpdateTask(ctx, t); err != nil {
		s.logger.Error("failed to update task", "project_id", projectID, "task_id", taskID, "error", err)
		return nil, fmt.Errorf("updating task: %w", err)
	}

	s.logger.Info("task updated", "project_id", projectID, "task_id", taskID)
	return t, nil
}

// DeleteTask deletes a task and all its artifacts.
func (s *TaskService) DeleteTask(ctx context.Context, projectID, taskID string) error {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)

	if err := s.repo.DeleteTask(ctx, pid, tid); err != nil {
		s.logger.Error("failed to delete task", "project_id", pid, "task_id", tid, "error", err)
		return err
	}

	s.logger.Info("task deleted", "project_id", pid, "task_id", tid)
	return nil
}

// Artifact operations

// SaveArtifactRequest contains parameters for saving an artifact.
type SaveArtifactRequest struct {
	ProjectID string
	TaskID    string
	Type      task.ArtifactType
	Content   string
	Metadata  map[string]string
}

// SaveArtifact saves an artifact to a task.
func (s *TaskService) SaveArtifact(ctx context.Context, req SaveArtifactRequest) (*task.Artifact, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.TaskID)

	// Verify task exists
	if _, err := s.repo.GetTask(ctx, projectID, taskID); err != nil {
		return nil, err
	}

	artifactType := req.Type
	if artifactType == "" {
		artifactType = task.ArtifactTypeGeneric
	}

	a := task.NewArtifact(projectID, taskID, artifactType, req.Content)
	if req.Metadata != nil {
		a.Metadata = req.Metadata
	}

	if err := s.repo.SaveArtifact(ctx, a); err != nil {
		s.logger.Error("failed to save artifact", "project_id", projectID, "task_id", taskID, "error", err)
		return nil, fmt.Errorf("saving artifact: %w", err)
	}

	s.logger.Info("artifact saved", "project_id", projectID, "task_id", taskID, "type", artifactType, "id", a.ID)
	return a, nil
}

// GetArtifact retrieves an artifact.
func (s *TaskService) GetArtifact(ctx context.Context, projectID, taskID, artifactID string) (*task.Artifact, error) {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)
	return s.repo.GetArtifact(ctx, pid, tid, artifactID)
}

// ListArtifacts returns all artifacts for a task.
func (s *TaskService) ListArtifacts(ctx context.Context, projectID, taskID string) ([]*task.Artifact, error) {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)
	return s.repo.ListArtifacts(ctx, pid, tid)
}

// SearchArtifactsRequest contains parameters for searching artifacts.
type SearchArtifactsRequest struct {
	Query     string
	ProjectID string // Optional: limit search to specific project
	TaskID    string // Optional: limit search to specific task
}

// SearchArtifacts searches artifact content.
func (s *TaskService) SearchArtifacts(ctx context.Context, req SearchArtifactsRequest) ([]*task.Artifact, error) {
	var projectID *task.ProjectID
	var taskID *task.TaskID

	if req.ProjectID != "" {
		pid := task.NewProjectID(req.ProjectID)
		projectID = &pid
	}
	if req.TaskID != "" {
		tid := task.NewTaskID(req.TaskID)
		taskID = &tid
	}

	return s.repo.SearchArtifacts(ctx, req.Query, projectID, taskID)
}

// DeleteArtifact removes an artifact.
func (s *TaskService) DeleteArtifact(ctx context.Context, projectID, taskID, artifactID string) error {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)

	if err := s.repo.DeleteArtifact(ctx, pid, tid, artifactID); err != nil {
		s.logger.Error("failed to delete artifact", "project_id", pid, "task_id", tid, "artifact_id", artifactID, "error", err)
		return err
	}

	s.logger.Info("artifact deleted", "project_id", pid, "task_id", tid, "artifact_id", artifactID)
	return nil
}

// GetEffectiveWorkspacePath returns the workspace path for a task,
// falling back to project workspace if task doesn't have one.
func (s *TaskService) GetEffectiveWorkspacePath(ctx context.Context, projectID, taskID string) (string, error) {
	pid := task.NewProjectID(projectID)
	tid := task.NewTaskID(taskID)

	t, err := s.repo.GetTask(ctx, pid, tid)
	if err != nil {
		return "", err
	}

	if t.WorkspacePath != "" {
		return t.WorkspacePath, nil
	}

	// Fall back to project workspace
	p, err := s.repo.GetProject(ctx, pid)
	if err != nil {
		return "", err
	}

	return p.WorkspacePath, nil
}
