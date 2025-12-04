package task

import (
	"testing"
)

func TestDefaultListOptions(t *testing.T) {
	opts := DefaultListOptions()

	if opts.Limit != 50 {
		t.Errorf("DefaultListOptions().Limit = %d, want 50", opts.Limit)
	}
	if opts.Offset != 0 {
		t.Errorf("DefaultListOptions().Offset = %d, want 0", opts.Offset)
	}
	if opts.Status != "" {
		t.Errorf("DefaultListOptions().Status = %q, want empty", opts.Status)
	}
}

func TestListOptions(t *testing.T) {
	opts := ListOptions{
		Limit:  10,
		Offset: 20,
		Status: TaskStatusCompleted,
	}

	if opts.Limit != 10 {
		t.Errorf("ListOptions.Limit = %d, want 10", opts.Limit)
	}
	if opts.Offset != 20 {
		t.Errorf("ListOptions.Offset = %d, want 20", opts.Offset)
	}
	if opts.Status != TaskStatusCompleted {
		t.Errorf("ListOptions.Status = %q, want %q", opts.Status, TaskStatusCompleted)
	}
}

func TestListResult(t *testing.T) {
	items := []string{"a", "b", "c"}
	result := ListResult[string]{
		Items:   items,
		Total:   100,
		Limit:   10,
		Offset:  0,
		HasMore: true,
	}

	if len(result.Items) != 3 {
		t.Errorf("ListResult.Items length = %d, want 3", len(result.Items))
	}
	if result.Total != 100 {
		t.Errorf("ListResult.Total = %d, want 100", result.Total)
	}
	if result.Limit != 10 {
		t.Errorf("ListResult.Limit = %d, want 10", result.Limit)
	}
	if result.Offset != 0 {
		t.Errorf("ListResult.Offset = %d, want 0", result.Offset)
	}
	if !result.HasMore {
		t.Error("ListResult.HasMore should be true")
	}
}
