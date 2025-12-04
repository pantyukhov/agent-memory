package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"

	"agent-memory/internal/application/service"
	"agent-memory/internal/domain/task"
)

// Project handlers

func (s *Server) handleCreateProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")
	name := request.GetString("name", "")
	description := request.GetString("description", "")
	workspacePath := request.GetString("workspace_path", "")

	req := service.CreateProjectRequest{
		ID:            id,
		Name:          name,
		Description:   description,
		WorkspacePath: workspacePath,
	}

	// Parse metadata
	args := request.GetArguments()
	if metaRaw, ok := args["metadata"].(map[string]interface{}); ok {
		meta := make(map[string]string, len(metaRaw))
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		req.Metadata = meta
	}

	p, err := s.taskService.CreateProject(ctx, req)
	if err != nil {
		if err == task.ErrProjectAlreadyExists {
			return errorResult(fmt.Sprintf("Project '%s' already exists", id)), nil
		}
		if err == task.ErrInvalidProjectID {
			return errorResult(fmt.Sprintf("Invalid project ID '%s'. Use lowercase letters, numbers, and dashes.", id)), nil
		}
		return errorResult(fmt.Sprintf("Failed to create project: %v", err)), nil
	}

	response := projectToMap(p)
	response["message"] = fmt.Sprintf("Project '%s' created successfully", p.ID)

	return jsonResult(response)
}

func (s *Server) handleGetProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")

	p, err := s.taskService.GetProject(ctx, id)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", id)), nil
		}
		return errorResult(fmt.Sprintf("Failed to get project: %v", err)), nil
	}

	return jsonResult(projectToMap(p))
}

func (s *Server) handleListProjects(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		if limitVal, ok := limitRaw.(float64); ok {
			limit = int(limitVal)
		}
	}

	offset := 0
	if offsetRaw, ok := args["offset"]; ok {
		if offsetVal, ok := offsetRaw.(float64); ok {
			offset = int(offsetVal)
		}
	}

	req := service.ListProjectsRequest{
		Limit:  limit,
		Offset: offset,
	}

	result, err := s.taskService.ListProjects(ctx, req)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to list projects: %v", err)), nil
	}

	projectMaps := make([]map[string]interface{}, 0, len(result.Items))
	for _, p := range result.Items {
		projectMaps = append(projectMaps, projectToMap(p))
	}

	response := map[string]interface{}{
		"projects": projectMaps,
		"total":    result.Total,
		"limit":    result.Limit,
		"offset":   result.Offset,
		"has_more": result.HasMore,
	}

	return jsonResult(response)
}

func (s *Server) handleUpdateProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")

	req := service.UpdateProjectRequest{
		ID: id,
	}

	args := request.GetArguments()

	if nameRaw, ok := args["name"]; ok {
		if name, ok := nameRaw.(string); ok {
			req.Name = &name
		}
	}

	if descRaw, ok := args["description"]; ok {
		if desc, ok := descRaw.(string); ok {
			req.Description = &desc
		}
	}

	if wsRaw, ok := args["workspace_path"]; ok {
		if ws, ok := wsRaw.(string); ok {
			req.WorkspacePath = &ws
		}
	}

	if metaRaw, ok := args["metadata"].(map[string]interface{}); ok {
		meta := make(map[string]string, len(metaRaw))
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		req.Metadata = meta
	}

	p, err := s.taskService.UpdateProject(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", id)), nil
		}
		return errorResult(fmt.Sprintf("Failed to update project: %v", err)), nil
	}

	response := projectToMap(p)
	response["message"] = fmt.Sprintf("Project '%s' updated successfully", p.ID)

	return jsonResult(response)
}

func (s *Server) handleDeleteProject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id := request.GetString("id", "")

	err := s.taskService.DeleteProject(ctx, id)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", id)), nil
		}
		return errorResult(fmt.Sprintf("Failed to delete project: %v", err)), nil
	}

	response := map[string]interface{}{
		"id":      id,
		"message": fmt.Sprintf("Project '%s' and all its tasks deleted successfully", id),
	}

	return jsonResult(response)
}

