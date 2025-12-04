package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"agent-memory/internal/application/service"
	"agent-memory/internal/infrastructure/config"
	"agent-memory/internal/infrastructure/storage/filesystem"
	mcptransport "agent-memory/internal/transport/mcp"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "", "Path to config file (default: auto-detect)")
	tasksPath := flag.String("tasks-path", "", "Path to tasks directory (overrides config)")
	logLevel := flag.String("log-level", "", "Log level: debug, info, warn, error (overrides config)")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error

	if *configPath != "" {
		cfg, err = config.Load(*configPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
	} else {
		cfg, err = config.LoadFromDefaultLocations()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
			os.Exit(1)
		}
	}

	// Override config with command line flags
	if *tasksPath != "" {
		cfg.TasksPath = *tasksPath
	}
	if *logLevel != "" {
		cfg.LogLevel = *logLevel
	}

	// Setup logger
	var level slog.Level
	switch cfg.LogLevel {
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

	// Resolve tasks path
	path, err := cfg.ResolveTasksPath()
	if err != nil {
		logger.Error("failed to resolve tasks path", "error", err)
		os.Exit(1)
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

	logger.Info("starting agent-memory MCP server",
		"version", cfg.Server.Version,
		"name", cfg.Server.Name,
	)

	if err := server.ServeStdio(); err != nil {
		logger.Error("server error", "error", err)
		fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		os.Exit(1)
	}
}
