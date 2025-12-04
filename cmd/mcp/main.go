package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"agent-memory/internal/application/service"
	"agent-memory/internal/infrastructure/storage/filesystem"
	mcptransport "agent-memory/internal/transport/mcp"
)

func main() {
	// Parse command line flags
	tasksPath := flag.String("tasks-path", "", "Path to tasks directory (default: ~/.agent-memory/tasks)")
	logLevel := flag.String("log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	// Setup logger
	var level slog.Level
	switch *logLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))

	// Determine tasks path
	path := *tasksPath
	if path == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logger.Error("failed to get home directory", "error", err)
			os.Exit(1)
		}
		path = filepath.Join(homeDir, ".agent-memory", "tasks")
	}

	// Create repository
	repo, err := filesystem.NewRepository(path)
	if err != nil {
		logger.Error("failed to create repository", "path", path, "error", err)
		os.Exit(1)
	}
	defer repo.Close()

	logger.Info("using filesystem storage", "path", path)

	// Create services
	taskSvc := service.NewTaskService(repo, logger)
	workspaceSvc := service.NewWorkspaceService(repo, logger)

	// Create and run MCP server
	server := mcptransport.NewServer(taskSvc, workspaceSvc, logger)

	logger.Info("starting agent-memory MCP server", "version", "1.0.0")

	if err := server.ServeStdio(); err != nil {
		logger.Error("server error", "error", err)
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