// Task handlers

func (s *Server) handleCreateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	id := request.GetString("id", "")
	name := request.GetString("name", "")
	description := request.GetString("description", "")
	workspacePath := request.GetString("workspace_path", "")

	req := service.CreateTaskRequest{
		ProjectID:     projectID,
		ID:            id,
		Name:          name,
		Description:   description,
		WorkspacePath: workspacePath,
	}

	// Parse metadata
	args := request.GetArguments()
	if metaRaw, ok := args["metadata"].(map[string]interface{}); ok {
		meta := make(map[string]string, len(metaRaw))
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		req.Metadata = meta
	}

	t, err := s.taskService.CreateTask(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found. Create the project first.", projectID)), nil
		}
		if err == task.ErrTaskAlreadyExists {
			return errorResult(fmt.Sprintf("Task '%s' already exists in project '%s'", id, projectID)), nil
		}
		if err == task.ErrInvalidTaskID {
			return errorResult(fmt.Sprintf("Invalid task ID '%s'. Use lowercase letters, numbers, and dashes.", id)), nil
		}
		return errorResult(fmt.Sprintf("Failed to create task: %v", err)), nil
	}

	response := taskToMap(t)
	response["message"] = fmt.Sprintf("Task '%s' created in project '%s'", t.ID, t.ProjectID)

	return jsonResult(response)
}

func (s *Server) handleGetTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")

	t, err := s.taskService.GetTask(ctx, projectID, taskID)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to get task: %v", err)), nil
	}

	return jsonResult(taskToMap(t))
}

func (s *Server) handleListTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	args := request.GetArguments()

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		if limitVal, ok := limitRaw.(float64); ok {
			limit = int(limitVal)
		}
	}

	offset := 0
	if offsetRaw, ok := args["offset"]; ok {
		if offsetVal, ok := offsetRaw.(float64); ok {
			offset = int(offsetVal)
		}
	}

	var status task.TaskStatus
	if statusRaw, ok := args["status"]; ok {
		if statusStr, ok := statusRaw.(string); ok {
			status = task.TaskStatus(statusStr)
		}
	}

	req := service.ListTasksRequest{
		ProjectID: projectID,
		Limit:     limit,
		Offset:    offset,
		Status:    status,
	}

	result, err := s.taskService.ListTasks(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to list tasks: %v", err)), nil
	}

	taskMaps := make([]map[string]interface{}, 0, len(result.Items))
	for _, t := range result.Items {
		taskMaps = append(taskMaps, taskToMap(t))
	}

	response := map[string]interface{}{
		"project_id": projectID,
		"tasks":      taskMaps,
		"total":      result.Total,
		"limit":      result.Limit,
		"offset":     result.Offset,
		"has_more":   result.HasMore,
	}

	return jsonResult(response)
}

func (s *Server) handleListAllTasks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := request.GetArguments()

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		if limitVal, ok := limitRaw.(float64); ok {
			limit = int(limitVal)
		}
	}

	offset := 0
	if offsetRaw, ok := args["offset"]; ok {
		if offsetVal, ok := offsetRaw.(float64); ok {
			offset = int(offsetVal)
		}
	}

	var status task.TaskStatus
	if statusRaw, ok := args["status"]; ok {
		if statusStr, ok := statusRaw.(string); ok {
			status = task.TaskStatus(statusStr)
		}
	}

	req := service.ListAllTasksRequest{
		Limit:  limit,
		Offset: offset,
		Status: status,
	}

	result, err := s.taskService.ListAllTasks(ctx, req)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to list all tasks: %v", err)), nil
	}

	taskMaps := make([]map[string]interface{}, 0, len(result.Items))
	for _, t := range result.Items {
		taskMaps = append(taskMaps, taskToMap(t))
	}

	response := map[string]interface{}{
		"tasks":    taskMaps,
		"total":    result.Total,
		"limit":    result.Limit,
		"offset":   result.Offset,
		"has_more": result.HasMore,
	}

	return jsonResult(response)
}

