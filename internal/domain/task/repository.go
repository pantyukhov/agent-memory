package task

import (
	"context"
)

// ListOptions contains pagination and filtering options for list operations.
type ListOptions struct {
	Limit  int        // Maximum number of items to return (0 = no limit, default 50)
	Offset int        // Number of items to skip
	Status TaskStatus // Filter by status (empty = all)
}

// ListResult contains paginated results with metadata.
type ListResult[T any] struct {
	Items      []T  `json:"items"`
	Total      int  `json:"total"`       // Total count without pagination
	Limit      int  `json:"limit"`       // Applied limit
	Offset     int  `json:"offset"`      // Applied offset
	HasMore    bool `json:"has_more"`    // Whether there are more items
}

// Repository defines the contract for project, task, and artifact storage.
type Repository interface {
	// Project operations

	// CreateProject creates a new project directory structure.
	CreateProject(ctx context.Context, project *Project) error

	// GetProject retrieves a project by ID.
	GetProject(ctx context.Context, id ProjectID) (*Project, error)

	// ListProjects returns projects with pagination.
	ListProjects(ctx context.Context, opts ListOptions) (*ListResult[*Project], error)

	// UpdateProject updates project metadata.
	UpdateProject(ctx context.Context, project *Project) error

	// DeleteProject removes a project and all its tasks.
	DeleteProject(ctx context.Context, id ProjectID) error

	// Task operations

	// CreateTask creates a new task directory structure within a project.
	CreateTask(ctx context.Context, task *Task) error

	// GetTask retrieves a task by project ID and task ID.
	GetTask(ctx context.Context, projectID ProjectID, taskID TaskID) (*Task, error)

	// ListTasks returns tasks for a project with pagination.
	ListTasks(ctx context.Context, projectID ProjectID, opts ListOptions) (*ListResult[*Task], error)

	// ListAllTasks returns tasks from all projects with pagination.
	ListAllTasks(ctx context.Context, opts ListOptions) (*ListResult[*Task], error)

	// UpdateTask updates task metadata.
	UpdateTask(ctx context.Context, task *Task) error

	// DeleteTask removes a task and all its artifacts.
	DeleteTask(ctx context.Context, projectID ProjectID, taskID TaskID) error

	// Artifact operations

	// SaveArtifact saves an artifact to a task.
	SaveArtifact(ctx context.Context, artifact *Artifact) error

	// GetArtifact retrieves an artifact by project ID, task ID, and artifact ID.
	GetArtifact(ctx context.Context, projectID ProjectID, taskID TaskID, artifactID string) (*Artifact, error)

	// ListArtifacts returns artifacts for a task with pagination.
	ListArtifacts(ctx context.Context, projectID ProjectID, taskID TaskID, opts ListOptions) (*ListResult[*Artifact], error)

	// SearchArtifacts searches artifact content across all projects/tasks or specific project/task.
	SearchArtifacts(ctx context.Context, query string, projectID *ProjectID, taskID *TaskID, opts ListOptions) (*ListResult[*Artifact], error)

	// DeleteArtifact removes an artifact.
	DeleteArtifact(ctx context.Context, projectID ProjectID, taskID TaskID, artifactID string) error

	// Close releases any resources.
	Close() error
}

// DefaultListOptions returns default pagination options.
func DefaultListOptions() ListOptions {
	return ListOptions{
		Limit:  50,
		Offset: 0,
	}
}
