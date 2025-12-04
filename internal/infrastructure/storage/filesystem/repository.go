package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"agent-memory/internal/domain/task"
)

const (
	projectMetadataFile = "project.json"
	taskMetadataFile    = "task.json"
	artifactsDir        = "artifacts"
)

// Repository implements task.Repository using filesystem storage.
// Directory structure:
//
//	/base_path/
//	  /project-id/
//	    project.json           (project metadata)
//	    /task-id/
//	      task.json            (task metadata)
//	      /artifacts/
//	        note.1234567890.md
//	        code.1234567891.md
type Repository struct {
	basePath string
}

// NewRepository creates a new filesystem repository.
func NewRepository(basePath string) (*Repository, error) {
	// Create base directory if it doesn't exist
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &Repository{basePath: basePath}, nil
}

// Project operations

// CreateProject creates a new project directory structure.
func (r *Repository) CreateProject(ctx context.Context, p *task.Project) error {
	projectDir := r.projectPath(p.ID)

	// Check if project already exists
	if _, err := os.Stat(projectDir); err == nil {
		return task.ErrProjectAlreadyExists
	}

	// Create project directory
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	// Save project metadata
	return r.saveProjectMetadata(p)
}

// GetProject retrieves a project by ID.
func (r *Repository) GetProject(ctx context.Context, id task.ProjectID) (*task.Project, error) {
	projectDir := r.projectPath(id)

	// Check if project exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, task.ErrProjectNotFound
	}

	return r.loadProjectMetadata(id)
}

// ListProjects returns all projects.
func (r *Repository) ListProjects(ctx context.Context) ([]*task.Project, error) {
	entries, err := os.ReadDir(r.basePath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var projects []*task.Project
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectID := task.ProjectID(entry.Name())
		p, err := r.loadProjectMetadata(projectID)
		if err != nil {
			continue // Skip invalid projects
		}
		projects = append(projects, p)
	}

	// Sort by updated time, newest first
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].UpdatedAt.After(projects[j].UpdatedAt)
	})

	return projects, nil
}

// UpdateProject updates project metadata.
func (r *Repository) UpdateProject(ctx context.Context, p *task.Project) error {
	projectDir := r.projectPath(p.ID)

	// Check if project exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return task.ErrProjectNotFound
	}

	p.UpdatedAt = time.Now().UTC()
	return r.saveProjectMetadata(p)
}

// DeleteProject removes a project and all its tasks.
func (r *Repository) DeleteProject(ctx context.Context, id task.ProjectID) error {
	projectDir := r.projectPath(id)

	// Check if project exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return task.ErrProjectNotFound
	}

	if err := os.RemoveAll(projectDir); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return nil
}

// Task operations

// CreateTask creates a new task directory structure within a project.
func (r *Repository) CreateTask(ctx context.Context, t *task.Task) error {
	// Verify project exists
	projectDir := r.projectPath(t.ProjectID)
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return task.ErrProjectNotFound
	}

	taskDir := r.taskPath(t.ProjectID, t.ID)

	// Check if task already exists
	if _, err := os.Stat(taskDir); err == nil {
		return task.ErrTaskAlreadyExists
	}

	// Create task directory
	if err := os.MkdirAll(taskDir, 0755); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	// Create artifacts subdirectory
	artifactsPath := filepath.Join(taskDir, artifactsDir)
	if err := os.MkdirAll(artifactsPath, 0755); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	// Save task metadata
	return r.saveTaskMetadata(t)
}

// GetTask retrieves a task by project ID and task ID.
func (r *Repository) GetTask(ctx context.Context, projectID task.ProjectID, taskID task.TaskID) (*task.Task, error) {
	taskDir := r.taskPath(projectID, taskID)

	// Check if task exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return nil, task.ErrTaskNotFound
	}

	return r.loadTaskMetadata(projectID, taskID)
}

// ListTasks returns all tasks for a project.
func (r *Repository) ListTasks(ctx context.Context, projectID task.ProjectID) ([]*task.Task, error) {
	projectDir := r.projectPath(projectID)

	// Check if project exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		return nil, task.ErrProjectNotFound
	}

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip if it's not a task directory (check for task.json)
		taskID := task.TaskID(entry.Name())
		taskDir := r.taskPath(projectID, taskID)
		metadataPath := filepath.Join(taskDir, taskMetadataFile)
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			continue
		}

		t, err := r.loadTaskMetadata(projectID, taskID)
		if err != nil {
			continue // Skip invalid tasks
		}
		tasks = append(tasks, t)
	}

	// Sort by updated time, newest first
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return tasks, nil
}