func (s *Server) handleUpdateTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")

	req := service.UpdateTaskRequest{
		ProjectID: projectID,
		ID:        taskID,
	}

	args := request.GetArguments()

	if nameRaw, ok := args["name"]; ok {
		if name, ok := nameRaw.(string); ok {
			req.Name = &name
		}
	}

	if descRaw, ok := args["description"]; ok {
		if desc, ok := descRaw.(string); ok {
			req.Description = &desc
		}
	}

	if wsRaw, ok := args["workspace_path"]; ok {
		if ws, ok := wsRaw.(string); ok {
			req.WorkspacePath = &ws
		}
	}

	if statusRaw, ok := args["status"]; ok {
		if statusStr, ok := statusRaw.(string); ok {
			status := task.TaskStatus(statusStr)
			req.Status = &status
		}
	}

	if metaRaw, ok := args["metadata"].(map[string]interface{}); ok {
		meta := make(map[string]string, len(metaRaw))
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		req.Metadata = meta
	}

	t, err := s.taskService.UpdateTask(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to update task: %v", err)), nil
	}

	response := taskToMap(t)
	response["message"] = fmt.Sprintf("Task '%s' updated successfully", t.ID)

	return jsonResult(response)
}

func (s *Server) handleDeleteTask(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")

	err := s.taskService.DeleteTask(ctx, projectID, taskID)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to delete task: %v", err)), nil
	}

	response := map[string]interface{}{
		"project_id": projectID,
		"task_id":    taskID,
		"message":    fmt.Sprintf("Task '%s' and all its artifacts deleted successfully", taskID),
	}

	return jsonResult(response)
}

// Artifact handlers

func (s *Server) handleSaveArtifact(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	content := request.GetString("content", "")
	artifactType := request.GetString("type", "artifact")

	req := service.SaveArtifactRequest{
		ProjectID: projectID,
		TaskID:    taskID,
		Content:   content,
		Type:      task.ArtifactType(artifactType),
	}

	// Parse metadata
	args := request.GetArguments()
	if metaRaw, ok := args["metadata"].(map[string]interface{}); ok {
		meta := make(map[string]string, len(metaRaw))
		for k, v := range metaRaw {
			if str, ok := v.(string); ok {
				meta[k] = str
			}
		}
		req.Metadata = meta
	}

	a, err := s.taskService.SaveArtifact(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'. Create the task first.", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to save artifact: %v", err)), nil
	}

	response := artifactToMap(a)
	response["message"] = fmt.Sprintf("Artifact saved to task '%s' as %s", taskID, a.Filename())

	return jsonResult(response)
}

