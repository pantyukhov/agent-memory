package task

import (
	"testing"
	"time"
)

func TestNewProjectID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ProjectID
	}{
		{
			name:     "simple lowercase",
			input:    "myproject",
			expected: ProjectID("myproject"),
		},
		{
			name:     "uppercase converted to lowercase",
			input:    "MyProject",
			expected: ProjectID("myproject"),
		},
		{
			name:     "spaces converted to dashes",
			input:    "my project",
			expected: ProjectID("my-project"),
		},
		{
			name:     "special chars converted to dashes",
			input:    "my_project!@#",
			expected: ProjectID("my-project"),
		},
		{
			name:     "multiple dashes collapsed",
			input:    "my---project",
			expected: ProjectID("my-project"),
		},
		{
			name:     "trimmed dashes",
			input:    "-my-project-",
			expected: ProjectID("my-project"),
		},
		{
			name:     "with numbers",
			input:    "project123",
			expected: ProjectID("project123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewProjectID(tt.input)
			if got != tt.expected {
				t.Errorf("NewProjectID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestProjectID_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		id       ProjectID
		expected bool
	}{
		{
			name:     "valid simple",
			id:       ProjectID("myproject"),
			expected: true,
		},
		{
			name:     "valid with dash",
			id:       ProjectID("my-project"),
			expected: true,
		},
		{
			name:     "valid single char",
			id:       ProjectID("a"),
			expected: true,
		},
		{
			name:     "valid with numbers",
			id:       ProjectID("project123"),
			expected: true,
		},
		{
			name:     "empty string",
			id:       ProjectID(""),
			expected: false,
		},
		{
			name:     "starts with dash",
			id:       ProjectID("-project"),
			expected: false,
		},
		{
			name:     "ends with dash",
			id:       ProjectID("project-"),
			expected: false,
		},
		{
			name:     "uppercase",
			id:       ProjectID("MyProject"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.IsValid()
			if got != tt.expected {
				t.Errorf("ProjectID(%q).IsValid() = %v, want %v", tt.id, got, tt.expected)
			}
		})
	}
}

func TestNewProject(t *testing.T) {
	id := ProjectID("test-project")
	name := "Test Project"

	p := NewProject(id, name)

	if p.ID != id {
		t.Errorf("NewProject().ID = %q, want %q", p.ID, id)
	}
	if p.Name != name {
		t.Errorf("NewProject().Name = %q, want %q", p.Name, name)
	}
	if p.Metadata == nil {
		t.Error("NewProject().Metadata should not be nil")
	}
	if p.CreatedAt.IsZero() {
		t.Error("NewProject().CreatedAt should not be zero")
	}
	if p.UpdatedAt.IsZero() {
		t.Error("NewProject().UpdatedAt should not be zero")
	}
}

func TestNewTaskID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected TaskID
	}{
		{
			name:     "simple lowercase",
			input:    "fix-bug",
			expected: TaskID("fix-bug"),
		},
		{
			name:     "uppercase converted",
			input:    "Fix-Bug",
			expected: TaskID("fix-bug"),
		},
		{
			name:     "spaces converted",
			input:    "fix login bug",
			expected: TaskID("fix-login-bug"),
		},
		{
			name:     "special chars removed",
			input:    "bug#123",
			expected: TaskID("bug-123"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewTaskID(tt.input)
			if got != tt.expected {
				t.Errorf("NewTaskID(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTaskID_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		id       TaskID
		expected bool
	}{
		{
			name:     "valid simple",
			id:       TaskID("fix-bug"),
			expected: true,
		},
		{
			name:     "valid single char",
			id:       TaskID("a"),
			expected: true,
		},
		{
			name:     "empty string",
			id:       TaskID(""),
			expected: false,
		},
		{
			name:     "starts with dash",
			id:       TaskID("-fix"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.id.IsValid()
			if got != tt.expected {
				t.Errorf("TaskID(%q).IsValid() = %v, want %v", tt.id, got, tt.expected)
			}
		})
	}
}

func TestNewTask(t *testing.T) {
	projectID := ProjectID("test-project")
	taskID := TaskID("fix-bug")
	name := "Fix Login Bug"

	task := NewTask(projectID, taskID, name)

	if task.ID != taskID {
		t.Errorf("NewTask().ID = %q, want %q", task.ID, taskID)
	}
	if task.ProjectID != projectID {
		t.Errorf("NewTask().ProjectID = %q, want %q", task.ProjectID, projectID)
	}
	if task.Name != name {
		t.Errorf("NewTask().Name = %q, want %q", task.Name, name)
	}
	if task.Status != TaskStatusOpen {
		t.Errorf("NewTask().Status = %q, want %q", task.Status, TaskStatusOpen)
	}
	if task.Metadata == nil {
		t.Error("NewTask().Metadata should not be nil")
	}
}

func TestTask_DirName(t *testing.T) {
	tests := []struct {
		name     string
		task     *Task
		expected string
	}{
		{
			name: "open status",
			task: &Task{
				ID:     TaskID("fix-bug"),
				Status: TaskStatusOpen,
			},
			expected: "[open]-fix-bug",
		},
		{
			name: "in progress status",
			task: &Task{
				ID:     TaskID("add-feature"),
				Status: TaskStatusInProgress,
			},
			expected: "[in_progress]-add-feature",
		},
		{
			name: "completed status",
			task: &Task{
				ID:     TaskID("done-task"),
				Status: TaskStatusCompleted,
			},
			expected: "[completed]-done-task",
		},
		{
			name: "archived status",
			task: &Task{
				ID:     TaskID("old-task"),
				Status: TaskStatusArchived,
			},
			expected: "[archived]-old-task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.task.DirName()
			if got != tt.expected {
				t.Errorf("Task.DirName() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTaskStatus_StatusEmoji(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected string
	}{
		{TaskStatusOpen, "📋"},
		{TaskStatusInProgress, "🔄"},
		{TaskStatusCompleted, "✅"},
		{TaskStatusArchived, "📦"},
		{TaskStatus("unknown"), "📋"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			got := tt.status.StatusEmoji()
			if got != tt.expected {
				t.Errorf("TaskStatus(%q).StatusEmoji() = %q, want %q", tt.status, got, tt.expected)
			}
		})
	}
}

func TestNewArtifact(t *testing.T) {
	projectID := ProjectID("test-project")
	taskID := TaskID("fix-bug")
	artifactType := ArtifactTypeNote
	content := "This is a note"

	artifact := NewArtifact(projectID, taskID, artifactType, content)

	if artifact.ProjectID != projectID {
		t.Errorf("NewArtifact().ProjectID = %q, want %q", artifact.ProjectID, projectID)
	}
	if artifact.TaskID != taskID {
		t.Errorf("NewArtifact().TaskID = %q, want %q", artifact.TaskID, taskID)
	}
	if artifact.Type != artifactType {
		t.Errorf("NewArtifact().Type = %q, want %q", artifact.Type, artifactType)
	}
	if artifact.Content != content {
		t.Errorf("NewArtifact().Content = %q, want %q", artifact.Content, content)
	}
	if artifact.ID == "" {
		t.Error("NewArtifact().ID should not be empty")
	}
	if artifact.Metadata == nil {
		t.Error("NewArtifact().Metadata should not be nil")
	}
}

func TestArtifact_Filename(t *testing.T) {
	timestamp := time.Unix(1733312000, 0)
	artifact := &Artifact{
		Type:      ArtifactTypeNote,
		CreatedAt: timestamp,
	}

	expected := "note.1733312000.md"
	got := artifact.Filename()

	if got != expected {
		t.Errorf("Artifact.Filename() = %q, want %q", got, expected)
	}
}

func TestArtifactTypes(t *testing.T) {
	tests := []struct {
		artifactType ArtifactType
		expected     string
	}{
		{ArtifactTypeNote, "note"},
		{ArtifactTypeCode, "code"},
		{ArtifactTypeDecision, "decision"},
		{ArtifactTypeDiscussion, "discussion"},
		{ArtifactTypeReference, "reference"},
		{ArtifactTypeFileRead, "file_read"},
		{ArtifactTypeFileList, "file_list"},
		{ArtifactTypeSearch, "search"},
		{ArtifactTypeGeneric, "artifact"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if string(tt.artifactType) != tt.expected {
				t.Errorf("ArtifactType = %q, want %q", tt.artifactType, tt.expected)
			}
		})
	}
}