// UpdateTask updates task metadata.
func (r *Repository) UpdateTask(ctx context.Context, t *task.Task) error {
	taskDir := r.taskPath(t.ProjectID, t.ID)

	// Check if task exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return task.ErrTaskNotFound
	}

	t.UpdatedAt = time.Now().UTC()
	return r.saveTaskMetadata(t)
}

// DeleteTask removes a task and all its artifacts.
func (r *Repository) DeleteTask(ctx context.Context, projectID task.ProjectID, taskID task.TaskID) error {
	taskDir := r.taskPath(projectID, taskID)

	// Check if task exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return task.ErrTaskNotFound
	}

	if err := os.RemoveAll(taskDir); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return nil
}

// Artifact operations

// SaveArtifact saves an artifact to a task.
func (r *Repository) SaveArtifact(ctx context.Context, a *task.Artifact) error {
	taskDir := r.taskPath(a.ProjectID, a.TaskID)

	// Check if task exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return task.ErrTaskNotFound
	}

	artifactsPath := filepath.Join(taskDir, artifactsDir)
	if err := os.MkdirAll(artifactsPath, 0755); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	// Create artifact file
	filename := a.Filename()
	filePath := filepath.Join(artifactsPath, filename)

	// Build markdown content with frontmatter
	content := r.buildArtifactMarkdown(a)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	// Update task's updated_at
	t, err := r.loadTaskMetadata(a.ProjectID, a.TaskID)
	if err == nil {
		t.UpdatedAt = time.Now().UTC()
		r.saveTaskMetadata(t)
	}

	return nil
}

// GetArtifact retrieves an artifact by project ID, task ID, and artifact ID.
func (r *Repository) GetArtifact(ctx context.Context, projectID task.ProjectID, taskID task.TaskID, artifactID string) (*task.Artifact, error) {
	artifacts, err := r.ListArtifacts(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}

	for _, a := range artifacts {
		if a.ID == artifactID {
			return a, nil
		}
	}

	return nil, task.ErrArtifactNotFound
}

// ListArtifacts returns all artifacts for a task.
func (r *Repository) ListArtifacts(ctx context.Context, projectID task.ProjectID, taskID task.TaskID) ([]*task.Artifact, error) {
	taskDir := r.taskPath(projectID, taskID)
	artifactsPath := filepath.Join(taskDir, artifactsDir)

	// Check if task exists
	if _, err := os.Stat(taskDir); os.IsNotExist(err) {
		return nil, task.ErrTaskNotFound
	}

	entries, err := os.ReadDir(artifactsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*task.Artifact{}, nil
		}
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var artifacts []*task.Artifact
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		a, err := r.loadArtifact(projectID, taskID, entry.Name())
		if err != nil {
			continue // Skip invalid artifacts
		}
		artifacts = append(artifacts, a)
	}

	// Sort by created time, newest first
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].CreatedAt.After(artifacts[j].CreatedAt)
	})

	return artifacts, nil
}

// SearchArtifacts searches artifact content across all projects/tasks or specific project/task.
func (r *Repository) SearchArtifacts(ctx context.Context, query string, projectID *task.ProjectID, taskID *task.TaskID) ([]*task.Artifact, error) {
	queryLower := strings.ToLower(query)
	var results []*task.Artifact

	// Determine which projects to search
	var projectIDs []task.ProjectID
	if projectID != nil {
		projectIDs = []task.ProjectID{*projectID}
	} else {
		projects, err := r.ListProjects(ctx)
		if err != nil {
			return nil, err
		}
		for _, p := range projects {
			projectIDs = append(projectIDs, p.ID)
		}
	}

	for _, pid := range projectIDs {
		// Determine which tasks to search within this project
		var taskIDs []task.TaskID
		if taskID != nil && projectID != nil {
			taskIDs = []task.TaskID{*taskID}
		} else {
			tasks, err := r.ListTasks(ctx, pid)
			if err != nil {
				continue
			}
			for _, t := range tasks {
				taskIDs = append(taskIDs, t.ID)
			}
		}

		for _, tid := range taskIDs {
			artifacts, err := r.ListArtifacts(ctx, pid, tid)
			if err != nil {
				continue
			}

			for _, a := range artifacts {
				if strings.Contains(strings.ToLower(a.Content), queryLower) {
					results = append(results, a)
				}
			}
		}
	}

	return results, nil
}