func (s *Server) handleGetArtifact(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	artifactID := request.GetString("artifact_id", "")

	a, err := s.taskService.GetArtifact(ctx, projectID, taskID, artifactID)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		if err == task.ErrArtifactNotFound {
			return errorResult(fmt.Sprintf("Artifact '%s' not found in task '%s'", artifactID, taskID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to get artifact: %v", err)), nil
	}

	return jsonResult(artifactToMap(a))
}

func (s *Server) handleListArtifacts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	args := request.GetArguments()

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		if limitVal, ok := limitRaw.(float64); ok {
			limit = int(limitVal)
		}
	}

	offset := 0
	if offsetRaw, ok := args["offset"]; ok {
		if offsetVal, ok := offsetRaw.(float64); ok {
			offset = int(offsetVal)
		}
	}

	req := service.ListArtifactsRequest{
		ProjectID: projectID,
		TaskID:    taskID,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := s.taskService.ListArtifacts(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to list artifacts: %v", err)), nil
	}

	artifactMaps := make([]map[string]interface{}, 0, len(result.Items))
	for _, a := range result.Items {
		// Include preview of content
		preview := a.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		m := artifactToMap(a)
		m["content_preview"] = preview
		artifactMaps = append(artifactMaps, m)
	}

	response := map[string]interface{}{
		"project_id": projectID,
		"task_id":    taskID,
		"artifacts":  artifactMaps,
		"total":      result.Total,
		"limit":      result.Limit,
		"offset":     result.Offset,
		"has_more":   result.HasMore,
	}

	return jsonResult(response)
}

func (s *Server) handleSearchArtifacts(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := request.GetString("query", "")
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	args := request.GetArguments()

	limit := 0
	if limitRaw, ok := args["limit"]; ok {
		if limitVal, ok := limitRaw.(float64); ok {
			limit = int(limitVal)
		}
	}

	offset := 0
	if offsetRaw, ok := args["offset"]; ok {
		if offsetVal, ok := offsetRaw.(float64); ok {
			offset = int(offsetVal)
		}
	}

	req := service.SearchArtifactsRequest{
		Query:     query,
		ProjectID: projectID,
		TaskID:    taskID,
		Limit:     limit,
		Offset:    offset,
	}

	result, err := s.taskService.SearchArtifacts(ctx, req)
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to search artifacts: %v", err)), nil
	}

	artifactMaps := make([]map[string]interface{}, 0, len(result.Items))
	for _, a := range result.Items {
		// Include preview of content
		preview := a.Content
		if len(preview) > 200 {
			preview = preview[:200] + "..."
		}
		m := artifactToMap(a)
		m["content_preview"] = preview
		artifactMaps = append(artifactMaps, m)
	}

	response := map[string]interface{}{
		"query":     query,
		"artifacts": artifactMaps,
		"total":     result.Total,
		"limit":     result.Limit,
		"offset":    result.Offset,
		"has_more":  result.HasMore,
	}

	return jsonResult(response)
}

