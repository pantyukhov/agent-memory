package service

import (
	"bufio"
	"context"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"agent-memory/internal/domain/task"
)

// WorkspaceService provides file operations within task workspace context.
type WorkspaceService struct {
	taskRepo task.Repository
	logger   *slog.Logger
}

// NewWorkspaceService creates a new workspace service.
func NewWorkspaceService(taskRepo task.Repository, logger *slog.Logger) *WorkspaceService {
	return &WorkspaceService{
		taskRepo: taskRepo,
		logger:   logger,
	}
}

// ReadFileRequest contains parameters for reading a file.
type ReadFileRequest struct {
	ProjectID string
	TaskID    string
	FilePath  string // Relative to workspace or absolute
	LogRead   bool   // Whether to log this read as artifact
}

// ReadFileResult contains the file content and metadata.
type ReadFileResult struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	Size      int64  `json:"size"`
	LineCount int    `json:"line_count"`
}

// ReadFile reads a file within task workspace context.
func (s *WorkspaceService) ReadFile(ctx context.Context, req ReadFileRequest) (*ReadFileResult, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.TaskID)

	t, err := s.taskRepo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}

	// Get effective workspace path (task or project)
	workspacePath := t.WorkspacePath
	if workspacePath == "" {
		p, err := s.taskRepo.GetProject(ctx, projectID)
		if err != nil {
			return nil, err
		}
		workspacePath = p.WorkspacePath
	}

	// Resolve file path
	filePath := req.FilePath
	if workspacePath != "" && !filepath.IsAbs(filePath) {
		filePath = filepath.Join(workspacePath, filePath)
	}

	// Security: ensure path is within workspace
	if workspacePath != "" {
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			return nil, fmt.Errorf("invalid path: %w", err)
		}
		absWorkspace, _ := filepath.Abs(workspacePath)
		if !strings.HasPrefix(absPath, absWorkspace) {
			return nil, fmt.Errorf("path outside workspace: %s", req.FilePath)
		}
		filePath = absPath
	}

	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	stat, _ := os.Stat(filePath)
	lineCount := strings.Count(string(content), "\n")

	result := &ReadFileResult{
		Path:      filePath,
		Content:   string(content),
		Size:      stat.Size(),
		LineCount: lineCount,
	}

	// Log read as artifact if requested
	if req.LogRead {
		artifactContent := fmt.Sprintf("# File Read: %s\n\nSize: %d bytes, Lines: %d\n\n```\n%s\n```",
			filePath, stat.Size(), lineCount, truncateContent(string(content), 5000))

		artifact := task.NewArtifact(projectID, taskID, task.ArtifactTypeFileRead, artifactContent)
		artifact.Metadata["file_path"] = filePath
		artifact.Metadata["size"] = fmt.Sprintf("%d", stat.Size())

		if err := s.taskRepo.SaveArtifact(ctx, artifact); err != nil {
			s.logger.Warn("failed to log file read", "error", err)
		}
	}

	s.logger.Debug("file read", "project_id", req.ProjectID, "task_id", req.TaskID, "path", filePath)
	return result, nil
}

// ListFilesRequest contains parameters for listing files.
type ListFilesRequest struct {
	ProjectID string
	TaskID    string
	Path      string // Relative to workspace or absolute (empty = workspace root)
	Pattern   string // Glob pattern (e.g., "*.go", "**/*.ts")
	Recursive bool   // List recursively
	MaxDepth  int    // Max depth for recursive listing (0 = unlimited)
	LogList   bool   // Whether to log this listing as artifact
}