// DeleteArtifact removes an artifact.
func (r *Repository) DeleteArtifact(ctx context.Context, projectID task.ProjectID, taskID task.TaskID, artifactID string) error {
	artifacts, err := r.ListArtifacts(ctx, projectID, taskID)
	if err != nil {
		return err
	}

	for _, a := range artifacts {
		if a.ID == artifactID {
			artifactsPath := filepath.Join(r.taskPath(projectID, taskID), artifactsDir)
			filePath := filepath.Join(artifactsPath, a.Filename())

			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
			}
			return nil
		}
	}

	return task.ErrArtifactNotFound
}

// Close releases any resources.
func (r *Repository) Close() error {
	return nil
}

// Helper methods

func (r *Repository) projectPath(id task.ProjectID) string {
	return filepath.Join(r.basePath, id.String())
}

func (r *Repository) taskPath(projectID task.ProjectID, taskID task.TaskID) string {
	return filepath.Join(r.basePath, projectID.String(), taskID.String())
}

func (r *Repository) saveProjectMetadata(p *task.Project) error {
	projectDir := r.projectPath(p.ID)
	metadataPath := filepath.Join(projectDir, projectMetadataFile)

	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return nil
}

func (r *Repository) loadProjectMetadata(id task.ProjectID) (*task.Project, error) {
	projectDir := r.projectPath(id)
	metadataPath := filepath.Join(projectDir, projectMetadataFile)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default project metadata for legacy directories
			return &task.Project{
				ID:        id,
				Name:      id.String(),
				Metadata:  make(map[string]string),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var p task.Project
	if err := json.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return &p, nil
}

func (r *Repository) saveTaskMetadata(t *task.Task) error {
	taskDir := r.taskPath(t.ProjectID, t.ID)
	metadataPath := filepath.Join(taskDir, taskMetadataFile)

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	if err := os.WriteFile(metadataPath, data, 0644); err != nil {
		return fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return nil
}

func (r *Repository) loadTaskMetadata(projectID task.ProjectID, taskID task.TaskID) (*task.Task, error) {
	taskDir := r.taskPath(projectID, taskID)
	metadataPath := filepath.Join(taskDir, taskMetadataFile)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create default task metadata for legacy directories
			return &task.Task{
				ID:        taskID,
				ProjectID: projectID,
				Name:      taskID.String(),
				Status:    task.TaskStatusOpen,
				Metadata:  make(map[string]string),
				CreatedAt: time.Now().UTC(),
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var t task.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	return &t, nil
}

func (r *Repository) buildArtifactMarkdown(a *task.Artifact) string {
	var sb strings.Builder

	// Frontmatter
	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("id: %s\n", a.ID))
	sb.WriteString(fmt.Sprintf("project_id: %s\n", a.ProjectID))
	sb.WriteString(fmt.Sprintf("task_id: %s\n", a.TaskID))
	sb.WriteString(fmt.Sprintf("type: %s\n", a.Type))
	sb.WriteString(fmt.Sprintf("created_at: %s\n", a.CreatedAt.Format(time.RFC3339)))

	if len(a.Metadata) > 0 {
		sb.WriteString("metadata:\n")
		for k, v := range a.Metadata {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", k, v))
		}
	}

	sb.WriteString("---\n\n")
	sb.WriteString(a.Content)

	return sb.String()
}

func (r *Repository) loadArtifact(projectID task.ProjectID, taskID task.TaskID, filename string) (*task.Artifact, error) {
	artifactsPath := filepath.Join(r.taskPath(projectID, taskID), artifactsDir)
	filePath := filepath.Join(artifactsPath, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	content := string(data)

	// Parse filename to extract type and timestamp
	// Format: type.timestamp.md
	parts := strings.Split(strings.TrimSuffix(filename, ".md"), ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid artifact filename: %s", filename)
	}

	artifactType := task.ArtifactType(parts[0])
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in filename: %s", filename)
	}

	createdAt := time.Unix(timestamp, 0).UTC()

	// Extract content from markdown (skip frontmatter)
	actualContent := content
	if strings.HasPrefix(content, "---\n") {
		endIdx := strings.Index(content[4:], "\n---\n")
		if endIdx != -1 {
			actualContent = strings.TrimSpace(content[4+endIdx+5:])
		}
	}

	return &task.Artifact{
		ID:        fmt.Sprintf("%d", timestamp),
		ProjectID: projectID,
		TaskID:    taskID,
		Type:      artifactType,
		Content:   actualContent,
		Metadata:  make(map[string]string),
		CreatedAt: createdAt,
	}, nil
}
