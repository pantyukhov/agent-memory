package task

import "errors"

var (
	// ErrProjectNotFound indicates the project was not found.
	ErrProjectNotFound = errors.New("project not found")

	// ErrProjectAlreadyExists indicates a project with the given ID already exists.
	ErrProjectAlreadyExists = errors.New("project already exists")

	// ErrInvalidProjectID indicates the provided project ID is invalid.
	ErrInvalidProjectID = errors.New("invalid project ID")

	// ErrTaskNotFound indicates the task was not found.
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskAlreadyExists indicates a task with the given ID already exists.
	ErrTaskAlreadyExists = errors.New("task already exists")

	// ErrArtifactNotFound indicates the artifact was not found.
	ErrArtifactNotFound = errors.New("artifact not found")

	// ErrInvalidTaskID indicates the provided task ID is invalid.
	ErrInvalidTaskID = errors.New("invalid task ID")

	// ErrStorageFailed indicates a storage operation failed.
	ErrStorageFailed = errors.New("storage operation failed")
)
