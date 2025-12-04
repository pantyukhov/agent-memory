package filesystem

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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
//	    project.json                    (project metadata)
//	    /[status]-task-id/              (kanban-style naming)
//	      task.json                     (task metadata)
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

// ListProjects returns projects with pagination.
func (r *Repository) ListProjects(ctx context.Context, opts task.ListOptions) (*task.ListResult[*task.Project], error) {
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

	return applyPagination(projects, opts), nil
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

	// Check if task already exists (any status)
	if existing := r.findTaskDir(t.ProjectID, t.ID); existing != "" {
		return task.ErrTaskAlreadyExists
	}

	taskDir := r.taskDirPath(t.ProjectID, t)

	// Create task directory with kanban-style name
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
	taskDir := r.findTaskDir(projectID, taskID)
	if taskDir == "" {
		return nil, task.ErrTaskNotFound
	}

	return r.loadTaskMetadataFromDir(taskDir)
}

// ListTasks returns tasks for a project with pagination.
func (r *Repository) ListTasks(ctx context.Context, projectID task.ProjectID, opts task.ListOptions) (*task.ListResult[*task.Task], error) {
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

		// Parse kanban-style directory name: [status]-task-id
		taskDir := filepath.Join(projectDir, entry.Name())
		metadataPath := filepath.Join(taskDir, taskMetadataFile)
		if _, err := os.Stat(metadataPath); os.IsNotExist(err) {
			continue
		}

		t, err := r.loadTaskMetadataFromDir(taskDir)
		if err != nil {
			continue // Skip invalid tasks
		}

		// Apply status filter if specified
		if opts.Status != "" && t.Status != opts.Status {
			continue
		}

		tasks = append(tasks, t)
	}

	// Sort by updated time, newest first
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].UpdatedAt.After(tasks[j].UpdatedAt)
	})

	return applyPagination(tasks, opts), nil
}

// UpdateTask updates task metadata and renames directory if status changed.
func (r *Repository) UpdateTask(ctx context.Context, t *task.Task) error {
	oldDir := r.findTaskDir(t.ProjectID, t.ID)
	if oldDir == "" {
		return task.ErrTaskNotFound
	}

	t.UpdatedAt = time.Now().UTC()

	// Check if we need to rename directory (status changed)
	newDir := r.taskDirPath(t.ProjectID, t)
	if oldDir != newDir {
		if err := os.Rename(oldDir, newDir); err != nil {
			return fmt.Errorf("%w: failed to rename task directory: %v", task.ErrStorageFailed, err)
		}
	}

	return r.saveTaskMetadataToDir(newDir, t)
}

// DeleteTask removes a task and all its artifacts.
func (r *Repository) DeleteTask(ctx context.Context, projectID task.ProjectID, taskID task.TaskID) error {
	taskDir := r.findTaskDir(projectID, taskID)
	if taskDir == "" {
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
	taskDir := r.findTaskDir(a.ProjectID, a.TaskID)
	if taskDir == "" {
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
	t, err := r.loadTaskMetadataFromDir(taskDir)
	if err == nil {
		t.UpdatedAt = time.Now().UTC()
		r.saveTaskMetadataToDir(taskDir, t)
	}

	return nil
}

// GetArtifact retrieves an artifact by project ID, task ID, and artifact ID.
func (r *Repository) GetArtifact(ctx context.Context, projectID task.ProjectID, taskID task.TaskID, artifactID string) (*task.Artifact, error) {
	result, err := r.ListArtifacts(ctx, projectID, taskID, task.ListOptions{Limit: 0})
	if err != nil {
		return nil, err
	}

	for _, a := range result.Items {
		if a.ID == artifactID {
			return a, nil
		}
	}

	return nil, task.ErrArtifactNotFound
}

// ListArtifacts returns artifacts for a task with pagination.
func (r *Repository) ListArtifacts(ctx context.Context, projectID task.ProjectID, taskID task.TaskID, opts task.ListOptions) (*task.ListResult[*task.Artifact], error) {
	taskDir := r.findTaskDir(projectID, taskID)
	if taskDir == "" {
		return nil, task.ErrTaskNotFound
	}

	artifactsPath := filepath.Join(taskDir, artifactsDir)
	entries, err := os.ReadDir(artifactsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &task.ListResult[*task.Artifact]{
				Items:   []*task.Artifact{},
				Total:   0,
				Limit:   opts.Limit,
				Offset:  opts.Offset,
				HasMore: false,
			}, nil
		}
		return nil, fmt.Errorf("%w: %v", task.ErrStorageFailed, err)
	}

	var artifacts []*task.Artifact
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		a, err := r.loadArtifact(projectID, taskID, taskDir, entry.Name())
		if err != nil {
			continue // Skip invalid artifacts
		}
		artifacts = append(artifacts, a)
	}

	// Sort by created time, newest first
	sort.Slice(artifacts, func(i, j int) bool {
		return artifacts[i].CreatedAt.After(artifacts[j].CreatedAt)
	})

	return applyPagination(artifacts, opts), nil
}

