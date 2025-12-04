package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.TasksPath != "" {
		t.Errorf("DefaultConfig().TasksPath = %q, want empty", cfg.TasksPath)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("DefaultConfig().LogLevel = %q, want \"info\"", cfg.LogLevel)
	}
	if cfg.Server.Name != "agent-memory" {
		t.Errorf("DefaultConfig().Server.Name = %q, want \"agent-memory\"", cfg.Server.Name)
	}
	if cfg.Server.Version != "1.0.0" {
		t.Errorf("DefaultConfig().Server.Version = %q, want \"1.0.0\"", cfg.Server.Version)
	}
}

func TestLoad(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
tasks_path: /custom/tasks
log_level: debug
server:
  name: test-server
  version: "2.0.0"
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.TasksPath != "/custom/tasks" {
		t.Errorf("Config.TasksPath = %q, want \"/custom/tasks\"", cfg.TasksPath)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("Config.LogLevel = %q, want \"debug\"", cfg.LogLevel)
	}
	if cfg.Server.Name != "test-server" {
		t.Errorf("Config.Server.Name = %q, want \"test-server\"", cfg.Server.Name)
	}
	if cfg.Server.Version != "2.0.0" {
		t.Errorf("Config.Server.Version = %q, want \"2.0.0\"", cfg.Server.Version)
	}
}

func TestLoad_NotFound(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("Load() should not error for missing file, got: %v", err)
	}

	// Should return default config
	if cfg.LogLevel != "info" {
		t.Errorf("Config.LogLevel = %q, want \"info\" (default)", cfg.LogLevel)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	if err := os.WriteFile(configPath, []byte("invalid: yaml: content: ["), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("Load() should error for invalid YAML")
	}
}

func TestConfig_ResolveTasksPath(t *testing.T) {
	tests := []struct {
		name      string
		tasksPath string
		wantErr   bool
	}{
		{
			name:      "empty path uses default",
			tasksPath: "",
			wantErr:   false,
		},
		{
			name:      "absolute path",
			tasksPath: "/custom/tasks",
			wantErr:   false,
		},
		{
			name:      "home directory expansion",
			tasksPath: "~/tasks",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{TasksPath: tt.tasksPath}
			path, err := cfg.ResolveTasksPath()

			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveTasksPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && path == "" {
				t.Error("ResolveTasksPath() returned empty path")
			}

			// Check that ~ is expanded
			if tt.tasksPath == "~/tasks" && path[0] == '~' {
				t.Error("ResolveTasksPath() did not expand ~")
			}
		})
	}
}

func TestConfig_ResolveTasksPath_AbsolutePath(t *testing.T) {
	cfg := &Config{TasksPath: "/custom/tasks"}
	path, err := cfg.ResolveTasksPath()

	if err != nil {
		t.Fatalf("ResolveTasksPath() error = %v", err)
	}

	if path != "/custom/tasks" {
		t.Errorf("ResolveTasksPath() = %q, want \"/custom/tasks\"", path)
	}
}

func TestConfig_Save(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "subdir", "config.yaml")

	cfg := &Config{
		TasksPath: "/saved/tasks",
		LogLevel:  "warn",
		Server: ServerConfig{
			Name:    "saved-server",
			Version: "3.0.0",
		},
	}

	if err := cfg.Save(configPath); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Reload and verify
	loaded, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.TasksPath != cfg.TasksPath {
		t.Errorf("Loaded TasksPath = %q, want %q", loaded.TasksPath, cfg.TasksPath)
	}
	if loaded.LogLevel != cfg.LogLevel {
		t.Errorf("Loaded LogLevel = %q, want %q", loaded.LogLevel, cfg.LogLevel)
	}
	if loaded.Server.Name != cfg.Server.Name {
		t.Errorf("Loaded Server.Name = %q, want %q", loaded.Server.Name, cfg.Server.Name)
	}
}

func TestLoadFromDefaultLocations(t *testing.T) {
	// This test just verifies it doesn't crash and returns a valid config
	cfg, err := LoadFromDefaultLocations()
	if err != nil {
		t.Fatalf("LoadFromDefaultLocations() error = %v", err)
	}

	if cfg == nil {
		t.Error("LoadFromDefaultLocations() returned nil config")
	}
}