// FileInfo represents basic file information.
type FileInfo struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	IsDir   bool   `json:"is_dir"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
}

// ListFilesResult contains the directory listing.
type ListFilesResult struct {
	BasePath string     `json:"base_path"`
	Files    []FileInfo `json:"files"`
	Total    int        `json:"total"`
}

// ListFiles lists files in workspace.
func (s *WorkspaceService) ListFiles(ctx context.Context, req ListFilesRequest) (*ListFilesResult, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.TaskID)

	t, err := s.taskRepo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}

	// Get effective workspace path
	workspacePath := t.WorkspacePath
	if workspacePath == "" {
		p, err := s.taskRepo.GetProject(ctx, projectID)
		if err != nil {
			return nil, err
		}
		workspacePath = p.WorkspacePath
	}

	// Resolve base path
	basePath := req.Path
	if basePath == "" {
		basePath = workspacePath
	} else if workspacePath != "" && !filepath.IsAbs(basePath) {
		basePath = filepath.Join(workspacePath, basePath)
	}

	if basePath == "" {
		return nil, fmt.Errorf("no workspace path configured for task or project")
	}

	var files []FileInfo
	maxDepth := req.MaxDepth
	if maxDepth == 0 {
		maxDepth = 10 // Default max depth
	}

	err = filepath.WalkDir(basePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Get relative path first
		relPath, _ := filepath.Rel(basePath, path)
		if relPath == "." {
			return nil // Skip the base directory itself
		}

		// Calculate depth from relative path (number of path separators)
		currentDepth := strings.Count(relPath, string(os.PathSeparator))

		// Check depth for non-recursive mode
		if !req.Recursive && currentDepth > 0 {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if currentDepth > maxDepth {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip hidden files and common ignored directories
		name := d.Name()
		if strings.HasPrefix(name, ".") || name == "node_modules" || name == "vendor" || name == "__pycache__" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Apply pattern filter
		if req.Pattern != "" && !d.IsDir() {
			matched, _ := filepath.Match(req.Pattern, name)
			if !matched {
				return nil
			}
		}

		info, _ := d.Info()

		var size int64
		var modTime string
		if info != nil {
			size = info.Size()
			modTime = info.ModTime().Format("2006-01-02 15:04:05")
		}

		files = append(files, FileInfo{
			Path:    relPath,
			Name:    name,
			IsDir:   d.IsDir(),
			Size:    size,
			ModTime: modTime,
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	result := &ListFilesResult{
		BasePath: basePath,
		Files:    files,
		Total:    len(files),
	}

	// Log listing as artifact if requested
	if req.LogList {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Directory Listing: %s\n\n", basePath))
		sb.WriteString(fmt.Sprintf("Total: %d items\n\n", len(files)))
		for _, f := range files {
			if f.IsDir {
				sb.WriteString(fmt.Sprintf("- [dir] %s/\n", f.Path))
			} else {
				sb.WriteString(fmt.Sprintf("- [file] %s (%d bytes)\n", f.Path, f.Size))
			}
		}

		artifact := task.NewArtifact(projectID, taskID, task.ArtifactTypeFileList, sb.String())
		artifact.Metadata["base_path"] = basePath
		artifact.Metadata["pattern"] = req.Pattern
		artifact.Metadata["total"] = fmt.Sprintf("%d", len(files))

		if err := s.taskRepo.SaveArtifact(ctx, artifact); err != nil {
			s.logger.Warn("failed to log file list", "error", err)
		}
	}

	s.logger.Debug("files listed", "project_id", req.ProjectID, "task_id", req.TaskID, "path", basePath, "count", len(files))
	return result, nil
}

// SearchFilesRequest contains parameters for searching files.
type SearchFilesRequest struct {
	ProjectID  string
	TaskID     string
	Query      string // Text to search for
	Pattern    string // File pattern (e.g., "*.go")
	IgnoreCase bool   // Case-insensitive search
	MaxResults int    // Limit results
	LogSearch  bool   // Whether to log this search as artifact
}

// SearchMatch represents a search match.
type SearchMatch struct {
	FilePath   string `json:"file_path"`
	LineNumber int    `json:"line_number"`
	Line       string `json:"line"`
}

// SearchFilesResult contains search results.
type SearchFilesResult struct {
	Query   string        `json:"query"`
	Matches []SearchMatch `json:"matches"`
	Total   int           `json:"total"`
}

// SearchFiles searches for text in workspace files.
func (s *WorkspaceService) SearchFiles(ctx context.Context, req SearchFilesRequest) (*SearchFilesResult, error) {
	projectID := task.NewProjectID(req.ProjectID)
	taskID := task.NewTaskID(req.TaskID)

	t, err := s.taskRepo.GetTask(ctx, projectID, taskID)
	if err != nil {
		return nil, err
	}

	// Get effective workspace path
	workspacePath := t.WorkspacePath
	if workspacePath == "" {
		p, err := s.taskRepo.GetProject(ctx, projectID)
		if err != nil {
			return nil, err
		}
		workspacePath = p.WorkspacePath
	}

	if workspacePath == "" {
		return nil, fmt.Errorf("no workspace path configured for task or project")
	}

	maxResults := req.MaxResults
	if maxResults == 0 {
		maxResults = 100
	}

	query := req.Query
	if req.IgnoreCase {
		query = strings.ToLower(query)
	}

	var matches []SearchMatch

	err = filepath.WalkDir(workspacePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		// Skip hidden and ignored
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			return nil
		}

		// Apply pattern
		if req.Pattern != "" {
			matched, _ := filepath.Match(req.Pattern, name)
			if !matched {
				return nil
			}
		}

		// Skip binary files (simple heuristic)
		ext := strings.ToLower(filepath.Ext(name))
		if isBinaryExtension(ext) {
			return nil
		}

		// Search in file
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		relPath, _ := filepath.Rel(workspacePath, path)

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			searchLine := line
			if req.IgnoreCase {
				searchLine = strings.ToLower(line)
			}

			if strings.Contains(searchLine, query) {
				matches = append(matches, SearchMatch{
					FilePath:   relPath,
					LineNumber: lineNum,
					Line:       truncateLine(line, 200),
				})

				if len(matches) >= maxResults {
					return filepath.SkipAll
				}
			}
		}

		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	result := &SearchFilesResult{
		Query:   req.Query,
		Matches: matches,
		Total:   len(matches),
	}

	// Log search as artifact if requested
	if req.LogSearch {
		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("# Search: \"%s\"\n\n", req.Query))
		sb.WriteString(fmt.Sprintf("Pattern: %s, Results: %d\n\n", req.Pattern, len(matches)))
		for _, m := range matches {
			sb.WriteString(fmt.Sprintf("- **%s:%d** `%s`\n", m.FilePath, m.LineNumber, m.Line))
		}

		artifact := task.NewArtifact(projectID, taskID, task.ArtifactTypeSearch, sb.String())
		artifact.Metadata["query"] = req.Query
		artifact.Metadata["pattern"] = req.Pattern
		artifact.Metadata["results"] = fmt.Sprintf("%d", len(matches))

		if err := s.taskRepo.SaveArtifact(ctx, artifact); err != nil {
			s.logger.Warn("failed to log search", "error", err)
		}
	}

	s.logger.Debug("search completed", "project_id", req.ProjectID, "task_id", req.TaskID, "query", req.Query, "matches", len(matches))
	return result, nil
}

// Helper functions

func truncateContent(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "\n... (truncated)"
}

func truncateLine(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func isBinaryExtension(ext string) bool {
	binary := map[string]bool{
		".exe": true, ".dll": true, ".so": true, ".dylib": true,
		".zip": true, ".tar": true, ".gz": true, ".rar": true,
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".ico": true,
		".pdf": true, ".doc": true, ".docx": true,
		".bin": true, ".dat": true, ".db": true, ".sqlite": true,
	}
	return binary[ext]
}