func (s *Server) handleDeleteArtifact(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	artifactID := request.GetString("artifact_id", "")

	err := s.taskService.DeleteArtifact(ctx, projectID, taskID, artifactID)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		if err == task.ErrArtifactNotFound {
			return errorResult(fmt.Sprintf("Artifact '%s' not found in task '%s'", artifactID, taskID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to delete artifact: %v", err)), nil
	}

	response := map[string]interface{}{
		"project_id":  projectID,
		"task_id":     taskID,
		"artifact_id": artifactID,
		"message":     fmt.Sprintf("Artifact '%s' deleted from task '%s'", artifactID, taskID),
	}

	return jsonResult(response)
}

// File operation handlers

func (s *Server) handleReadFile(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	filePath := request.GetString("file_path", "")

	// Default log_read to true
	logRead := true
	args := request.GetArguments()
	if logRaw, ok := args["log_read"]; ok {
		if logVal, ok := logRaw.(bool); ok {
			logRead = logVal
		}
	}

	req := service.ReadFileRequest{
		ProjectID: projectID,
		TaskID:    taskID,
		FilePath:  filePath,
		LogRead:   logRead,
	}

	result, err := s.workspaceService.ReadFile(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to read file: %v", err)), nil
	}

	response := map[string]interface{}{
		"path":       result.Path,
		"content":    result.Content,
		"size":       result.Size,
		"line_count": result.LineCount,
	}

	return jsonResult(response)
}

func (s *Server) handleListFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	path := request.GetString("path", "")
	pattern := request.GetString("pattern", "")

	args := request.GetArguments()

	recursive := false
	if recRaw, ok := args["recursive"]; ok {
		if recVal, ok := recRaw.(bool); ok {
			recursive = recVal
		}
	}

	maxDepth := 0
	if depthRaw, ok := args["max_depth"]; ok {
		if depthVal, ok := depthRaw.(float64); ok {
			maxDepth = int(depthVal)
		}
	}

	logList := false
	if logRaw, ok := args["log_list"]; ok {
		if logVal, ok := logRaw.(bool); ok {
			logList = logVal
		}
	}

	req := service.ListFilesRequest{
		ProjectID: projectID,
		TaskID:    taskID,
		Path:      path,
		Pattern:   pattern,
		Recursive: recursive,
		MaxDepth:  maxDepth,
		LogList:   logList,
	}

	result, err := s.workspaceService.ListFiles(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to list files: %v", err)), nil
	}

	fileInfos := make([]map[string]interface{}, 0, len(result.Files))
	for _, f := range result.Files {
		fileInfos = append(fileInfos, map[string]interface{}{
			"path":     f.Path,
			"name":     f.Name,
			"is_dir":   f.IsDir,
			"size":     f.Size,
			"mod_time": f.ModTime,
		})
	}

	response := map[string]interface{}{
		"base_path": result.BasePath,
		"files":     fileInfos,
		"total":     result.Total,
	}

	return jsonResult(response)
}

func (s *Server) handleSearchFiles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	projectID := request.GetString("project_id", "")
	taskID := request.GetString("task_id", "")
	query := request.GetString("query", "")
	pattern := request.GetString("pattern", "")

	args := request.GetArguments()

	ignoreCase := false
	if icRaw, ok := args["ignore_case"]; ok {
		if icVal, ok := icRaw.(bool); ok {
			ignoreCase = icVal
		}
	}

	maxResults := 0
	if mrRaw, ok := args["max_results"]; ok {
		if mrVal, ok := mrRaw.(float64); ok {
			maxResults = int(mrVal)
		}
	}

	// Default log_search to true
	logSearch := true
	if logRaw, ok := args["log_search"]; ok {
		if logVal, ok := logRaw.(bool); ok {
			logSearch = logVal
		}
	}

	req := service.SearchFilesRequest{
		ProjectID:  projectID,
		TaskID:     taskID,
		Query:      query,
		Pattern:    pattern,
		IgnoreCase: ignoreCase,
		MaxResults: maxResults,
		LogSearch:  logSearch,
	}

	result, err := s.workspaceService.SearchFiles(ctx, req)
	if err != nil {
		if err == task.ErrProjectNotFound {
			return errorResult(fmt.Sprintf("Project '%s' not found", projectID)), nil
		}
		if err == task.ErrTaskNotFound {
			return errorResult(fmt.Sprintf("Task '%s' not found in project '%s'", taskID, projectID)), nil
		}
		return errorResult(fmt.Sprintf("Failed to search files: %v", err)), nil
	}

	matches := make([]map[string]interface{}, 0, len(result.Matches))
	for _, m := range result.Matches {
		matches = append(matches, map[string]interface{}{
			"file_path":   m.FilePath,
			"line_number": m.LineNumber,
			"line":        m.Line,
		})
	}

	response := map[string]interface{}{
		"query":   result.Query,
		"matches": matches,
		"total":   result.Total,
	}

	return jsonResult(response)
}

// Helper functions

func projectToMap(p *task.Project) map[string]interface{} {
	return map[string]interface{}{
		"id":             p.ID,
		"name":           p.Name,
		"description":    p.Description,
		"workspace_path": p.WorkspacePath,
		"metadata":       p.Metadata,
		"created_at":     p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":     p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func taskToMap(t *task.Task) map[string]interface{} {
	return map[string]interface{}{
		"id":             t.ID,
		"project_id":     t.ProjectID,
		"name":           t.Name,
		"description":    t.Description,
		"status":         t.Status,
		"workspace_path": t.WorkspacePath,
		"metadata":       t.Metadata,
		"created_at":     t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		"updated_at":     t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func artifactToMap(a *task.Artifact) map[string]interface{} {
	return map[string]interface{}{
		"id":         a.ID,
		"project_id": a.ProjectID,
		"task_id":    a.TaskID,
		"type":       a.Type,
		"content":    a.Content,
		"metadata":   a.Metadata,
		"filename":   a.Filename(),
		"created_at": a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func jsonResult(data interface{}) (*mcp.CallToolResult, error) {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errorResult(fmt.Sprintf("Failed to marshal response: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonBytes)), nil
}

func errorResult(message string) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: message,
			},
		},
		IsError: true,
	}
}