// SearchArtifacts searches artifact content across all projects/tasks or specific project/task.
func (r *Repository) SearchArtifacts(ctx context.Context, query string, projectID *task.ProjectID, taskID *task.TaskID, opts task.ListOptions) (*task.ListResult[*task.Artifact], error) {
	queryLower := strings.ToLower(query)
	var results []*task.Artifact

	// Determine which projects to search
	var projectIDs []task.ProjectID
	if projectID != nil {
		projectIDs = []task.ProjectID{*projectID}
	} else {
		projectsResult, err := r.ListProjects(ctx, task.ListOptions{Limit: 0})
		if err != nil {
			return nil, err
		}
		for _, p := range projectsResult.Items {
			projectIDs = append(projectIDs, p.ID)
		}
	}

	for _, pid := range projectIDs {
		// Determine which tasks to search within this project
		var taskIDs []task.TaskID
		if taskID != nil && projectID != nil {
			taskIDs = []task.TaskID{*taskID}
		} else {
			tasksResult, err := r.ListTasks(ctx, pid, task.ListOptions{Limit: 0})
			if err != nil {
				continue
			}
			for _, t := range tasksResult.Items {
				taskIDs = append(taskIDs, t.ID)
			}
		}

		for _, tid := range taskIDs {
			artifactsResult, err := r.ListArtifacts(ctx, pid, tid, task.ListOptions{Limit: 0})
			if err != nil {
				continue
			}

			for _, a := range artifactsResult.Items {
				if strings.Contains(strings.ToLower(a.Content), queryLower) {
					results = append(results, a)
				}
			}
		}
	}

	return applyPagination(results, opts), nil
}

// DeleteArtifact removes an artifact.
func (r *Repository) DeleteArtifact(ctx context.Context, projectID task.ProjectID, taskID task.TaskID, artifactID string) error {
	taskDir := r.findTaskDir(projectID, taskID)
	if taskDir == "" {
		return task.ErrTaskNotFound
	}

	artifactsResult, err := r.ListArtifacts(ctx, projectID, taskID, task.ListOptions{Limit: 0})
	if err != nil {
		return err
	}

	for _, a := range artifactsResult.Items {
		if a.ID == artifactID {
			artifactsPath := filepath.Join(taskDir, artifactsDir)
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

// taskDirPath returns the path for a task directory using kanban-style naming.
func (r *Repository) taskDirPath(projectID task.ProjectID, t *task.Task) string {
	return filepath.Join(r.basePath, projectID.String(), t.DirName())
}

// findTaskDir finds the actual directory for a task (handling different statuses).
// Returns empty string if not found.
func (r *Repository) findTaskDir(projectID task.ProjectID, taskID task.TaskID) string {
	projectDir := r.projectPath(projectID)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return ""
	}

	// Pattern: [status]-taskid
	pattern := regexp.MustCompile(`^\[([a-z_]+)\]-(.+)$`)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		matches := pattern.FindStringSubmatch(name)
		if matches != nil && task.TaskID(matches[2]) == taskID {
			return filepath.Join(projectDir, name)
		}

		// Also check for legacy directories without status prefix
		if task.TaskID(name) == taskID {
			return filepath.Join(projectDir, name)
		}
	}

	return ""
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
	taskDir := r.taskDirPath(t.ProjectID, t)
	return r.saveTaskMetadataToDir(taskDir, t)
}

func (r *Repository) saveTaskMetadataToDir(taskDir string, t *task.Task) error {
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

func (r *Repository) loadTaskMetadataFromDir(taskDir string) (*task.Task, error) {
	metadataPath := filepath.Join(taskDir, taskMetadataFile)

	data, err := os.ReadFile(metadataPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Try to extract task ID from directory name
			dirName := filepath.Base(taskDir)
			pattern := regexp.MustCompile(`^\[([a-z_]+)\]-(.+)$`)
			matches := pattern.FindStringSubmatch(dirName)

			var taskID task.TaskID
			var status task.TaskStatus = task.TaskStatusOpen
			if matches != nil {
				status = task.TaskStatus(matches[1])
				taskID = task.TaskID(matches[2])
			} else {
				taskID = task.TaskID(dirName)
			}

			// Extract project ID from parent directory
			projectID := task.ProjectID(filepath.Base(filepath.Dir(taskDir)))

			return &task.Task{
				ID:        taskID,
				ProjectID: projectID,
				Name:      taskID.String(),
				Status:    status,
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

func (r *Repository) loadArtifact(projectID task.ProjectID, taskID task.TaskID, taskDir, filename string) (*task.Artifact, error) {
	artifactsPath := filepath.Join(taskDir, artifactsDir)
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

// applyPagination applies pagination options to a slice and returns ListResult.
func applyPagination[T any](items []T, opts task.ListOptions) *task.ListResult[T] {
	total := len(items)

	// Apply default limit if not specified
	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}

	offset := opts.Offset
	if offset < 0 {
		offset = 0
	}

	// Apply offset
	if offset >= total {
		return &task.ListResult[T]{
			Items:   []T{},
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: false,
		}
	}

	items = items[offset:]

	// Apply limit
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}

	return &task.ListResult[T]{
		Items:   items,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
		HasMore: hasMore,
	}
}
