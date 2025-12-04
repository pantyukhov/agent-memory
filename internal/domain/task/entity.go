package task

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// ProjectID represents a unique project identifier (like Jira project key: PROJ)
type ProjectID string

// NewProjectID creates a ProjectID from a string.
func NewProjectID(id string) ProjectID {
	// Normalize: lowercase, replace spaces with dashes
	normalized := strings.ToLower(strings.TrimSpace(id))
	normalized = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(normalized, "-")
	normalized = regexp.MustCompile(`-+`).ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return ProjectID(normalized)
}

func (p ProjectID) String() string {
	return string(p)
}

// IsValid checks if the ProjectID is valid.
func (p ProjectID) IsValid() bool {
	if p == "" {
		return false
	}
	// Must be alphanumeric with optional dashes
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`, string(p))
	return matched
}

// Project represents a project that contains tasks (like Jira project).
type Project struct {
	ID            ProjectID         `json:"id"`
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	WorkspacePath string            `json:"workspace_path,omitempty"` // Default workspace for tasks
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// NewProject creates a new Project.
func NewProject(id ProjectID, name string) *Project {
	now := time.Now().UTC()
	return &Project{
		ID:        id,
		Name:      name,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// TaskID represents a unique task identifier (like Jira key: PROJECT-123)
type TaskID string

// NewTaskID creates a TaskID from a string.
func NewTaskID(id string) TaskID {
	// Normalize: lowercase, replace spaces with dashes
	normalized := strings.ToLower(strings.TrimSpace(id))
	normalized = regexp.MustCompile(`[^a-z0-9-]`).ReplaceAllString(normalized, "-")
	normalized = regexp.MustCompile(`-+`).ReplaceAllString(normalized, "-")
	normalized = strings.Trim(normalized, "-")
	return TaskID(normalized)
}

func (t TaskID) String() string {
	return string(t)
}

// IsValid checks if the TaskID is valid.
func (t TaskID) IsValid() bool {
	if t == "" {
		return false
	}
	// Must start with letter or number
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$|^[a-z0-9]$`, string(t))
	return matched
}

// Task represents a task/issue that the agent is working on.
type Task struct {
	ID            TaskID            `json:"id"`
	ProjectID     ProjectID         `json:"project_id"`               // Parent project
	Name          string            `json:"name"`
	Description   string            `json:"description,omitempty"`
	Status        TaskStatus        `json:"status"`
	WorkspacePath string            `json:"workspace_path,omitempty"` // Root directory for file operations (overrides project)
	Metadata      map[string]string `json:"metadata,omitempty"`
	CreatedAt     time.Time         `json:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at"`
}

// TaskStatus represents the status of a task.
type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "open"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusArchived   TaskStatus = "archived"
)

// NewTask creates a new Task within a project.
func NewTask(projectID ProjectID, id TaskID, name string) *Task {
	now := time.Now().UTC()
	return &Task{
		ID:        id,
		ProjectID: projectID,
		Name:      name,
		Status:    TaskStatusOpen,
		Metadata:  make(map[string]string),
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Artifact represents a piece of work artifact associated with a task.
type Artifact struct {
	ID        string            `json:"id"`
	ProjectID ProjectID         `json:"project_id"`
	TaskID    TaskID            `json:"task_id"`
	Type      ArtifactType      `json:"type"`
	Content   string            `json:"content"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	CreatedAt time.Time         `json:"created_at"`
}

// ArtifactType defines the type of artifact.
type ArtifactType string

const (
	ArtifactTypeNote       ArtifactType = "note"
	ArtifactTypeCode       ArtifactType = "code"
	ArtifactTypeDecision   ArtifactType = "decision"
	ArtifactTypeDiscussion ArtifactType = "discussion"
	ArtifactTypeReference  ArtifactType = "reference"
	ArtifactTypeFileRead   ArtifactType = "file_read"   // Log of file read operation
	ArtifactTypeFileList   ArtifactType = "file_list"   // Log of directory listing
	ArtifactTypeSearch     ArtifactType = "search"      // Log of search operation
	ArtifactTypeGeneric    ArtifactType = "artifact"
)

// NewArtifact creates a new Artifact within a project and task.
func NewArtifact(projectID ProjectID, taskID TaskID, artifactType ArtifactType, content string) *Artifact {
	now := time.Now().UTC()
	return &Artifact{
		ID:        generateArtifactID(now),
		ProjectID: projectID,
		TaskID:    taskID,
		Type:      artifactType,
		Content:   content,
		Metadata:  make(map[string]string),
		CreatedAt: now,
	}
}

// Filename returns the filename for this artifact.
func (a *Artifact) Filename() string {
	return fmt.Sprintf("%s.%d.md", a.Type, a.CreatedAt.Unix())
}

func generateArtifactID(t time.Time) string {
	return fmt.Sprintf("%d", t.UnixNano())
}
