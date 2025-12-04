package task

import (
	"context"
)

// Repository defines the contract for project, task, and artifact storage.
type Repository interface {
	// Project operations

	// CreateProject creates a new project directory structure.
	CreateProject(ctx context.Context, project *Project) error

	// GetProject retrieves a project by ID.
	GetProject(ctx context.Context, id ProjectID) (*Project, error)

	// ListProjects returns all projects.
	ListProjects(ctx context.Context) ([]*Project, error)

	// UpdateProject updates project metadata.
	UpdateProject(ctx context.Context, project *Project) error

	// DeleteProject removes a project and all its tasks.
	DeleteProject(ctx context.Context, id ProjectID) error

	// Task operations

	// CreateTask creates a new task directory structure within a project.
	CreateTask(ctx context.Context, task *Task) error

	// GetTask retrieves a task by project ID and task ID.
	GetTask(ctx context.Context, projectID ProjectID, taskID TaskID) (*Task, error)

	// ListTasks returns all tasks for a project.
	ListTasks(ctx context.Context, projectID ProjectID) ([]*Task, error)

	// UpdateTask updates task metadata.
	UpdateTask(ctx context.Context, task *Task) error

	// DeleteTask removes a task and all its artifacts.
	DeleteTask(ctx context.Context, projectID ProjectID, taskID TaskID) error

	// Artifact operations

	// SaveArtifact saves an artifact to a task.
	SaveArtifact(ctx context.Context, artifact *Artifact) error

	// GetArtifact retrieves an artifact by project ID, task ID, and artifact ID.
	GetArtifact(ctx context.Context, projectID ProjectID, taskID TaskID, artifactID string) (*Artifact, error)

	// ListArtifacts returns all artifacts for a task.
	ListArtifacts(ctx context.Context, projectID ProjectID, taskID TaskID) ([]*Artifact, error)

	// SearchArtifacts searches artifact content across all projects/tasks or specific project/task.
	SearchArtifacts(ctx context.Context, query string, projectID *ProjectID, taskID *TaskID) ([]*Artifact, error)

	// DeleteArtifact removes an artifact.
	DeleteArtifact(ctx context.Context, projectID ProjectID, taskID TaskID, artifactID string) error

	// Close releases any resources.
	Close() error
}
